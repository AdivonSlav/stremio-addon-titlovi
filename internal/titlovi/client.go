package titlovi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-titlovi/internal/config"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/avast/retry-go"
)

// Client is an implementation to fetch search results from Titlovi.com.
type Client struct {
	// A map of usernames and their corresponding tokens.
	clientLoginData map[string]*LoginData
	mtx             sync.RWMutex
	http            http.Client
	retryAttempts   uint
	retryDelay      time.Duration
}

func NewClient(retryAttempts uint, retryDelay time.Duration) *Client {
	return &Client{
		clientLoginData: make(map[string]*LoginData, 0),
		retryAttempts:   retryAttempts,
		retryDelay:      retryDelay,
	}
}

// Login attempts a login to the Titlovi.com API and internally stores the retrieved token if successful.
func (c *Client) Login(ctx context.Context, username, password string) (*LoginData, error) {
	params := url.Values{}
	params.Add("username", username)
	params.Add("password", password)
	url := fmt.Sprintf("%s/gettoken?%s", config.TitloviApi, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("post login: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("login: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("response read: %w", err)
	}

	loginData := &LoginData{}
	err = json.Unmarshal(body, loginData)
	if err != nil {
		return nil, fmt.Errorf("response unmarshal: %w", err)
	}

	return loginData, nil
}

// Search performs a search on the Titlovi.com API and returns a slice of titlovi.SubtitleData if successful.
func (c *Client) Search(ctx context.Context, imdbId, season, episode string, languages []string, username, password string) ([]SubtitleData, error) {
	d, err := c.getLoginData(ctx, username, password, false)
	if err != nil {
		return nil, fmt.Errorf("get login data: %w", err)
	}

	params := url.Values{}
	params.Add("token", d.Token)
	params.Add("userid", strconv.Itoa(int(d.UserId)))
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

	err = retry.Do(func() error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("create search request: %w", err)
		}
		resp, err := c.http.Do(req)
		if err != nil {
			return fmt.Errorf("get search: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == 401 {
			// Retry search with new token
			d, loginErr := c.getLoginData(ctx, username, password, true)
			if loginErr != nil {
				return fmt.Errorf("get login data retry: %w", err)
			}

			params.Set("token", d.Token)
			params.Set("userid", strconv.Itoa(int(d.UserId)))
			url = fmt.Sprintf("%s/search?%s", config.TitloviApi, params.Encode())
			return errors.New("retry with new token")
		}

		if resp.StatusCode > 299 {
			return fmt.Errorf("get: %s", resp.Status)
		} else {
			body, err = io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("response read: %w", err)
			}
		}

		return nil
	}, retry.Attempts(c.retryAttempts), retry.Delay(c.retryDelay))
	if err != nil {
		return nil, err
	}

	subtitleResponse := SubtitleDataResponse{}
	err = json.Unmarshal(body, &subtitleResponse)
	if err != nil {
		return nil, fmt.Errorf("response unmarshal: %w", err)
	}

	return subtitleResponse.Subtitles, nil
}

// Download downloads a subtitle from Titlovi.com based on the provided type and ID and returns it as a blob.
func (c *Client) Download(ctx context.Context, mediaType string, mediaId string) ([]byte, error) {
	url := fmt.Sprintf("%s/?type=%s&mediaid=%s", config.TitloviDownload, mediaType, mediaId)
	var body []byte

	err := retry.Do(func() error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("create download request: %w", err)
		}
		resp, err := c.http.Do(req)
		if err != nil {
			return fmt.Errorf("get subtitle: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode > 299 {
			return fmt.Errorf("get: %w", err)
		} else {
			body, err = io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("response read: %w", err)
			}
		}

		return nil
	}, retry.Attempts(c.retryAttempts), retry.Delay(c.retryDelay))
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *Client) getLoginData(ctx context.Context, username, password string, forceLogin bool) (*LoginData, error) {
	var err error

	c.mtx.RLock()
	d, ok := c.clientLoginData[username]
	c.mtx.RUnlock()

	// If we don't have it, get it
	if !ok || forceLogin {
		d, err = c.Login(ctx, username, password)
		if err != nil {
			return nil, fmt.Errorf("login: %w", err)
		}

		c.mtx.Lock()
		c.clientLoginData[username] = d
		c.mtx.Unlock()
	}

	return d, nil
}
