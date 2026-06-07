package auth

import (
	"context"

	"github.com/semmidev/restful-template/internal/shared/httpapi"
	"github.com/semmidev/restful-template/internal/shared/wideevent"
)

func (h *authHandler) handleRegister(ctx context.Context, in *authRegisterReq) (*authRegisterRes, error) {
	wideevent.Add(ctx, "user_email", in.Body.Email)
	pair, err := h.auth.Register(ctx, RegisterInput{Email: in.Body.Email, Password: in.Body.Password})
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}
	res := &authRegisterRes{}
	res.Body.Data.AccessToken = pair.AccessToken
	res.Body.Data.RefreshToken = pair.RefreshToken
	res.Body.Data.ExpiresIn = pair.ExpiresIn
	return res, nil
}

func (h *authHandler) handleLogin(ctx context.Context, in *authLoginReq) (*authLoginRes, error) {
	wideevent.Add(ctx, "user_email", in.Body.Email)
	pair, err := h.auth.Login(ctx, LoginInput{Email: in.Body.Email, Password: in.Body.Password})
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}
	res := &authLoginRes{}
	res.Body.Data.AccessToken = pair.AccessToken
	res.Body.Data.RefreshToken = pair.RefreshToken
	res.Body.Data.ExpiresIn = pair.ExpiresIn
	return res, nil
}

func (h *authHandler) handleRefresh(ctx context.Context, in *authRefreshReq) (*authRefreshRes, error) {
	pair, err := h.auth.Refresh(ctx, in.Body.RefreshToken)
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}
	res := &authRefreshRes{}
	res.Body.Data.AccessToken = pair.AccessToken
	res.Body.Data.RefreshToken = pair.RefreshToken
	res.Body.Data.ExpiresIn = pair.ExpiresIn
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

func (h *authHandler) handleGoogleLogin(ctx context.Context, in *authGoogleLoginReq) (*authGoogleLoginRes, error) {
	pair, err := h.auth.GoogleLogin(ctx, in.Body.Code, in.Body.CodeVerifier)
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}
	res := &authGoogleLoginRes{}
	res.Body.Data.AccessToken = pair.AccessToken
	res.Body.Data.RefreshToken = pair.RefreshToken
	res.Body.Data.ExpiresIn = pair.ExpiresIn
	return res, nil
}

func (h *authHandler) handleGoogleConfig(ctx context.Context, in *authGoogleConfigReq) (*authGoogleConfigRes, error) {
	clientID, redirectURI := h.auth.GoogleConfig()
	res := &authGoogleConfigRes{}
	res.Body.ClientID = clientID
	res.Body.RedirectURI = redirectURI
	return res, nil
}
