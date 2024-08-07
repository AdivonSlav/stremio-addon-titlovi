package titlovi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Client struct {
	titloviApi string

	titloviUsername string
	titloviPassword string
	loginData       LoginData
}

func New(titloviUsername string, titloviPassword string) *Client {
	return &Client{
		titloviApi: "https://kodi.titlovi.com/api/subtitles",
	}

}

func (c *Client) login() error {
	payload, err := json.Marshal(&LoginRequest{
		Username: c.titloviUsername,
		Password: c.titloviPassword,
	})
	if err != nil {
		log.Printf("login: failed to marshal login request: %s\n", err.Error())
	}

	resp, err := http.Post(fmt.Sprintf("%s/gettoken", c.titloviPassword), "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("login: failed to login: %s\n", err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("login: failed to read response body: %s\n", err.Error())
	}

	err = json.Unmarshal(body, &c.loginData)
	if err != nil {
		log.Printf("login: failed to unmarshal response body: %s\n", err.Error())
	}

	return err
}
