package zitadel

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type LoginResult struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// TODO: maybe caching the session will make things a bit faster?
func (c *Client) Login(username, password string) (bool, error) {
	// go func() {
	// 	err := c.fetchAll()
	// 	if err != nil {
	// 		c.log.Err(err).Msg("Cannot fetch during data during bind")
	// 	}
	// }()

	c.log.Debug().Str("username", username).Msg("Logging in...")
	hash := c.hashLogin(username, password)

	// TODO: maybe we should check against the API if the session still exists...
	if ct, ok := c.sessionCache.Load(hash); ok {
		timeSince := time.Since(ct)
		c.log.Debug().Str("username", username).Msg(fmt.Sprintf("Existing session available, %.2f seconds old", timeSince.Seconds()))
		if timeSince.Seconds() > 3600 {
			c.log.Info().Str("username", username).Msg("Session expired, doing a new login")
			return c.newSession(username, password)
		}
		return true, nil
	}

	c.log.Debug().Str("username", username).Msg("No existing session available")
	return c.newSession(username, password)
}

func (c *Client) newSession(username, password string) (bool, error) {
	url := fmt.Sprintf("%s/v2/sessions", c.url)
	method := "POST"

	loginQuery := fmt.Sprintf(`{
	  "checks": {
	    "user": {
	      "loginName": "%s"
	    },
	    "password": {
	      "password": "%s"
	    }
	  },
	  "lifetime": "3600s"
	}`, username, password)

	rawBody, err := c.doRequest(method, url, strings.NewReader(loginQuery))
	if err != nil {
		c.log.Debug().Str("username", username).Str("response", string(rawBody)).Msg("Error in querying Zitadel")
		return false, err
	}

	c.log.Debug().Str("username", username).Str("response", string(rawBody)).Msg("Response from Zitadel")

	res := &LoginResult{}

	err = json.Unmarshal(rawBody, res)
	if err != nil {
		c.log.Debug().Str("username", username).Str("response", string(rawBody)).Msg("Cannot unmarshal response")
		return false, err
	}

	if res.Code != 0 {
		c.log.Debug().Str("username", username).Msg("Login failed")
		return false, nil
	}

	c.sessionCache.Store(c.hashLogin(username, password), time.Now())

	c.log.Debug().Str("username", username).Msg("Login successful")

	return true, nil
}

func (c *Client) hashLogin(username, password string) string {
	h := sha256.New()
	h.Write([]byte(username + password))
	return string(h.Sum(nil))
}
