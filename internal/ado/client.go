package ado

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type Client struct {
	Org        string
	Project    string
	BaseURL    string
	PAT        string
	HTTPClient *http.Client
}

func NewClient(org, project, baseURL, pat string) *Client {
	return &Client{
		Org:        org,
		Project:    project,
		BaseURL:    strings.TrimRight(baseURL, "/"),
		PAT:        pat,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) authHeader() string {
	token := ":" + c.PAT
	b64 := base64.StdEncoding.EncodeToString([]byte(token))
	return "Basic " + b64
}

func (c *Client) makeURL(apiPath string, params map[string]string) (string, error) {
	cleanPath := strings.TrimPrefix(path.Clean("/"+apiPath), "/")
	fullURL := fmt.Sprintf("%s/%s/%s/%s", c.BaseURL, c.Org, c.Project, cleanPath)
	u, err := url.Parse(fullURL)
	if err != nil {
		return "", err
	}
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (c *Client) do(method, apiPath string, params map[string]string, body any) (*http.Response, error) {
	fullURL, err := c.makeURL(apiPath, params)
	if err != nil {
		return nil, err
	}

	var reqBody io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("ado request failed: %s %s: status %d body=%q", method, fullURL, resp.StatusCode, string(snippet))
	}
	return resp, nil
}

func (c *Client) RequestJSON(method, apiPath string, params map[string]string, body any, out any) error {
	resp, err := c.do(method, apiPath, params, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) RequestText(method, apiPath string, params map[string]string) (string, error) {
	resp, err := c.do(method, apiPath, params, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
