package mq

const (
	LoginQueue    = "auth_login"
	ValidateQueue = "auth_validate"
)

type AuthRequest struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token,omitempty"`
}

type AuthResponse struct {
	Token    string `json:"token,omitempty"`
	Success  bool   `json:"success,omitempty"`
	Message  string `json:"message,omitempty"`
	Valid    bool   `json:"valid,omitempty"`
	UserID   int    `json:"user_id,omitempty"`
	Username string `json:"username,omitempty"`
}
