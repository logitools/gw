package login

type WebLoginSessionInfoWithBackendAPI struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserIDStr    string `json:"uid"`
	Key          string `json:"-"`
}

type WebLoginSessionInfoUID struct {
	UserIDStr string `json:"uid"`
	Key       string `json:"-"`
}
