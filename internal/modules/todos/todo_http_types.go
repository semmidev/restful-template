package todos

import (
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/conditional"
	"github.com/google/uuid"
)

// Shared single-todo body
type todoBody struct {
	Data *Todo `json:"data"`
}

// List

type listTodosReq struct {
	Page    int    `query:"page"      default:"1"  minimum:"1"`
	PerPage int    `query:"per_page"  default:"20" minimum:"1" maximum:"100"`
	Status  string `query:"status"  enum:"pending,in_progress,done," doc:"Filter by status (empty = all)"`
	Keyword string `query:"q"       maxLength:"100"              doc:"Case-insensitive substring search on title and description"`
	SortBy  string `query:"sort_by" default:"created_at" enum:"created_at,updated_at,title,status" doc:"Column to sort by"`
	SortDir string `query:"sort_dir" default:"desc" enum:"asc,desc" doc:"Sort direction"`
}

type listTodosData struct {
	Items   []*Todo           `json:"items"`
	Total   int               `json:"total"`
	Page    int               `json:"page"`
	PerPage int               `json:"per_page"`
	Links   map[string]string `json:"_links,omitempty"`
	Keyword string            `json:"keyword,omitempty" doc:"Active keyword filter (empty if none)"`
	SortBy  string            `json:"sort_by" doc:"Column sorted by"`
	SortDir string            `json:"sort_dir" doc:"Sort direction"`
}

type listTodosBody struct {
	Data listTodosData `json:"data"`
}

type listTodosRes struct {
	XTotalCount int    `header:"X-Total-Count" doc:"Total number of items matching the query"`
	Link        string `header:"Link" doc:"RFC 8288 pagination links"`
	Body        listTodosBody
}

// Create

type CreateTodoForm struct {
	Title       string        `form:"title" minLength:"1" maxLength:"200"`
	Description string        `form:"description" maxLength:"2000" required:"false"`
	Cover       huma.FormFile `form:"cover" contentType:"image/*" doc:"Image file to upload as cover (max 5 MB)" required:"false"`
}

type createTodoReq struct {
	RawBody huma.MultipartFormFiles[CreateTodoForm]
}

type createTodoRes struct {
	ETag         string    `header:"ETag" doc:"Entity tag for optimistic locking"`
	LastModified time.Time `header:"Last-Modified" doc:"Last modification time"`
	Body         todoBody
}

// Get

type getTodoReq struct {
	ID uuid.UUID `path:"id" doc:"Todo UUID"`
	conditional.Params
}

type getTodoRes struct {
	ETag         string    `header:"ETag" doc:"Entity tag for optimistic locking"`
	LastModified time.Time `header:"Last-Modified" doc:"Last modification time"`
	Body         todoBody
}

// Update

type UpdateTodoForm struct {
	Title       string        `form:"title" maxLength:"200" required:"false"`
	Description string        `form:"description" maxLength:"2000" required:"false"`
	Status      string        `form:"status" enum:"pending,in_progress,done" required:"false"`
	Cover       huma.FormFile `form:"cover" contentType:"image/*" doc:"Image file to upload as cover (max 5 MB)" required:"false"`
}

type updateTodoReq struct {
	ID         uuid.UUID `path:"id"`
	UpdateMask string    `query:"update_mask" doc:"Comma-separated list of fields to update (AIP-134)"`
	RawBody    huma.MultipartFormFiles[UpdateTodoForm]
	conditional.Params
}

type updateTodoRes struct {
	ETag         string    `header:"ETag" doc:"Entity tag for optimistic locking"`
	LastModified time.Time `header:"Last-Modified" doc:"Last modification time"`
	Body         todoBody
}

// Delete

type deleteTodoReq struct {
	ID uuid.UUID `path:"id"`
}

type deleteTodoRes struct{}
