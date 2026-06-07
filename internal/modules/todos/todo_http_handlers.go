package todos

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/semmidev/restful-template/internal/shared/httpapi"
	"github.com/semmidev/restful-template/internal/shared/wideevent"
)

func (h *todoHandler) handleList(ctx context.Context, in *listTodosReq) (*listTodosRes, error) {
	userID, err := httpapi.ExtractUserID(ctx)
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}
	offset := (in.Page - 1) * in.PerPage

	items, total, err := h.todos.List(ctx, ListTodosQuery{
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

	resp := &listTodosRes{}
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
}

func (h *todoHandler) handleCreate(ctx context.Context, in *createTodoReq) (*createTodoRes, error) {
	data := in.RawBody.Data()
	wideevent.Add(ctx, "todo_title", data.Title)

	userID, err := httpapi.ExtractUserID(ctx)
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}

	coverBase64, err := processCoverImage(data.Cover)
	if err != nil {
		return nil, err
	}

	t, err := h.todos.Create(ctx, CreateTodoInput{
		UserID:      userID,
		Title:       data.Title,
		Description: data.Description,
		Cover:       coverBase64,
	})
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}
	wideevent.Add(ctx, "todo_id", t.ID.String())
	resp := &createTodoRes{}
	resp.ETag = fmt.Sprintf(`"%s"`, t.UpdatedAt.Format(time.RFC3339Nano))
	resp.LastModified = t.UpdatedAt
	resp.Body.Data = t
	return resp, nil
}

func (h *todoHandler) handleGet(ctx context.Context, in *getTodoReq) (*getTodoRes, error) {
	wideevent.Add(ctx, "todo_id", in.ID.String())
	userID, err := httpapi.ExtractUserID(ctx)
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}
	t, err := h.todos.Get(ctx, userID, in.ID)
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}

	rawEtag := t.UpdatedAt.Format(time.RFC3339Nano)
	if err := in.PreconditionFailed(rawEtag, t.UpdatedAt); err != nil {
		return nil, err
	}

	resp := &getTodoRes{}
	resp.ETag = fmt.Sprintf(`"%s"`, rawEtag)
	resp.LastModified = t.UpdatedAt
	resp.Body.Data = t
	return resp, nil
}

// handleUpdate applies a partial update to a todo.
//
// The entity is fetched once here for ETag validation and then passed
// directly into the service, which no longer does an internal re-fetch.
// Total DB calls: GET (1) + UPDATE (1) = 2, down from 3.
func (h *todoHandler) handleUpdate(ctx context.Context, in *updateTodoReq) (*updateTodoRes, error) {
	wideevent.Add(ctx, "todo_id", in.ID.String())
	userID, err := httpapi.ExtractUserID(ctx)
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}

	// Optimistic locking check before update (1st and only GET)
	existing, err := h.todos.Get(ctx, userID, in.ID)
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}
	rawEtag := existing.UpdatedAt.Format(time.RFC3339Nano)
	if err := in.PreconditionFailed(rawEtag, existing.UpdatedAt); err != nil {
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

	coverBase64, err := processCoverImage(data.Cover)
	if err != nil {
		return nil, err
	}
	// If no new cover was uploaded but the cover field was sent as empty, clear it
	if coverBase64 == nil {
		if _, ok := in.RawBody.Form.File["cover"]; ok || (len(in.RawBody.Form.Value["cover"]) > 0 && in.RawBody.Form.Value["cover"][0] == "") {
			empty := ""
			coverBase64 = &empty
		}
	}

	var updateMask []string
	if in.UpdateMask != "" {
		updateMask = strings.Split(in.UpdateMask, ",")
		for i, v := range updateMask {
			updateMask[i] = strings.TrimSpace(v)
		}
	}

	// Pass pre-fetched entity — service skips the re-fetch
	t, err := h.todos.Update(ctx, existing, UpdateTodoInput{
		ID:          in.ID,
		UserID:      userID,
		Title:       title,
		Description: desc,
		Cover:       coverBase64,
		Status:      status,
		UpdateMask:  updateMask,
	})
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}
	wideevent.Add(ctx, "todo_title", t.Title)
	wideevent.Add(ctx, "todo_status", string(t.Status))
	resp := &updateTodoRes{}
	resp.ETag = fmt.Sprintf(`"%s"`, t.UpdatedAt.Format(time.RFC3339Nano))
	resp.LastModified = t.UpdatedAt
	resp.Body.Data = t
	return resp, nil
}

func (h *todoHandler) handleDelete(ctx context.Context, in *deleteTodoReq) (*deleteTodoRes, error) {
	wideevent.Add(ctx, "todo_id", in.ID.String())
	userID, err := httpapi.ExtractUserID(ctx)
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}
	if err := h.todos.Delete(ctx, userID, in.ID); err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}
	return &deleteTodoRes{}, nil
}

// processCoverImage reads, size-checks, and MIME-validates an uploaded cover file.
// Returns nil if no file was set, a *string base64 data-URI if valid, or an error.
//
// LimitReader enforces maxCoverSize (5 MB), preventing memory exhaustion.
// http.DetectContentType sniffs the actual bytes to verify it's really an image,
// regardless of the client-supplied Content-Type header.
func processCoverImage(f huma.FormFile) (*string, error) {
	if !f.IsSet {
		return nil, nil
	}

	// Limit read to maxCoverSize+1 to distinguish "exactly limit" from "over limit"
	limited := io.LimitReader(f, maxCoverSize+1)
	b, err := io.ReadAll(limited)
	if err != nil {
		return nil, huma.Error400BadRequest("failed to read cover image")
	}
	if len(b) > maxCoverSize {
		return nil, &httpapi.APIError{
			Status: http.StatusRequestEntityTooLarge,
			Title:  "Request Entity Too Large",
			Code:   "COVER_TOO_LARGE",
			Detail: fmt.Sprintf("cover image must not exceed %d MB", maxCoverSize>>20),
		}
	}

	// Sniff the actual content type — do not trust the client's Content-Type
	contentType := http.DetectContentType(b)
	if !strings.HasPrefix(contentType, "image/") {
		return nil, &httpapi.APIError{
			Status: http.StatusBadRequest,
			Title:  "Bad Request",
			Code:   "INVALID_COVER_TYPE",
			Detail: fmt.Sprintf("cover must be an image file, got %q", contentType),
		}
	}

	encoded := fmt.Sprintf("data:%s;base64,%s", contentType, base64.StdEncoding.EncodeToString(b))
	return &encoded, nil
}
