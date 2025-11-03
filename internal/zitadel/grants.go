package zitadel

import (
	"encoding/json"
	"fmt"
	"strings"
)

type GrantResults struct {
	Details Details `json:"details,omitempty"`
	Result  []Grant `json:"result,omitempty"`
}

type Grant struct {
	ID                 string   `json:"id,omitempty"`
	RoleKeys           []string `json:"roleKeys,omitempty"`
	State              string   `json:"state,omitempty"`
	UserID             string   `json:"userId,omitempty"`
	UserName           string   `json:"userName,omitempty"`
	FirstName          string   `json:"firstName,omitempty"`
	LastName           string   `json:"lastName,omitempty"`
	Email              string   `json:"email,omitempty"`
	DisplayName        string   `json:"displayName,omitempty"`
	OrgID              string   `json:"orgId,omitempty"`
	OrgName            string   `json:"orgName,omitempty"`
	OrgDomain          string   `json:"orgDomain,omitempty"`
	ProjectID          string   `json:"projectId,omitempty"`
	ProjectName        string   `json:"projectName,omitempty"`
	PreferredLoginName string   `json:"preferredLoginName,omitempty"`
	UserType           string   `json:"userType,omitempty"`
	GrantedOrgID       string   `json:"grantedOrgId,omitempty"`
	GrantedOrgName     string   `json:"grantedOrgName,omitempty"`
	GrantedOrgDomain   string   `json:"grantedOrgDomain,omitempty"`
}

func (c *Client) ListGrants() (*GrantResults, error) {

	if c.grants != nil {
		return c.grants, nil
	}

	return c.fetchGrants()
}

func (c *Client) fetchGrants() (*GrantResults, error) {

	url := fmt.Sprintf("%s/management/v1/users/grants/_search", c.url)
	method := "POST"

	rawBody, err := c.doRequest(method, url, strings.NewReader(payload))
	if err != nil {
		return nil, err
	}

	res := &GrantResults{}

	err = json.Unmarshal(rawBody, res)

	c.grants = res

	return res, err
}
