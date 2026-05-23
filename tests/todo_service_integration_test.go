package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTodoHTTP_Integration(t *testing.T) {
	pgDSN, redisDSN, cleanup := SetupTestInfrastructure(t)
	defer cleanup()

	ctx := context.Background()
	api, appCleanup, err := newTestAPI(ctx, pgDSN, redisDSN)
	if err != nil {
		t.Fatalf("failed to setup app: %v", err)
	}
	defer appCleanup()

	Convey("Given a connected API with an authenticated user", t, func() {
		email := fmt.Sprintf("test-%s@example.com", uuid.New().String())
		token, err := registerAndLogin(api, email, "password123")
		So(err, ShouldBeNil)
		So(token, ShouldNotBeEmpty)

		Convey("When creating a new todo", func() {
			body, ct := buildMultipartBody(map[string]string{"title": "Integration HTTP Todo"})
			w := doRequest(api, http.MethodPost, "/api/v1/todos", token, body, ct)

			So(w.Code, ShouldEqual, http.StatusCreated)

			var resp struct {
				Data struct {
					ID    string `json:"id"`
					Title string `json:"title"`
				} `json:"data"`
			}
			err = json.Unmarshal(w.Body.Bytes(), &resp)
			So(err, ShouldBeNil)
			So(resp.Data.Title, ShouldEqual, "Integration HTTP Todo")
			So(resp.Data.ID, ShouldNotBeEmpty)

			createdID := resp.Data.ID

			Convey("Then getting the list returns at least 1 item", func() {
				wList := doRequest(api, http.MethodGet, "/api/v1/todos", token, nil, "")
				So(wList.Code, ShouldEqual, http.StatusOK)

				var respList struct {
					Data struct {
						Items []struct {
							ID string `json:"id"`
						} `json:"items"`
						Total int `json:"total"`
					} `json:"data"`
				}
				err = json.Unmarshal(wList.Body.Bytes(), &respList)
				So(err, ShouldBeNil)
				So(respList.Data.Total, ShouldBeGreaterThanOrEqualTo, 1)
			})

			Convey("Then getting the single todo by ID works", func() {
				wGet := doRequest(api, http.MethodGet, "/api/v1/todos/"+createdID, token, nil, "")
				So(wGet.Code, ShouldEqual, http.StatusOK)

				var respGet struct {
					Data struct {
						ID    string `json:"id"`
						Title string `json:"title"`
					} `json:"data"`
				}
				err = json.Unmarshal(wGet.Body.Bytes(), &respGet)
				So(err, ShouldBeNil)
				So(respGet.Data.ID, ShouldEqual, createdID)
				So(respGet.Data.Title, ShouldEqual, "Integration HTTP Todo")
			})

			Convey("Then updating the title works", func() {
				bodyUpd, ctUpd := buildMultipartBody(map[string]string{"title": "Updated Title"})
				wPatch := doRequest(api, http.MethodPatch, "/api/v1/todos/"+createdID, token, bodyUpd, ctUpd)
				So(wPatch.Code, ShouldEqual, http.StatusOK)

				var respPatch struct {
					Data struct {
						Title string `json:"title"`
					} `json:"data"`
				}
				err = json.Unmarshal(wPatch.Body.Bytes(), &respPatch)
				So(err, ShouldBeNil)
				So(respPatch.Data.Title, ShouldEqual, "Updated Title")
			})

			Convey("Then deleting the todo returns 204", func() {
				wDel := doRequest(api, http.MethodDelete, "/api/v1/todos/"+createdID, token, nil, "")
				So(wDel.Code, ShouldEqual, http.StatusNoContent)

				Convey("And getting it again returns 404", func() {
					wGetDel := doRequest(api, http.MethodGet, "/api/v1/todos/"+createdID, token, nil, "")
					So(wGetDel.Code, ShouldEqual, http.StatusNotFound)
				})
			})

			Convey("Then another user cannot access this user's todo", func() {
				email2 := fmt.Sprintf("other-%s@example.com", uuid.New().String())
				token2, err2 := registerAndLogin(api, email2, "password123")
				So(err2, ShouldBeNil)

				wGetOther := doRequest(api, http.MethodGet, "/api/v1/todos/"+createdID, token2, nil, "")
				So(wGetOther.Code, ShouldEqual, http.StatusNotFound)
			})
		})

		Convey("When creating a todo with a missing title", func() {
			body, ct := buildMultipartBody(map[string]string{"title": ""})
			w := doRequest(api, http.MethodPost, "/api/v1/todos", token, body, ct)

			Convey("Then it returns 422 Unprocessable Entity", func() {
				So(w.Code, ShouldEqual, http.StatusUnprocessableEntity)
			})
		})
	})

	Convey("Given an unauthenticated request", t, func() {
		Convey("When trying to get todos", func() {
			w := doRequest(api, http.MethodGet, "/api/v1/todos", "", nil, "")

			Convey("Then it returns 401 Unauthorized", func() {
				So(w.Code, ShouldEqual, http.StatusUnauthorized)
			})
		})
	})
}

var _ = humatest.New
