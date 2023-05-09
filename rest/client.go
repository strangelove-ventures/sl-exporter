// Package rest provides a client for the Cosmos REST API (aka LCD - light client daemon).
package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type Client struct {
	httpDo func(req *http.Request) (*http.Response, error)
}

func NewClient(c *http.Client) *Client {
	return &Client{httpDo: c.Do}
}

// response must be a pointer to a datatype (typically a struct)
func (c *Client) get(ctx context.Context, url string, response any) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("malformed request: %w", err)
	}
	req = req.WithContext(ctx)
	resp, err := c.httpDo(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return fmt.Errorf("malformed json: %w", err)
	}
	return nil
}
