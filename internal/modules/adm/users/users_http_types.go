package users

import (
	"github.com/google/uuid"
)

type userResponse struct {
	ID         string   `json:"id"`
	Email      string   `json:"email"`
	ActiveRole string   `json:"active_role"`
	Roles      []string `json:"roles"`
}

type singleUserBody struct {
	Data userResponse `json:"data"`
}

type listUsersReq struct {
	Page    int    `query:"page" default:"1" minimum:"1"`
	PerPage int    `query:"per_page" default:"20" minimum:"1" maximum:"100"`
	Keyword string `query:"q" maxLength:"100" doc:"Substring search on user email"`
	SortBy  string `query:"sort_by" default:"created_at" enum:"created_at,updated_at,email,active_role" doc:"Column to sort by"`
	SortDir string `query:"sort_dir" default:"desc" enum:"asc,desc" doc:"Sort direction"`
}

type listUsersData struct {
	Items   []userResponse    `json:"items"`
	Total   int               `json:"total"`
	Page    int               `json:"page"`
	PerPage int               `json:"per_page"`
	Links   map[string]string `json:"_links,omitempty"`
	Keyword string            `json:"keyword,omitempty"`
	SortBy  string            `json:"sort_by"`
	SortDir string            `json:"sort_dir"`
}

type listUsersBody struct {
	Data listUsersData `json:"data"`
}

type listUsersRes struct {
	XTotalCount int    `header:"X-Total-Count" doc:"Total number of users matching the query"`
	Link        string `header:"Link" doc:"RFC 8288 pagination links"`
	Body        listUsersBody
}

type createUserReq struct {
	Body struct {
		Email      string   `json:"email" format:"email" minLength:"3" maxLength:"254" doc:"User email address"`
		Password   string   `json:"password" minLength:"8" maxLength:"72" doc:"User password (8-72 chars)"`
		ActiveRole string   `json:"active_role" minLength:"1" doc:"Active default role"`
		Roles      []string `json:"roles" doc:"User roles array"`
	}
}

type createUserRes struct {
	Body singleUserBody
}

type getUserReq struct {
	ID uuid.UUID `path:"id" doc:"User UUID"`
}

type getUserRes struct {
	Body singleUserBody
}

type updateUserReq struct {
	ID   uuid.UUID `path:"id"`
	Body struct {
		Email      *string  `json:"email,omitempty" format:"email"`
		Password   *string  `json:"password,omitempty" minLength:"8" maxLength:"72"`
		ActiveRole *string  `json:"active_role,omitempty"`
		Roles      []string `json:"roles,omitempty"`
	}
}

type updateUserRes struct {
	Body singleUserBody
}

type deleteUserReq struct {
	ID uuid.UUID `path:"id"`
}

type deleteUserRes struct{}
