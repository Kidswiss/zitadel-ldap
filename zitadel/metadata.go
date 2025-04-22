package zitadel

import (
	"encoding/json"
	"fmt"
	"strings"
)

type MetadataResults struct {
	Details Details          `json:"details,omitempty"`
	Result  []MetadataResult `json:"result,omitempty"`
}

type MetadataResult struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

func (c *Client) ListMetadata(id string) (*MetadataResults, error) {

	if m, ok := c.metadata[id]; ok {
		return m, nil
	}

	return c.fetchMetadata(id)

}

func (c *Client) fetchMetadata(id string) (*MetadataResults, error) {
	c.log.Debug().Str("id", id).Msg("Fetching metadata")

	url := fmt.Sprintf("%s/management/v1/users/%s/metadata/_search", c.url, id)
	method := "POST"

	rawBody, err := c.doRequest(method, url, strings.NewReader(payload))
	if err != nil {
		return nil, err
	}

	res := &MetadataResults{}

	err = json.Unmarshal(rawBody, res)

	c.metadata[id] = res

	return res, err
}

func (c *Client) fetchAllMetadata() error {
	users, err := c.ListUsers()
	if err != nil {
		return err
	}

	for _, u := range users.Result {
		if u.Human != nil {
			_, err := c.fetchMetadata(u.UserID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
