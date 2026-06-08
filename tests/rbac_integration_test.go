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

func TestRBAC_Integration(t *testing.T) {
	pgDSN, redisDSN, cleanup := SetupTestInfrastructure(t)
	defer cleanup()

	ctx := context.Background()
	api, appCleanup, err := newTestAPI(ctx, pgDSN, redisDSN)
	if err != nil {
		t.Fatalf("failed to setup app: %v", err)
	}
	defer appCleanup()

	Convey("Given a running API with RBAC enabled", t, func() {
		Convey("When registering a normal user", func() {
			email := fmt.Sprintf("normal-%s@example.com", uuid.New().String())
			body, _ := json.Marshal(map[string]string{
				"email":    email,
				"password": "password123!",
			})
			w := doRequest(api, http.MethodPost, "/api/v1/auth/register", "", body, "application/json")
			So(w.Code, ShouldEqual, http.StatusOK)

			var resp struct {
				User struct {
					ActiveRole string   `json:"active_role"`
					Roles      []string `json:"roles"`
				} `json:"user"`
			}
			err = json.Unmarshal(w.Body.Bytes(), &resp)
			So(err, ShouldBeNil)

			Convey("Then the user has default role 'user' and only 'user' role", func() {
				So(resp.User.ActiveRole, ShouldEqual, "user")
				So(resp.User.Roles, ShouldResemble, []string{"user"})
			})

			Convey("And trying to switch to 'admin' role fails with 403 Forbidden", func() {
				// Login to get access token
				loginBody, _ := json.Marshal(map[string]string{
					"email":    email,
					"password": "password123!",
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

				switchBody, _ := json.Marshal(map[string]string{
					"role": "admin",
				})
				wSwitch := doRequest(api, http.MethodPost, "/api/v1/auth/switch-role", token, switchBody, "application/json")
				So(wSwitch.Code, ShouldEqual, http.StatusForbidden)
			})
		})

		Convey("When registering an admin user", func() {
			email := fmt.Sprintf("admin-user-%s@example.com", uuid.New().String())
			body, _ := json.Marshal(map[string]string{
				"email":    email,
				"password": "password123!",
			})
			w := doRequest(api, http.MethodPost, "/api/v1/auth/register", "", body, "application/json")
			So(w.Code, ShouldEqual, http.StatusOK)

			var resp struct {
				User struct {
					ActiveRole string   `json:"active_role"`
					Roles      []string `json:"roles"`
				} `json:"user"`
			}
			err = json.Unmarshal(w.Body.Bytes(), &resp)
			So(err, ShouldBeNil)

			Convey("Then the user has active role 'user' but has both 'user' and 'admin' roles", func() {
				So(resp.User.ActiveRole, ShouldEqual, "user")
				So(resp.User.Roles, ShouldContain, "user")
				So(resp.User.Roles, ShouldContain, "admin")
			})

			Convey("And the user can switch their active role to 'admin'", func() {
				loginBody, _ := json.Marshal(map[string]string{
					"email":    email,
					"password": "password123!",
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

				switchBody, _ := json.Marshal(map[string]string{
					"role": "admin",
				})
				wSwitch := doRequest(api, http.MethodPost, "/api/v1/auth/switch-role", token, switchBody, "application/json")
				So(wSwitch.Code, ShouldEqual, http.StatusOK)

				var respSwitch struct {
					User struct {
						ActiveRole string `json:"active_role"`
					} `json:"user"`
				}
				err = json.Unmarshal(wSwitch.Body.Bytes(), &respSwitch)
				So(err, ShouldBeNil)
				So(respSwitch.User.ActiveRole, ShouldEqual, "admin")
			})
		})
	})
}
