package cookiesession

type ExpireMode string

const (
	ExpireAbsolute ExpireMode = "absolute"
	ExpireSliding  ExpireMode = "sliding" // sliding expiration
)

type Conf struct {
	EncryptionKey string     `json:"enckey"`
	ExpireIn      int        `json:"expire_in"` // seconds
	ExpireMode    ExpireMode `json:"expire_mode"`

	// For Web Login Sessions
	LoginPath     string `json:"login_path"`
	MaxCntPerUser int64  `json:"max_cnt_per_user"`
}
