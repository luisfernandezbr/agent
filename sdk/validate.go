package sdk

// ValidatedAccount is a result that can be sent back to the integration UI for a validate account
type ValidatedAccount struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	AvatarURL   string `json:"avatarUrl"`
	TotalCount  int    `json:"totalCount"`
	Type        string `json:"type"`
	Public      bool   `json:"public"`
}

// Validate is a control interface for validating a configuration before enroll
type Validate interface {
	Control
	Config() Config
}
