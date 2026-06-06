package auth

// Token data shared by all token-issuing responses.
type tokenData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in" doc:"Unix timestamp when the access token expires"`
}

// Register

type authRegisterReq struct {
	Body struct {
		Email    string `json:"email" format:"email" minLength:"3" maxLength:"254" doc:"User email address"`
		Password string `json:"password" minLength:"8" maxLength:"72" doc:"User password (8-72 chars)"`
	}
}

type authRegisterBody struct {
	Data tokenData `json:"data"`
}

type authRegisterRes struct {
	Body authRegisterBody
}

// Login

type authLoginReq struct {
	Body struct {
		Email    string `json:"email" format:"email" minLength:"3" maxLength:"254" doc:"User email address"`
		Password string `json:"password" minLength:"8" maxLength:"72" doc:"User password (8-72 chars)"`
	}
}

type authLoginBody struct {
	Data tokenData `json:"data"`
}

type authLoginRes struct {
	Body authLoginBody
}

// Refresh

type authRefreshReq struct {
	Body struct {
		RefreshToken string `json:"refresh_token" minLength:"1"`
	}
}

type authRefreshBody struct {
	Data tokenData `json:"data"`
}

type authRefreshRes struct {
	Body authRefreshBody
}

// Delete account

type authDeleteAccountReq struct{}

type authDeleteAccountRes struct{}
