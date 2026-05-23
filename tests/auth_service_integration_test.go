package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthHTTP_Integration(t *testing.T) {
	pgDSN, redisDSN, cleanup := SetupTestInfrastructure(t)
	defer cleanup()

	ctx := context.Background()
	api, appCleanup, err := newTestAPI(ctx, pgDSN, redisDSN)
	if err != nil {
		t.Fatalf("failed to setup app: %v", err)
	}
	defer appCleanup()

	Convey("Given a running API", t, func() {
		validEmail := fmt.Sprintf("auth-test-%s@example.com", uuid.New().String())
		validPassword := "securePassword123!"

		Convey("When registering a new user", func() {
			body, _ := json.Marshal(map[string]string{
				"email":    validEmail,
				"password": validPassword,
			})
			w := doRequest(api, http.MethodPost, "/api/v1/auth/register", "", body, "application/json")

			Convey("Then it succeeds and returns tokens", func() {
				So(w.Code, ShouldEqual, http.StatusOK)

				var resp struct {
					Data struct {
						AccessToken  string `json:"access_token"`
						RefreshToken string `json:"refresh_token"`
						ExpiresIn    int64  `json:"expires_in"`
					} `json:"data"`
				}
				err = json.Unmarshal(w.Body.Bytes(), &resp)
				So(err, ShouldBeNil)
				So(resp.Data.AccessToken, ShouldNotBeEmpty)
				So(resp.Data.RefreshToken, ShouldNotBeEmpty)
				So(resp.Data.ExpiresIn, ShouldBeGreaterThan, 0)
			})

			Convey("Then registering the exact same email again fails with 409", func() {
				wDuplicate := doRequest(api, http.MethodPost, "/api/v1/auth/register", "", body, "application/json")
				So(wDuplicate.Code, ShouldEqual, http.StatusConflict)
			})
		})

		Convey("When registering with invalid data", func() {
			Convey("Then short password returns 422", func() {
				body, _ := json.Marshal(map[string]string{
					"email":    fmt.Sprintf("test-%s@example.com", uuid.New().String()),
					"password": "short", // minLength is 8
				})
				w := doRequest(api, http.MethodPost, "/api/v1/auth/register", "", body, "application/json")
				So(w.Code, ShouldEqual, http.StatusUnprocessableEntity)
			})

			Convey("Then invalid email returns 422", func() {
				body, _ := json.Marshal(map[string]string{
					"email":    "not-an-email",
					"password": validPassword,
				})
				w := doRequest(api, http.MethodPost, "/api/v1/auth/register", "", body, "application/json")
				So(w.Code, ShouldEqual, http.StatusUnprocessableEntity)
			})
		})

		Convey("Given a registered user", func() {
			// Register first to ensure the user exists
			registerBody, _ := json.Marshal(map[string]string{
				"email":    validEmail,
				"password": validPassword,
			})
			doRequest(api, http.MethodPost, "/api/v1/auth/register", "", registerBody, "application/json")

			Convey("When logging in with correct credentials", func() {
				loginBody, _ := json.Marshal(map[string]string{
					"email":    validEmail,
					"password": validPassword,
				})
				w := doRequest(api, http.MethodPost, "/api/v1/auth/login", "", loginBody, "application/json")

				Convey("Then it succeeds and returns tokens", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
					var resp struct {
						Data struct {
							AccessToken  string `json:"access_token"`
							RefreshToken string `json:"refresh_token"`
						} `json:"data"`
					}
					err = json.Unmarshal(w.Body.Bytes(), &resp)
					So(err, ShouldBeNil)
					So(resp.Data.AccessToken, ShouldNotBeEmpty)

					Convey("And the refresh token can be used to get new tokens", func() {
						refreshBody, _ := json.Marshal(map[string]string{
							"refresh_token": resp.Data.RefreshToken,
						})
						wRefresh := doRequest(api, http.MethodPost, "/api/v1/auth/refresh", "", refreshBody, "application/json")
						So(wRefresh.Code, ShouldEqual, http.StatusOK)
					})
				})
			})

			Convey("When logging in with incorrect password", func() {
				loginBody, _ := json.Marshal(map[string]string{
					"email":    validEmail,
					"password": "wrongpassword!",
				})
				w := doRequest(api, http.MethodPost, "/api/v1/auth/login", "", loginBody, "application/json")

				Convey("Then it returns 401 Unauthorized", func() {
					So(w.Code, ShouldEqual, http.StatusUnauthorized)
				})
			})

			Convey("When deleting the account", func() {
				// Login to get token
				loginBody, _ := json.Marshal(map[string]string{
					"email":    validEmail,
					"password": validPassword,
				})
				wLogin := doRequest(api, http.MethodPost, "/api/v1/auth/login", "", loginBody, "application/json")
				
				So(wLogin.Code, ShouldEqual, http.StatusOK)

				var resp struct {
					Data struct {
						AccessToken string `json:"access_token"`
					} `json:"data"`
				}
				err := json.Unmarshal(wLogin.Body.Bytes(), &resp)
				So(err, ShouldBeNil)
				So(resp.Data.AccessToken, ShouldNotBeEmpty)

				token := resp.Data.AccessToken

				// Delete the account
				wDel := doRequest(api, http.MethodDelete, "/api/v1/auth/account", token, nil, "")
				
				Convey("Then it returns 204 No Content", func() {
					if wDel.Code != http.StatusNoContent {
						t.Logf("Delete Account failed: %s", wDel.Body.String())
					}
					So(wDel.Code, ShouldEqual, http.StatusNoContent)
				})

				Convey("And subsequent logins fail", func() {
					wLoginAfterDel := doRequest(api, http.MethodPost, "/api/v1/auth/login", "", loginBody, "application/json")
					So(wLoginAfterDel.Code, ShouldEqual, http.StatusUnauthorized)
				})
			})
		})

		Convey("When refreshing with an invalid token", func() {
			refreshBody, _ := json.Marshal(map[string]string{
				"refresh_token": "this-is-not-a-valid-token",
			})
			wRefresh := doRequest(api, http.MethodPost, "/api/v1/auth/refresh", "", refreshBody, "application/json")

			Convey("Then it returns 401 Unauthorized", func() {
				So(wRefresh.Code, ShouldEqual, http.StatusUnauthorized)
			})
		})

		Convey("When deleting account without a token", func() {
			wDel := doRequest(api, http.MethodDelete, "/api/v1/auth/account", "", nil, "")

			Convey("Then it returns 401 Unauthorized", func() {
				So(wDel.Code, ShouldEqual, http.StatusUnauthorized)
			})
		})
	})
}
