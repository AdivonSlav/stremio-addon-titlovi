package titlovi

import (
	"encoding/json"
	"fmt"
	"go-titlovi/logger"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/avast/retry-go"
)

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

func (c *Client) Login() error {
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
	params.Add("userid", strconv.Itoa(int(c.loginData.UserId)))
	params.Add("query", imdbId)
	params.Add("lang", strings.Join(languages, "|"))

	url := fmt.Sprintf("%s/search?%s", c.titloviApi, params.Encode())
	var body []byte

	err := retry.Do(func() error {
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("failed to search: %w", err)
		}

		if resp.StatusCode == 401 {
			// Retry search with new token
			loginErr := c.Login()
			if loginErr != nil {
				logger.LogError.Printf("Search: failed to search due to login failure: %s", loginErr.Error())
			}

			params.Set("token", c.loginData.Token)
			params.Set("userid", strconv.Itoa(int(c.loginData.UserId)))
			url = fmt.Sprintf("%s/search?%s", c.titloviApi, params.Encode())
		}

		if resp.StatusCode > 299 {
			return fmt.Errorf("status %d, %s: %s", resp.StatusCode, url, resp.Status)
		} else {
			body, err = io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response body: %w", err)
			}
		}

		logger.LogInfo.Printf("Search: %d, %s", resp.StatusCode, url)

		return nil
	}, retry.Attempts(3), retry.Delay(300*time.Millisecond))
	if err != nil {
		return nil, fmt.Errorf("Search: %w", err)
	}

	subtitleResponse := SubtitleDataResponse{}
	err = json.Unmarshal(body, &subtitleResponse)
	if err != nil {
		return nil, fmt.Errorf("Search: failed to unmarshal response body: %w", err)
	}

	return subtitleResponse.Subtitles, nil
}
