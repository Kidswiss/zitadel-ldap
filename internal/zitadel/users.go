package zitadel

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type UserResults struct {
	Details Details `json:"details,omitempty"`
	Result  []User  `json:"result,omitempty"`
}
type Details struct {
	TotalResult string    `json:"totalResult,omitempty"`
	Timestamp   time.Time `json:"timestamp,omitempty"`
}

type Profile struct {
	GivenName         string `json:"givenName,omitempty"`
	FamilyName        string `json:"familyName,omitempty"`
	NickName          string `json:"nickName,omitempty"`
	DisplayName       string `json:"displayName,omitempty"`
	PreferredLanguage string `json:"preferredLanguage,omitempty"`
	Gender            string `json:"gender,omitempty"`
}
type Email struct {
	Email      string `json:"email,omitempty"`
	IsVerified bool   `json:"isVerified,omitempty"`
}

type Human struct {
	Profile         Profile   `json:"profile,omitempty"`
	Email           Email     `json:"email,omitempty"`
	PasswordChanged time.Time `json:"passwordChanged,omitempty"`
	Groups          []string  `json:"groups,omitempty"`
}
type Machine struct {
	Name            string `json:"name,omitempty"`
	Description     string `json:"description,omitempty"`
	HasSecret       bool   `json:"hasSecret,omitempty"`
	AccessTokenType string `json:"accessTokenType,omitempty"`
}

type User struct {
	UserID             string   `json:"userId,omitempty"`
	State              string   `json:"state,omitempty"`
	Username           string   `json:"username,omitempty"`
	LoginNames         []string `json:"loginNames,omitempty"`
	PreferredLoginName string   `json:"preferredLoginName,omitempty"`
	Human              *Human   `json:"human,omitempty"`
	Machine            *Machine `json:"machine,omitempty"`
}

func (c *Client) ListUsers() (*UserResults, error) {
	if c.users != nil {
		return c.users, nil
	}

	return c.fetchUsers()
}

func (c *Client) FindUserByName(name string) (*UserResults, error) {

	if c.users == nil {
		_, err := c.fetchUsers()
		if err != nil {
			return nil, err
		}
	}

	res := c.users

	for _, u := range res.Result {
		if u.Human != nil {
			if name == u.Username {
				return &UserResults{
					Result: []User{
						u,
					},
				}, nil
			}
		}
	}

	return nil, nil
}

func (c *Client) FindUserByMail(mail string) (*UserResults, error) {

	if c.users == nil {
		_, err := c.fetchUsers()
		if err != nil {
			return nil, err
		}
	}

	res := c.users

	for _, u := range res.Result {
		if u.Human != nil {
			if mail == u.Human.Email.Email {
				return &UserResults{
					Result: []User{
						u,
					},
				}, nil
			}
		}
	}

	return nil, nil

}

func (c *Client) fetchUsers() (*UserResults, error) {

	c.log.Debug().Msg("Fetching all users")

	url := fmt.Sprintf("%s/v2/users", c.url)
	method := "POST"

	body, err := c.doRequest(method, url, strings.NewReader(payload))
	if err != nil {
		return nil, err
	}

	parsed := UserResults{}

	err = json.Unmarshal(body, &parsed)
	if err != nil {
		return nil, err
	}

	grants, err := c.ListGrants()
	if err != nil {
		return nil, err
	}

	groupedGrants := map[string][]string{}

	for _, grant := range grants.Result {
		groupedGrants[grant.UserID] = append(groupedGrants[grant.UserID], grant.RoleKeys...)
	}

	finalUsers := UserResults{
		Result: []User{},
	}

	for _, user := range parsed.Result {
		groups := groupedGrants[user.UserID]
		if user.Human != nil {
			user.Human.Groups = groups
		}

		finalUsers.Result = append(finalUsers.Result, user)
	}

	c.users = &finalUsers

	return &finalUsers, nil
}
