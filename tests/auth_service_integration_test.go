package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

			Convey("Then it succeeds and returns tokens in cookies", func() {
				So(w.Code, ShouldEqual, http.StatusOK)

				respHttp := &http.Response{Header: w.Header()}
				cookies := respHttp.Cookies()
				var hasAccess, hasRefresh bool
				for _, cookie := range cookies {
					if cookie.Name == "access_token" {
						hasAccess = true
					} else if cookie.Name == "refresh_token" {
						hasRefresh = true
					}
				}
				So(hasAccess, ShouldBeTrue)
				So(hasRefresh, ShouldBeTrue)
			})

			Convey("Then registering the exact same email again fails with 409", func() {
				wDuplicate := doRequest(api, http.MethodPost, "/api/v1/auth/register", "", body, "application/json")
				So(wDuplicate.Code, ShouldEqual, http.StatusConflict)
			})
		})

		Convey("When retrieving Google configuration", func() {
			w := doRequest(api, http.MethodGet, "/api/v1/auth/google/config", "", nil, "")

			Convey("Then it succeeds and returns configuration values", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
				var resp struct {
					ClientID    string `json:"client_id"`
					RedirectURI string `json:"redirect_uri"`
				}
				err = json.Unmarshal(w.Body.Bytes(), &resp)
				So(err, ShouldBeNil)
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

				Convey("Then it succeeds and returns tokens in cookies", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
					respHttp := &http.Response{Header: w.Header()}
					cookies := respHttp.Cookies()
					var accessTokenVal, refreshTokenVal string
					for _, cookie := range cookies {
						if cookie.Name == "access_token" {
							accessTokenVal = cookie.Value
						} else if cookie.Name == "refresh_token" {
							refreshTokenVal = cookie.Value
						}
					}
					So(accessTokenVal, ShouldNotBeEmpty)
					So(refreshTokenVal, ShouldNotBeEmpty)

					Convey("And the refresh token can be used to get new tokens", func() {
						reqRefresh := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
						reqRefresh.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshTokenVal})
						wRefresh := httptest.NewRecorder()
						api.ServeHTTP(wRefresh, reqRefresh)
						So(wRefresh.Code, ShouldEqual, http.StatusOK)

						respRefresh := &http.Response{Header: wRefresh.Header()}
						cookiesRefresh := respRefresh.Cookies()
						var hasNewAccess, hasNewRefresh bool
						for _, c := range cookiesRefresh {
							if c.Name == "access_token" {
								hasNewAccess = true
							} else if c.Name == "refresh_token" {
								hasNewRefresh = true
							}
						}
						So(hasNewAccess, ShouldBeTrue)
						So(hasNewRefresh, ShouldBeTrue)
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

				respHttp := &http.Response{Header: wLogin.Header()}
				cookies := respHttp.Cookies()
				var token string
				for _, cookie := range cookies {
					if cookie.Name == "access_token" {
						token = cookie.Value
					}
				}
				So(token, ShouldNotBeEmpty)

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
			reqRefresh := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
			reqRefresh.AddCookie(&http.Cookie{Name: "refresh_token", Value: "this-is-not-a-valid-token"})
			wRefresh := httptest.NewRecorder()
			api.ServeHTTP(wRefresh, reqRefresh)

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
