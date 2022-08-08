package rest

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	URL string
}

func NewClient(url string) *Client {
	return &Client{URL: url}
}

func (sc Client) GET(ctx context.Context, uri string) (string, error) {
	getReq, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s%s", sc.URL, uri), nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(getReq)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
