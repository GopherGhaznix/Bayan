package config

type BaseConfiguration struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Key      []byte `json:"key"`

	WebsiteRoot string `json:"website_root"`
}

var (
	BaseConfig BaseConfiguration
)
