package zitadel

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type ProjectResults struct {
	Details Details         `json:"details,omitempty"`
	Result  []ProjectResult `json:"result,omitempty"`
}

type ProjectDetails struct {
	Sequence      string    `json:"sequence,omitempty"`
	CreationDate  time.Time `json:"creationDate,omitempty"`
	ChangeDate    time.Time `json:"changeDate,omitempty"`
	ResourceOwner string    `json:"resourceOwner,omitempty"`
}

type ProjectResult struct {
	ID                   string         `json:"id,omitempty"`
	Details              ProjectDetails `json:"details,omitempty"`
	Name                 string         `json:"name,omitempty"`
	State                string         `json:"state,omitempty"`
	ProjectRoleCheck     bool           `json:"projectRoleCheck,omitempty"`
	ProjectRoleAssertion bool           `json:"projectRoleAssertion,omitempty"`
}

type RoleResults struct {
	Details Details      `json:"details,omitempty"`
	Result  []RoleResult `json:"result,omitempty"`
}

type RoleDetails struct {
	Sequence      string    `json:"sequence,omitempty"`
	CreationDate  time.Time `json:"creationDate,omitempty"`
	ChangeDate    time.Time `json:"changeDate,omitempty"`
	ResourceOwner string    `json:"resourceOwner,omitempty"`
}
type RoleResult struct {
	Key         string      `json:"key,omitempty"`
	Details     RoleDetails `json:"details,omitempty"`
	DisplayName string      `json:"displayName,omitempty"`
	Group       string      `json:"group,omitempty"`
}

func (c *Client) ListProjects() (*ProjectResults, error) {

	if c.projects != nil {
		return c.projects, nil
	}

	return c.fetchProjects()
}

func (c *Client) fetchProjects() (*ProjectResults, error) {

	c.log.Debug().Msg("Fetching all projects")

	url := fmt.Sprintf("%s/management/v1/projects/_search", c.url)
	method := "POST"

	rawBody, err := c.doRequest(method, url, strings.NewReader(payload))
	if err != nil {
		return nil, err
	}

	res := &ProjectResults{}

	err = json.Unmarshal(rawBody, res)

	c.projects = res

	return res, err
}

func (c *Client) ListRoles(projectID string) (*RoleResults, error) {

	if r, ok := c.roles.Load(projectID); ok {
		return r, nil
	}

	return c.fetchRoles(projectID)
}

func (c *Client) fetchRoles(projectID string) (*RoleResults, error) {

	c.log.Debug().Str("project", projectID).Msg("Fetching roles")

	url := fmt.Sprintf("%s/management/v1/projects/%s/roles/_search", c.url, projectID)
	method := "POST"

	rawBody, err := c.doRequest(method, url, strings.NewReader(payload))
	if err != nil {
		return nil, err
	}

	res := &RoleResults{}

	err = json.Unmarshal(rawBody, res)

	c.roles.Store(projectID, res)

	return res, err
}

func (c *Client) fetchAllRolesAndProjects() error {

	_, err := c.fetchProjects()
	if err != nil {
		return err
	}

	for _, p := range c.projects.Result {
		_, err := c.fetchRoles(p.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
