package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Client represent Service backed by an HTTP server living at the remote instance.
type Client struct {
	BaseURL   *url.URL
	UserAgent string
	AppKey    string

	httpClient *http.Client
}

//NewClient creates Service backed by an HTTP server living at the remote instance
func NewClient(httpClient *http.Client, baseURL, appKey string) (FFService, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	return &Client{
		BaseURL:    u,
		AppKey:     appKey,
		httpClient: httpClient,
	}, nil
}

//GetNPGroups implement Service
func (c *Client) GetNPGroups(ctx context.Context, statuses []int, fromTS int64) ([]NPGroup, error) {
	//https://fabrika-fotoknigi.ru/api/?appkey=e5ea49c386479f7c30f60e52e8b9107b&action=fk:get_groups_by_status_and_period&debug=1&status=40&start=1574334313
	data := url.Values{}
	data.Set("action", "fk:get_groups_by_status_and_period")
	data.Set("start", strconv.FormatInt(fromTS, 10))
	for _, s := range statuses {
		data.Add("status[]", strconv.Itoa(s))
	}
	res := []NPGroup{}
	rq, err := c.newRequest(ctx, "POST", "", data)
	if err != nil {
		return nil, err
	}
	r, err := c.do(rq, &res)
	if err != nil {
		return nil, err
	}
	if r.StatusCode != http.StatusOK {
		return nil, statusError(r.StatusCode)
	}
	return res, err
}

func statusError(code int) error {
	return fmt.Errorf("Wrong http status %d. %s", code, http.StatusText(code))
}

func (c *Client) newRequest(ctx context.Context, method, path string, data url.Values) (*http.Request, error) {
	rel := &url.URL{Path: path}
	u := c.BaseURL.ResolveReference(rel)
	if data == nil {
		data = url.Values{}
	}
	data.Set("appkey", c.AppKey)
	var reader io.Reader
	if method == "POST" {
		reader = strings.NewReader(data.Encode())
	} else {
		u.RawQuery = data.Encode()
	}
	//fmt.Println(data.Encode())
	req, err := http.NewRequestWithContext(ctx, method, u.String(), reader)
	if err != nil {
		return nil, err
	}

	if method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req.Header.Set("Accept", "application/json")
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var raw bytes.Buffer
	tee := io.TeeReader(resp.Body, &raw)
	err = json.NewDecoder(tee).Decode(v)
	if err != nil {
		ae := apiError{}
		if e := json.NewDecoder(&raw).Decode(&ae); e == nil {
			//intrenal api error
			err = fmt.Errorf("Error: %s; Code: %d; Exception: %s", ae.Error, ae.Code, ae.Exception)
		} else {
			err = fmt.Errorf("%s; Response: %s", err.Error(), raw.String())
		}
	}
	return resp, err
}

type apiError struct {
	Code      int    `json:"code"`
	Error     string `json:"error"`
	Exception string `json:"exception"`
}
