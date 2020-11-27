package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	if ctx == nil {
		ctx = context.Background()
	}
	//https://fabrika-fotoknigi.ru/api/?appkey=e5ea49c386479f7c30f60e52e8b9107b&action=fk:get_groups_by_status_and_period&debug=1&status=40&start=1574334313
	data := url.Values{}
	data.Set("action", "fk:get_groups_by_status_and_period")
	data.Set("start", strconv.FormatInt(fromTS, 10))
	for _, s := range statuses {
		data.Add("status[]", strconv.Itoa(s))
	}
	res := []NPGroup{}
	rq, err := c.newRequest(ctx, "POST", "api/", data)
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

//GetGroup implement Service
func (c *Client) GetGroup(ctx context.Context, groupID int) (map[string]interface{}, error) {
	//https://fabrika-fotoknigi.ru/apiclient.php?cmd=group&args[number]=349141
	if ctx == nil {
		ctx = context.Background()
	}
	data := url.Values{}
	data.Set("appkey", "sp0oULbDnJfk7AjBNtVG")
	data.Set("cmd", "group")
	data.Set("args[number]", strconv.Itoa(groupID))
	rq, err := c.newRequest(ctx, "POST", "api.php/", data)
	if err != nil {
		return nil, err
	}
	var res interface{}
	r, err := c.do(rq, &res)
	if err != nil {
		return nil, err
	}
	if r.StatusCode != http.StatusOK {
		return nil, statusError(r.StatusCode)
	}
	raw, ok := res.(map[string]interface{})
	if !ok {
		return raw, errors.New("Empty or wrong responce")
	}
	res, ok = raw["result"]
	if !ok {
		return raw, errors.New("Empty or wrong responce")
	}
	raw, ok = res.(map[string]interface{})
	if !ok {
		return raw, errors.New("Empty or wrong responce")
	}
	return raw, err
}

//GetBoxes implement Service
func (c *Client) GetBoxes(ctx context.Context, groupID int) (*GroupBoxes, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	//http://fotokniga.by/api/?appkey=91b06dc1105454167c8aad18a96c4572&action=fk:get_group_boxes&id=43314
	data := url.Values{}
	data.Set("action", "fk:get_group_boxes")
	data.Set("id", strconv.Itoa(groupID))
	res := &GroupBoxes{}
	rq, err := c.newRequest(ctx, "POST", "api/", data)
	if err != nil {
		return nil, err
	}
	r, err := c.do(rq, res)
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
	if data.Get("appkey") == "" {
		data.Set("appkey", c.AppKey)
	}
	var reader io.Reader
	if method == "POST" {
		reader = strings.NewReader(data.Encode())
	} else {
		//api bug with reserved char :
		q := data.Encode()
		q = strings.Replace(q, "%3A", ":", -1)
		q = strings.Replace(q, "%5B", "[", -1)
		q = strings.Replace(q, "%5D", "]", -1)
		u.RawQuery = q //data.Encode()
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
		errStr := raw.String()
		err = fmt.Errorf("%s; Response: %s", err.Error(), errStr)

		raw.Reset()
		raw.WriteString(errStr)
		ae := apiError{}
		if e := json.NewDecoder(&raw).Decode(&ae); e == nil && (ae.Code != 0 || ae.Error != "") {
			//intrenal api error
			err = fmt.Errorf("Error: %s; Code: %d; Exception: %s", ae.Error, ae.Code, ae.Exception)
		}
	}
	return resp, err
}

type apiError struct {
	Code      int    `json:"code"`
	Error     string `json:"error"`
	Exception string `json:"exception"`
}
