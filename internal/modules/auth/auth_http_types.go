package auth

type authUserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// Register

type authRegisterReq struct {
	Body struct {
		Email    string `json:"email" format:"email" minLength:"3" maxLength:"254" doc:"User email address"`
		Password string `json:"password" minLength:"8" maxLength:"72" doc:"User password (8-72 chars)"`
	}
}

type authRegisterRes struct {
	SetCookie []string `header:"Set-Cookie"`
	Body      struct {
		User authUserResponse `json:"user"`
	}
}

// Login

type authLoginReq struct {
	Body struct {
		Email    string `json:"email" format:"email" minLength:"3" maxLength:"254" doc:"User email address"`
		Password string `json:"password" minLength:"8" maxLength:"72" doc:"User password (8-72 chars)"`
	}
}

type authLoginRes struct {
	SetCookie []string `header:"Set-Cookie"`
	Body      struct {
		User authUserResponse `json:"user"`
	}
}

// Refresh

type authRefreshReq struct {
	Cookie string `header:"Cookie"`
}

type authRefreshRes struct {
	SetCookie []string `header:"Set-Cookie"`
	Body      struct {
		User authUserResponse `json:"user"`
	}
}

// Logout

type authLogoutReq struct {
	Cookie string `header:"Cookie"`
}

type authLogoutRes struct {
	SetCookie []string `header:"Set-Cookie"`
}

// Delete account

type authDeleteAccountReq struct{}

type authDeleteAccountRes struct{}

// Google Login

type authGoogleLoginReq struct {
	Body struct {
		Code         string `json:"code" minLength:"1" doc:"Google authorization code"`
		CodeVerifier string `json:"code_verifier" minLength:"1" doc:"PKCE code verifier"`
	}
}

type authGoogleLoginRes struct {
	SetCookie []string `header:"Set-Cookie"`
	Body      struct {
		User authUserResponse `json:"user"`
	}
}

// Google Config

type authGoogleConfigReq struct{}

type authGoogleConfigRes struct {
	Body struct {
		ClientID    string `json:"client_id"`
		RedirectURI string `json:"redirect_uri"`
	}
}
