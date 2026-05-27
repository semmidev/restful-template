package todos

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/conditional"
	"github.com/google/uuid"
	"github.com/semmidev/restful-template/internal/shared/httpapi"
	"github.com/semmidev/restful-template/internal/shared/wideevent"
)

type CreateTodoForm struct {
	Title       string        `form:"title" minLength:"1" maxLength:"200"`
	Description string        `form:"description" maxLength:"2000"`
	Cover       huma.FormFile `form:"cover" contentType:"image/*" doc:"Image file to upload as cover" required:"false"`
}

type UpdateTodoForm struct {
	Title       string        `form:"title" maxLength:"200" required:"false"`
	Description string        `form:"description" maxLength:"2000" required:"false"`
	Status      string        `form:"status" enum:"pending,in_progress,done" required:"false"`
	Cover       huma.FormFile `form:"cover" contentType:"image/*" doc:"Image file to upload as cover" required:"false"`
}

type TodoResp struct {
	ETag         string    `header:"ETag" doc:"Entity tag for optimistic locking"`
	LastModified time.Time `header:"Last-Modified" doc:"Last modification time"`
	Body         struct {
		Data *Todo `json:"data"`
	}
}

type ListData struct {
	Items   []*Todo           `json:"items"`
	Total   int               `json:"total"`
	Page    int               `json:"page"`
	PerPage int               `json:"per_page"`
	Links   map[string]string `json:"_links,omitempty"`
	Keyword string            `json:"keyword,omitempty" doc:"Active keyword filter (empty if none)"`
	SortBy  string            `json:"sort_by" doc:"Column sorted by"`
	SortDir string            `json:"sort_dir" doc:"Sort direction"`
}

type ListResp struct {
	XTotalCount int    `header:"X-Total-Count" doc:"Total number of items matching the query"`
	Link        string `header:"Link" doc:"RFC 8288 pagination links"`
	Body        struct {
		Data ListData `json:"data"`
	}
}

func RegisterTodoRoutes(api huma.API, todos TodoService) {
	huma.Register(api, huma.Operation{
		OperationID: "list-todos",
		Method:      http.MethodGet,
		Path:        "/api/v1/todos",
		Summary:     "List todos for the authenticated user",
		Tags:        []string{"Todos"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, func(ctx context.Context, in *struct {
		Page    int    `query:"page"      default:"1"  minimum:"1"`
		PerPage int    `query:"per_page"  default:"20" minimum:"1" maximum:"100"`
		Status  string `query:"status"  enum:"pending,in_progress,done," doc:"Filter by status (empty = all)"`
		Keyword string `query:"q"       maxLength:"100"              doc:"Case-insensitive substring search on title and description"`
		SortBy  string `query:"sort_by" default:"created_at" enum:"created_at,updated_at,title,status" doc:"Column to sort by"`
		SortDir string `query:"sort_dir" default:"desc" enum:"asc,desc" doc:"Sort direction"`
	}) (*ListResp, error) {
		userID, err := httpapi.ExtractUserID(ctx)
		if err != nil {
			return nil, httpapi.ToHumaErr(ctx, err)
		}
		offset := (in.Page - 1) * in.PerPage

		items, total, err := todos.List(ctx, ListTodosQuery{
			UserID: userID,
			Limit:  in.PerPage,
			Offset: offset,
			Status: func() *TodoStatus {
				if in.Status == "" {
					return nil
				}
				s := TodoStatus(in.Status)
				return &s
			}(),
			Keyword: in.Keyword,
			SortBy:  in.SortBy,
			SortDir: in.SortDir,
		})
		if err != nil {
			return nil, httpapi.ToHumaErr(ctx, err)
		}
		wideevent.Add(ctx, "todo_count", len(items))
		wideevent.Add(ctx, "todo_total", total)
		resp := &ListResp{}
		resp.XTotalCount = total
		resp.Body.Data.Items = items
		resp.Body.Data.Total = total
		resp.Body.Data.Page = in.Page
		resp.Body.Data.PerPage = in.PerPage
		resp.Body.Data.Keyword = in.Keyword
		resp.Body.Data.SortBy = in.SortBy
		resp.Body.Data.SortDir = in.SortDir

		links := make(map[string]string)
		baseURL := "/api/v1/todos"
		links["self"] = fmt.Sprintf("%s?page=%d&per_page=%d", baseURL, in.Page, in.PerPage)
		links["first"] = fmt.Sprintf("%s?page=%d&per_page=%d", baseURL, 1, in.PerPage)

		lastPage := (total + in.PerPage - 1) / in.PerPage
		if lastPage < 1 {
			lastPage = 1
		}
		links["last"] = fmt.Sprintf("%s?page=%d&per_page=%d", baseURL, lastPage, in.PerPage)

		if in.Page > 1 {
			links["prev"] = fmt.Sprintf("%s?page=%d&per_page=%d", baseURL, in.Page-1, in.PerPage)
		}
		if in.Page < lastPage {
			links["next"] = fmt.Sprintf("%s?page=%d&per_page=%d", baseURL, in.Page+1, in.PerPage)
		}
		resp.Body.Data.Links = links

		// Format RFC 8288 Link header
		linkHeader := ""
		if links["next"] != "" {
			linkHeader += fmt.Sprintf(`<%s>; rel="next", `, links["next"])
		}
		if links["prev"] != "" {
			linkHeader += fmt.Sprintf(`<%s>; rel="prev", `, links["prev"])
		}
		linkHeader += fmt.Sprintf(`<%s>; rel="first", <%s>; rel="last"`, links["first"], links["last"])
		resp.Link = linkHeader

		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "create-todo",
		Method:        http.MethodPost,
		Path:          "/api/v1/todos",
		Summary:       "Create a new todo",
		Tags:          []string{"Todos"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, in *struct {
		RawBody huma.MultipartFormFiles[CreateTodoForm]
	}) (*TodoResp, error) {
		userID, err := httpapi.ExtractUserID(ctx)
		if err != nil {
			return nil, httpapi.ToHumaErr(ctx, err)
		}

		data := in.RawBody.Data()
		var coverBase64 *string
		if data.Cover.IsSet {
			b, err := io.ReadAll(data.Cover)
			if err != nil {
				return nil, huma.Error400BadRequest("failed to read cover image")
			}
			encoded := fmt.Sprintf("data:%s;base64,%s", data.Cover.ContentType, base64.StdEncoding.EncodeToString(b))
			coverBase64 = &encoded
		}

		t, err := todos.Create(ctx, CreateTodoInput{
			UserID:      userID,
			Title:       data.Title,
			Description: data.Description,
			Cover:       coverBase64,
		})
		if err != nil {
			return nil, httpapi.ToHumaErr(ctx, err)
		}
		wideevent.Add(ctx, "todo_id", t.ID.String())
		wideevent.Add(ctx, "todo_title", t.Title)
		resp := &TodoResp{}
		resp.ETag = fmt.Sprintf(`"%s"`, t.UpdatedAt.Format(time.RFC3339Nano))
		resp.LastModified = t.UpdatedAt
		resp.Body.Data = t
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-todo",
		Method:      http.MethodGet,
		Path:        "/api/v1/todos/{id}",
		Summary:     "Get a single todo by ID",
		Tags:        []string{"Todos"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, func(ctx context.Context, in *struct {
		ID uuid.UUID `path:"id" doc:"Todo UUID"`
		conditional.Params
	}) (*TodoResp, error) {
		userID, err := httpapi.ExtractUserID(ctx)
		if err != nil {
			return nil, httpapi.ToHumaErr(ctx, err)
		}
		t, err := todos.Get(ctx, userID, in.ID)
		if err != nil {
			wideevent.Add(ctx, "todo_id", in.ID.String())
			return nil, httpapi.ToHumaErr(ctx, err)
		}
		wideevent.Add(ctx, "todo_id", t.ID.String())

		etag := fmt.Sprintf(`"%s"`, t.UpdatedAt.Format(time.RFC3339Nano))
		if err := in.PreconditionFailed(etag, t.UpdatedAt); err != nil {
			return nil, err
		}

		resp := &TodoResp{}
		resp.ETag = etag
		resp.LastModified = t.UpdatedAt
		resp.Body.Data = t
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "update-todo",
		Method:      http.MethodPatch,
		Path:        "/api/v1/todos/{id}",
		Summary:     "Update a todo",
		Tags:        []string{"Todos"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, func(ctx context.Context, in *struct {
		ID      uuid.UUID `path:"id"`
		RawBody huma.MultipartFormFiles[UpdateTodoForm]
		conditional.Params
	}) (*TodoResp, error) {
		userID, err := httpapi.ExtractUserID(ctx)
		if err != nil {
			return nil, httpapi.ToHumaErr(ctx, err)
		}

		// Optimistic locking check before update
		existing, err := todos.Get(ctx, userID, in.ID)
		if err != nil {
			wideevent.Add(ctx, "todo_id", in.ID.String())
			return nil, httpapi.ToHumaErr(ctx, err)
		}
		etag := fmt.Sprintf(`"%s"`, existing.UpdatedAt.Format(time.RFC3339Nano))
		if err := in.PreconditionFailed(etag, existing.UpdatedAt); err != nil {
			return nil, err // Returns 412 Precondition Failed if If-Match doesn't match
		}

		data := in.RawBody.Data()

		var title *string
		if _, ok := in.RawBody.Form.Value["title"]; ok {
			title = &data.Title
		}

		var desc *string
		if _, ok := in.RawBody.Form.Value["description"]; ok {
			desc = &data.Description
		}

		var status *TodoStatus
		if _, ok := in.RawBody.Form.Value["status"]; ok && data.Status != "" {
			st := TodoStatus(data.Status)
			status = &st
		}

		var coverBase64 *string
		if data.Cover.IsSet {
			b, err := io.ReadAll(data.Cover)
			if err != nil {
				return nil, huma.Error400BadRequest("failed to read cover image")
			}
			encoded := fmt.Sprintf("data:%s;base64,%s", data.Cover.ContentType, base64.StdEncoding.EncodeToString(b))
			coverBase64 = &encoded
		} else if _, ok := in.RawBody.Form.File["cover"]; ok || (len(in.RawBody.Form.Value["cover"]) > 0 && in.RawBody.Form.Value["cover"][0] == "") {
			// Field was sent but empty, meaning user wants to remove the cover
			empty := ""
			coverBase64 = &empty
		}

		t, err := todos.Update(ctx, UpdateTodoInput{
			ID:          in.ID,
			UserID:      userID,
			Title:       title,
			Description: desc,
			Cover:       coverBase64,
			Status:      status,
		})
		if err != nil {
			wideevent.Add(ctx, "todo_id", in.ID.String())
			return nil, httpapi.ToHumaErr(ctx, err)
		}
		wideevent.Add(ctx, "todo_id", t.ID.String())
		wideevent.Add(ctx, "todo_title", t.Title)
		wideevent.Add(ctx, "todo_status", string(t.Status))
		resp := &TodoResp{}
		resp.ETag = fmt.Sprintf(`"%s"`, t.UpdatedAt.Format(time.RFC3339Nano))
		resp.LastModified = t.UpdatedAt
		resp.Body.Data = t
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "delete-todo",
		Method:        http.MethodDelete,
		Path:          "/api/v1/todos/{id}",
		Summary:       "Delete a todo",
		Tags:          []string{"Todos"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, handleDeleteTodo(todos))
}

func handleDeleteTodo(todos TodoService) func(ctx context.Context, in *struct {
	ID uuid.UUID `path:"id"`
}) (*struct{}, error) {
	return func(ctx context.Context, in *struct {
		ID uuid.UUID `path:"id"`
	}) (*struct{}, error) {
		userID, err := httpapi.ExtractUserID(ctx)
		if err != nil {
			return nil, httpapi.ToHumaErr(ctx, err)
		}
		if err := todos.Delete(ctx, userID, in.ID); err != nil {
			wideevent.Add(ctx, "todo_id", in.ID.String())
			return nil, httpapi.ToHumaErr(ctx, err)
		}
		wideevent.Add(ctx, "todo_id", in.ID.String())
		return &struct{}{}, nil
	}
}
