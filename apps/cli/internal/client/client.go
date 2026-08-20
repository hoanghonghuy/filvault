package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Error is a parsed API error body (spec 04 §1.2).
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Client talks to the Filvault API over HTTP.
type Client struct {
	Base       string
	HTTP       *http.Client
	Access     string
	Refresh    string
	OnRefresh  func(access, refresh string)
	OnAuthFail func()
}

// New returns a Client with a default HTTP client.
func New(base string) *Client {
	return &Client{Base: base, HTTP: &http.Client{}}
}

// Do performs a request, transparently refreshing on 401 once.
func (c *Client) Do(method, path string, body any, out any) error {
	req, err := c.newRequest(method, path, body)
	if err != nil {
		return err
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode == http.StatusUnauthorized && c.Refresh != "" {
		res.Body.Close()
		if err := c.refresh(); err != nil {
			return err
		}
		req, err = c.newRequest(method, path, body)
		if err != nil {
			return err
		}
		res, err = c.HTTP.Do(req)
		if err != nil {
			return err
		}
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNoContent {
		return nil
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		return parseError(data, res.StatusCode)
	}
	if out != nil {
		return json.Unmarshal(data, out)
	}
	return nil
}

func (c *Client) newRequest(method, path string, body any) (*http.Request, error) {
	var buf io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, c.Base+path, buf)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Access != "" {
		req.Header.Set("Authorization", "Bearer "+c.Access)
	}
	return req, nil
}

func (c *Client) refresh() error {
	req, err := c.newRequest(http.MethodPost, "/auth/refresh", map[string]string{
		"refreshToken": c.Refresh,
	})
	if err != nil {
		return err
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		if c.OnAuthFail != nil {
			c.OnAuthFail()
		}
		return parseError(data, res.StatusCode)
	}
	var session struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	}
	if err := json.Unmarshal(data, &session); err != nil {
		return err
	}
	c.Access = session.AccessToken
	c.Refresh = session.RefreshToken
	if c.OnRefresh != nil {
		c.OnRefresh(session.AccessToken, session.RefreshToken)
	}
	return nil
}

func parseError(data []byte, status int) error {
	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &body); err != nil || body.Error.Code == "" {
		return &Error{Code: "INTERNAL", Message: http.StatusText(status), Status: status}
	}
	return &Error{Code: body.Error.Code, Message: body.Error.Message, Status: status}
}
