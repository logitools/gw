package cookiesession

type KVDBBackendAPIData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserIDStr    string `json:"uid"`
	Key          string `json:"-"`
}

// ToDo: make tokens fields as map[string]string for many different APIs
