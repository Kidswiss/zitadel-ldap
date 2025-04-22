package zitadel

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type Client struct {
	hclient      http.Client
	pat          string
	url          string
	users        *UserResults
	grants       *GrantResults
	projects     *ProjectResults
	roles        map[string]*RoleResults
	sessionCache map[string]time.Time
	metadata     map[string]*MetadataResults
	log          *zerolog.Logger
}

func NewClient(url, pat string, log *zerolog.Logger) *Client {
	c := &Client{
		hclient:      http.Client{},
		pat:          pat,
		url:          url,
		log:          log,
		sessionCache: map[string]time.Time{},
		roles:        map[string]*RoleResults{},
		metadata:     map[string]*MetadataResults{},
	}

	c.fetchAll()
	c.fetchTimer()

	return c
}

var payload = `{
	"query": {
	  "offset": "0",
	  "limit": 100,
	  "asc": true
	}
  }`

func (c Client) doRequest(method, url string, payload io.Reader) ([]byte, error) {

	c.log.Debug().Msg("Sending request to Zitadel")

	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.pat))

	res, err := c.hclient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *Client) fetchAll() error {
	c.log.Debug().Msg("Fetching all data")

	err := c.fetchAllRolesAndProjects()
	if err != nil {
		return err
	}

	_, err = c.fetchGrants()
	if err != nil {
		return err
	}

	_, err = c.fetchUsers()
	if err != nil {
		return err
	}

	err = c.fetchAllMetadata()
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) fetchTimer() {
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				err := c.fetchAll()
				if err != nil {
					c.log.Err(err).Msg("Fetch timer failed")
				}
			}
		}
	}()
}
