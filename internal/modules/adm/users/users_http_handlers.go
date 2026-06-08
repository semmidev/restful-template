package users

import (
	"context"
	"fmt"

	"github.com/danielgtaylor/huma/v2"
	"github.com/semmidev/restful-template/internal/shared/httpapi"
	"github.com/semmidev/restful-template/internal/shared/policy"
	"github.com/semmidev/restful-template/internal/shared/wideevent"
)

type userHandler struct {
	service UserService
}

func (h *userHandler) toUserResponse(u *User) userResponse {
	return userResponse{
		ID:         u.ID.String(),
		Email:      u.Email,
		ActiveRole: u.ActiveRole,
		Roles:      u.Roles,
	}
}

func (h *userHandler) handleList(ctx context.Context, in *listUsersReq) (*listUsersRes, error) {
	if err := policy.Authorize(ctx, "user:list", ""); err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}

	items, total, err := h.service.List(ctx, in.Page, in.PerPage, in.Keyword, in.SortBy, in.SortDir)
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}

	respItems := make([]userResponse, len(items))
	for i, item := range items {
		respItems[i] = h.toUserResponse(item)
	}

	resp := &listUsersRes{}
	resp.XTotalCount = total
	resp.Body.Data.Items = respItems
	resp.Body.Data.Total = total
	resp.Body.Data.Page = in.Page
	resp.Body.Data.PerPage = in.PerPage
	resp.Body.Data.Keyword = in.Keyword
	resp.Body.Data.SortBy = in.SortBy
	resp.Body.Data.SortDir = in.SortDir

	links := make(map[string]string)
	baseURL := "/api/v1/adm/users"
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
}

func (h *userHandler) handleGet(ctx context.Context, in *getUserReq) (*getUserRes, error) {
	if err := policy.Authorize(ctx, "user:read", ""); err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}

	u, err := h.service.GetByID(ctx, in.ID)
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}

	resp := &getUserRes{}
	resp.Body.Data = h.toUserResponse(u)
	return resp, nil
}

func (h *userHandler) handleCreate(ctx context.Context, in *createUserReq) (*createUserRes, error) {
	if err := policy.Authorize(ctx, "user:create", ""); err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}

	u, err := h.service.Create(ctx, CreateUserInput{
		Email:      in.Body.Email,
		Password:   in.Body.Password,
		ActiveRole: in.Body.ActiveRole,
		Roles:      in.Body.Roles,
	})
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}

	resp := &createUserRes{}
	resp.Body.Data = h.toUserResponse(u)
	return resp, nil
}

func (h *userHandler) handleUpdate(ctx context.Context, in *updateUserReq) (*updateUserRes, error) {
	if err := policy.Authorize(ctx, "user:update", ""); err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}

	u, err := h.service.Update(ctx, in.ID, UpdateUserInput{
		Email:      in.Body.Email,
		Password:   in.Body.Password,
		ActiveRole: in.Body.ActiveRole,
		Roles:      in.Body.Roles,
	})
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}

	resp := &updateUserRes{}
	resp.Body.Data = h.toUserResponse(u)
	return resp, nil
}

func (h *userHandler) handleDelete(ctx context.Context, in *deleteUserReq) (*deleteUserRes, error) {
	if err := policy.Authorize(ctx, "user:delete", ""); err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}

	currentUserID, err := httpapi.ExtractUserID(ctx)
	if err == nil && currentUserID == in.ID {
		return nil, huma.Error400BadRequest("Admins cannot delete their own accounts to prevent lockouts")
	}

	wideevent.Add(ctx, "deleted_user_id", in.ID.String())

	if err := h.service.Delete(ctx, in.ID); err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}

	return &deleteUserRes{}, nil
}
