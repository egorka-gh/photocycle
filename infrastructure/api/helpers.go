package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func newRequest(ctx context.Context, method string, endpoint *url.URL, data url.Values, unescape bool) (*http.Request, error) {
	if data == nil {
		data = url.Values{}
	}
	var reader io.Reader
	if method == "POST" {
		reader = strings.NewReader(data.Encode())
	} else {
		//api bug with reserved char :
		q := data.Encode()
		if unescape {
			q = strings.Replace(q, "%3A", ":", -1)
			q = strings.Replace(q, "%5B", "[", -1)
			q = strings.Replace(q, "%5D", "]", -1)
		}
		endpoint.RawQuery = q //data.Encode()
	}
	//fmt.Println(data.Encode())
	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), reader)
	if err != nil {
		return nil, err
	}

	if method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req.Header.Set("Accept", "application/json")

	return req, nil
}

type transportError error

func do(httpClient *http.Client, req *http.Request, value, errorvalue interface{}) (*http.Response, error) {
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, transportError(err)
	}
	defer resp.Body.Close()
	var raw bytes.Buffer
	tee := io.TeeReader(resp.Body, &raw)
	err = json.NewDecoder(tee).Decode(value)
	errStr := raw.String()
	//not a json?
	if err != nil {
		err = fmt.Errorf("%s; Response: %s", err.Error(), errStr)
		return nil, err
	}

	//can be error response
	if errorvalue != nil {
		raw.Reset()
		raw.WriteString(errStr)
		json.NewDecoder(&raw).Decode(errorvalue)
	}

	return resp, err
}

func statusError(code int) error {
	return fmt.Errorf("wrong http status %d. %s", code, http.StatusText(code))
}
