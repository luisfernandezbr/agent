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
	Selected    bool   `json:"selected"`
}

// Validate is a control interface for validating a configuration before enroll
type Validate interface {
	Control
	Config() Config
	// State is a customer specific state object for this integration and customer
	State() State
	// Logger the logger object to use in the integration
	Logger() Logger
}
