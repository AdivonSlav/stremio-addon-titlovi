package titlovi

import "time"

type LoginData struct {
	Username       string    `json:"UserName"`
	UserId         string    `json:"UserId"`
	Token          string    `json:"Token"`
	ExpirationDate time.Time `json:"ExpirationDate"`
}

type LoginRequest struct {
	Username string `json:"UserName"`
	Password string `json:"Password"`
}
