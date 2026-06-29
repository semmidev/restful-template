package auth

import (
	"context"
	"net/http"

	apperrors "github.com/semmidev/restful-template/internal/shared/errors"
	"github.com/semmidev/restful-template/internal/shared/httpapi"
	"github.com/semmidev/restful-template/internal/shared/wideevent"
)

func (h *authHandler) makeCookie(name, value, path string, maxAge int) string {
	cookie := &http.Cookie{ //nolint:gosec // Secure is intentionally set based on runtime env config; HttpOnly and SameSite are always present
		Name:     name,
		Value:    value,
		Path:     path,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   h.cfg.App.Env != "development",
		SameSite: http.SameSiteLaxMode,
	}
	return cookie.String()
}

func (h *authHandler) parseCookie(cookieHeader, name string) string {
	req := &http.Request{Header: http.Header{"Cookie": {cookieHeader}}}
	cookie, err := req.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (h *authHandler) handleRegister(ctx context.Context, in *authRegisterReq) (*authRegisterRes, error) {
	wideevent.Add(ctx, "user_email", in.Body.Email)
	pair, err := h.auth.Register(ctx, RegisterInput{Email: in.Body.Email, Password: in.Body.Password})
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}
	res := &authRegisterRes{
		SetCookie: []string{
			h.makeCookie("access_token", pair.AccessToken, "/", int(h.cfg.JWT.AccessTTL.Seconds())),
			h.makeCookie("refresh_token", pair.RefreshToken, "/api/v1/auth", int(h.cfg.JWT.RefreshTTL.Seconds())),
		},
	}
	res.Body.User.ID = pair.UserID.String()
	res.Body.User.Email = pair.UserEmail
	res.Body.User.ActiveRole = pair.ActiveRole
	res.Body.User.Roles = pair.Roles
	res.Body.User.Permissions = pair.Permissions
	return res, nil
}

func (h *authHandler) handleLogin(ctx context.Context, in *authLoginReq) (*authLoginRes, error) {
	wideevent.Add(ctx, "user_email", in.Body.Email)
	pair, err := h.auth.Login(ctx, LoginInput{Email: in.Body.Email, Password: in.Body.Password})
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}
	res := &authLoginRes{
		SetCookie: []string{
			h.makeCookie("access_token", pair.AccessToken, "/", int(h.cfg.JWT.AccessTTL.Seconds())),
			h.makeCookie("refresh_token", pair.RefreshToken, "/api/v1/auth", int(h.cfg.JWT.RefreshTTL.Seconds())),
		},
	}
	res.Body.User.ID = pair.UserID.String()
	res.Body.User.Email = pair.UserEmail
	res.Body.User.ActiveRole = pair.ActiveRole
	res.Body.User.Roles = pair.Roles
	res.Body.User.Permissions = pair.Permissions
	return res, nil
}

func (h *authHandler) handleRefresh(ctx context.Context, in *authRefreshReq) (*authRefreshRes, error) {
	refreshToken := h.parseCookie(in.Cookie, "refresh_token")
	if refreshToken == "" {
		return nil, httpapi.ToHumaErr(ctx, apperrors.NewUnauthorized("Missing refresh token", apperrors.ErrUnauthorized))
	}

	pair, err := h.auth.Refresh(ctx, refreshToken)
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}
	res := &authRefreshRes{
		SetCookie: []string{
			h.makeCookie("access_token", pair.AccessToken, "/", int(h.cfg.JWT.AccessTTL.Seconds())),
			h.makeCookie("refresh_token", pair.RefreshToken, "/api/v1/auth", int(h.cfg.JWT.RefreshTTL.Seconds())),
		},
	}
	res.Body.User.ID = pair.UserID.String()
	res.Body.User.Email = pair.UserEmail
	res.Body.User.ActiveRole = pair.ActiveRole
	res.Body.User.Roles = pair.Roles
	res.Body.User.Permissions = pair.Permissions
	return res, nil
}

func (h *authHandler) handleLogout(ctx context.Context, in *authLogoutReq) (*authLogoutRes, error) {
	refreshToken := h.parseCookie(in.Cookie, "refresh_token")

	// Execute logout to invalidate session in DB/cache.
	// Best-effort: we ignore error or log it, but always clear cookies.
	if refreshToken != "" {
		_ = h.auth.Logout(ctx, refreshToken)
	}

	res := &authLogoutRes{
		SetCookie: []string{
			h.makeCookie("access_token", "", "/", -1),
			h.makeCookie("refresh_token", "", "/api/v1/auth", -1),
		},
	}
	return res, nil
}

func (h *authHandler) handleDeleteAccount(ctx context.Context, in *authDeleteAccountReq) (*authDeleteAccountRes, error) {
	userID, err := httpapi.ExtractUserID(ctx)
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}
	wideevent.Add(ctx, "user_id", userID.String())

	if err := h.auth.DeleteAccount(ctx, userID); err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}
	return &authDeleteAccountRes{}, nil
}

func (h *authHandler) handleSwitchRole(ctx context.Context, in *authSwitchRoleReq) (*authSwitchRoleRes, error) {
	userID, err := httpapi.ExtractUserID(ctx)
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}
	wideevent.Add(ctx, "user_id", userID.String())
	wideevent.Add(ctx, "switch_role", in.Body.Role)

	pair, err := h.auth.SwitchRole(ctx, userID, in.Body.Role)
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}

	res := &authSwitchRoleRes{
		SetCookie: []string{
			h.makeCookie("access_token", pair.AccessToken, "/", int(h.cfg.JWT.AccessTTL.Seconds())),
			h.makeCookie("refresh_token", pair.RefreshToken, "/api/v1/auth", int(h.cfg.JWT.RefreshTTL.Seconds())),
		},
	}
	res.Body.User.ID = pair.UserID.String()
	res.Body.User.Email = pair.UserEmail
	res.Body.User.ActiveRole = pair.ActiveRole
	res.Body.User.Roles = pair.Roles
	res.Body.User.Permissions = pair.Permissions
	return res, nil
}

func (h *authHandler) handleGoogleLogin(ctx context.Context, in *authGoogleLoginReq) (*authGoogleLoginRes, error) {
	pair, err := h.auth.GoogleLogin(ctx, in.Body.Code, in.Body.CodeVerifier)
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}
	res := &authGoogleLoginRes{
		SetCookie: []string{
			h.makeCookie("access_token", pair.AccessToken, "/", int(h.cfg.JWT.AccessTTL.Seconds())),
			h.makeCookie("refresh_token", pair.RefreshToken, "/api/v1/auth", int(h.cfg.JWT.RefreshTTL.Seconds())),
		},
	}
	res.Body.User.ID = pair.UserID.String()
	res.Body.User.Email = pair.UserEmail
	res.Body.User.ActiveRole = pair.ActiveRole
	res.Body.User.Roles = pair.Roles
	res.Body.User.Permissions = pair.Permissions
	return res, nil
}

func (h *authHandler) handleGoogleConfig(ctx context.Context, in *authGoogleConfigReq) (*authGoogleConfigRes, error) {
	clientID, redirectURI := h.auth.GoogleConfig()
	res := &authGoogleConfigRes{}
	res.Body.ClientID = clientID
	res.Body.RedirectURI = redirectURI
	return res, nil
}
