package cosmos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// RestClient is a client for the Cosmos REST API.
// To find a list of endpoints, try: https://docs.cosmos.network/swagger/
type RestClient struct {
	client HTTPClient
}

type HTTPClient interface {
	Get(ctx context.Context, path string) (*http.Response, error)
}

func NewRestClient(c HTTPClient) *RestClient {
	return &RestClient{
		client: c,
	}
}

// response must be a pointer to a datatype (typically a struct)
func (c RestClient) get(ctx context.Context, url string, response any) error {
	resp, err := c.client.Get(ctx, url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return fmt.Errorf("malformed json: %w", err)
	}
	return nil
}
