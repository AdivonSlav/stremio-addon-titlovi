package titlovi

import (
	"encoding/json"
	"fmt"
	"go-titlovi/internal/config"
	"go-titlovi/internal/logger"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/avast/retry-go"
)

// Client is an implementation to fetch search results from Titlovi.com.
type Client struct {
	loginData LoginData

	retryAttempts uint
	retryDelay    time.Duration
}

func NewClient(retryAttempts uint, retryDelay time.Duration) *Client {
	return &Client{
		retryAttempts: retryAttempts,
		retryDelay:    retryDelay,
	}
}

// Login attempts a login to the Titlovi.com API and internally stores the retrieved token if succesful.
func (c *Client) Login(username, password string) error {
	params := url.Values{}
	params.Add("username", username)
	params.Add("password", password)
	url := fmt.Sprintf("%s/gettoken?%s", config.TitloviApi, params.Encode())

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

// Search performs a search on the Titlovi.com API and returns a slice of titlovi.SubtitleData if successful.
func (c *Client) Search(imdbId string, season, episode string, languages []string, username, password string) ([]SubtitleData, error) {
	params := url.Values{}
	params.Add("token", c.loginData.Token)
	params.Add("userid", strconv.Itoa(int(c.loginData.UserId)))
	params.Add("query", imdbId)
	params.Add("lang", strings.Join(languages, "|"))

	if season != "" {
		params.Add("season", season)
	}
	if episode != "" {
		params.Add("episode", episode)
	}

	url := fmt.Sprintf("%s/search?%s", config.TitloviApi, params.Encode())
	var body []byte

	err := retry.Do(func() error {
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("failed to search: %w", err)
		}

		if resp.StatusCode == 401 {
			// Retry search with new token
			loginErr := c.Login(username, password)
			if loginErr != nil {
				logger.LogError.Printf("failed to search due to login failure: %s", loginErr.Error())
			}

			params.Set("token", c.loginData.Token)
			params.Set("userid", strconv.Itoa(int(c.loginData.UserId)))
			url = fmt.Sprintf("%s/search?%s", config.TitloviApi, params.Encode())
		}

		if resp.StatusCode > 299 {
			return fmt.Errorf("status %d, %s: %s", resp.StatusCode, url, resp.Status)
		} else {
			body, err = io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response body: %w", err)
			}
		}

		return nil
	}, retry.Attempts(c.retryAttempts), retry.Delay(c.retryDelay))
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

// Download downloads a subtitle from Titlovi.com based on the provided type and ID and returns it as a blob.
func (c *Client) Download(mediaType string, mediaId string) ([]byte, error) {
	url := fmt.Sprintf("%s/?type=%s&mediaid=%s", config.TitloviDownload, mediaType, mediaId)
	var body []byte

	err := retry.Do(func() error {
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("failed to download subtitle at %s: %s", url, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode > 299 {
			return fmt.Errorf("status %d, %s: %s", resp.StatusCode, url, resp.Status)
		} else {
			body, err = io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response body: %w", err)
			}
		}

		return nil
	}, retry.Attempts(c.retryAttempts), retry.Delay(c.retryDelay))
	if err != nil {
		return nil, fmt.Errorf("Download: %w", err)
	}

	return body, nil
}
