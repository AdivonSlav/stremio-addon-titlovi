package web

import "strings"

type Credentials struct {
	Username string
	Password string
	Errors   map[string]string
}

func (c *Credentials) Validate() bool {
	c.Errors = make(map[string]string)

	if strings.TrimSpace(c.Username) == "" {
		c.Errors["Username"] = "You must enter a username"
	}

	if strings.TrimSpace(c.Password) == "" {
		c.Errors["Password"] = "You must enter a password"
	}

	return len(c.Errors) == 0
}
