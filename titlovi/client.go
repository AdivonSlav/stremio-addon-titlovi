package titlovi

import (
	"encoding/json"
	"fmt"
	"go-titlovi/logger"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var LangMap = map[string]string{
	"Bosanski": "bos",
	"Hrvatski": "hrv",
	"Srpski":   "srb",
	"Cirilica": "cpb",
	"English":  "eng",
}

type Client struct {
	titloviApi string

	titloviUsername string
	titloviPassword string
	loginData       LoginData
}

func NewClient(titloviUsername string, titloviPassword string) *Client {
	return &Client{
		titloviUsername: titloviUsername,
		titloviPassword: titloviPassword,
		titloviApi:      "https://kodi.titlovi.com/api/subtitles",
	}

}

func (c *Client) Login(force bool) error {
	if c.loginData.Token != "" && !force {
		expirationTime, err := time.Parse(time.RFC3339, c.loginData.ExpirationDate)
		if err != nil {
			logger.LogError.Printf("Login: unable to parse '%s' to an expiration date: %s", c.loginData.ExpirationDate, err)
		}

		if expirationTime.Before(time.Now()) {
			return nil
		}
	}

	params := url.Values{}
	params.Add("username", c.titloviUsername)
	params.Add("password", c.titloviPassword)
	url := fmt.Sprintf("%s/gettoken?%s", c.titloviApi, params.Encode())

	resp, err := http.Post(url, "application/x-www-form-urlencoded", nil)
	if err != nil {
		err = fmt.Errorf("Login: failed to login: %w", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		logger.LogError.Printf("Login: %d, %s: %s", resp.StatusCode, url, resp.Status)
		return fmt.Errorf("Login: failed to login with message: %s", resp.Status)
	} else {
		logger.LogInfo.Printf("Login: %d, %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Login: failed to read response body: %w", err)
	}

	err = json.Unmarshal(body, &c.loginData)
	if err != nil {
		return fmt.Errorf("Login: failed to unmarshal respone body: %w", err)
	}

	logger.LogInfo.Printf("Login: received new login data : %+v", c.loginData)

	return err
}

func (c *Client) Search(imdbId string, languages []string) ([]SubtitleData, error) {
	params := url.Values{}
	params.Add("token", c.loginData.Token)
	params.Add("userId", string(c.loginData.UserId))
	params.Add("imdbID", imdbId)
	params.Add("lang", strings.Join(languages, "|"))

	url := fmt.Sprintf("%s/search?%s", c.titloviApi, params.Encode())
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Search: failed to search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		logger.LogError.Printf("Search: %d, %s: %s", resp.StatusCode, url, resp.Status)
		return nil, fmt.Errorf("Search: failed to search with message: %s", resp.Status)
	} else {
		logger.LogInfo.Printf("Search: %d, %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Search: failed to read response body: %w", err)
	}

	subtitleResponse := SubtitleDataResponse{}
	err = json.Unmarshal(body, &subtitleResponse)
	if err != nil {
		return nil, fmt.Errorf("Search: failed to unmarshal response body: %w", err)
	}

	return subtitleResponse.Subtitles, nil
}
