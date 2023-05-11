package cosmos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

type RestClient struct {
	baseURL url.URL
	httpDo  func(req *http.Request) (*http.Response, error)
}

func NewRestClient(c *http.Client, baseURL url.URL) *RestClient {
	return &RestClient{
		baseURL: baseURL,
		httpDo:  c.Do,
	}
}

// response must be a pointer to a datatype (typically a struct)
func (c RestClient) get(ctx context.Context, url string, response any) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("malformed request: %w", err)
	}
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
