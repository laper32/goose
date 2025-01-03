package steam

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	APIUrlDefault  = "https://api.steampowered.com"
	TimeoutDefault = 1
)

type APIClient interface {
	IsPlayingSharedGame(steamID uint64, appID int) (res PlayingSharedGame, err error)
	GetUserGroupList(steamID uint64) (res UserGroupList, err error)
}

type API struct {
	APIUrl string
	APIKey string
	// API request timeout (sec)
	Timeout int
	// Proxy
	HttpProxy string
}

type APIRequest struct {
	Service string
	Method  string
	Version int
}

func NewAPIClient(url, key string, timeout int) *API {
	if len(url) == 0 {
		url = APIUrlDefault
	}
	if timeout == 0 {
		timeout = TimeoutDefault
	}

	return &API{
		APIUrl:  url,
		APIKey:  key,
		Timeout: timeout,
	}
}

func (a *API) request(req APIRequest, values url.Values, v interface{}) error {
	if values == nil {
		return ErrInvalidRequestValues
	}
	values.Add("format", "json")
	values.Add("key", a.APIKey)

	apiURL := fmt.Sprintf("%s/%s/%s/v%d/?%s", a.APIUrl, req.Service, req.Method, req.Version, values.Encode())
	client := http.Client{Timeout: time.Duration(a.Timeout) * time.Second}
	if len(a.HttpProxy) > 0 {
		proxy, err := url.Parse("http://" + a.HttpProxy)
		if err != nil {
			return err
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}
	}
	resp, err := client.Get(apiURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s > %v", apiURL, resp.StatusCode)
	}

	d := json.NewDecoder(resp.Body)
	return d.Decode(v)
}
