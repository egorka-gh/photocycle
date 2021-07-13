package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/spf13/viper"
)

const apiPath string = "live/api/v5/"

//EFI represent EFI api client
type EFI struct {
	baseURL *url.URL
	key     string
	user    string
	pass    string
	client  *http.Client
}

//New init new EFI
func NewEFI() (*EFI, error) {
	us := viper.GetString("efi.url")
	if us == "" {
		return nil, fmt.Errorf("initCheckPrinted error: efi.url not set")
	}
	u, err := url.Parse(us)
	if err != nil {
		return nil, err
	}
	u = u.ResolveReference(&url.URL{Path: apiPath})

	key := viper.GetString("efi.key")
	if key == "" {
		return nil, fmt.Errorf("initCheckPrinted error: efi.key not set")
	}
	user := viper.GetString("efi.user")
	if user == "" {
		return nil, fmt.Errorf("initCheckPrinted error: efi.user not set")
	}
	pass := viper.GetString("efi.pass")
	if pass == "" {
		return nil, fmt.Errorf("initCheckPrinted error: efi.pass not set")
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("got error while creating cookie jar %s", err.Error())
	}

	cl := &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig: &tls.Config{
				// UNSAFE!
				// DON'T USE IN PRODUCTION!
				InsecureSkipVerify: true,
			},
		},
	}
	return &EFI{
		baseURL: u,
		key:     key,
		client:  cl,
		user:    user,
		pass:    pass,
	}, nil
}

func (e *EFI) do(req *http.Request, v interface{}) (*http.Response, error) {

	ee := efiError{}
	resp, err := do(e.client, req, v, &ee)
	if ee.Code != 0 {
		//intrenal api error
		err = fmt.Errorf("сode: %d; message: %s; details: %v ", ee.Code, ee.Message, ee.Errors)
	}
	return resp, err
}

func (e *EFI) Login(ctx context.Context) error {
	//https://localhost/live/api/v5/login/
	u := e.baseURL.ResolveReference(&url.URL{Path: "login/"})
	data := url.Values{}
	data.Set("apikey", e.key)
	data.Set("username", e.user)
	data.Set("password", e.pass)

	req, err := newRequest(ctx, "POST", u, data, false)
	if err != nil {
		return err
	}
	var m map[string]interface{}
	r, err := e.do(req, &m)
	if err != nil {
		return err
	}
	if r.StatusCode != http.StatusOK {
		return statusError(r.StatusCode)
	}
	return nil
}

func (e *EFI) List(ctx context.Context, title string) ([]Item, error) {
	//https://localhost/live/api/v5/jobs?title=1504660-2-blok001.pdf
	u := e.baseURL.ResolveReference(&url.URL{Path: "jobs/"})
	data := url.Values{}
	data.Set("title", title)
	data.Set("print status", "OK")
	req, err := newRequest(ctx, "GET", u, data, false)
	if err != nil {
		return nil, err
	}
	var res listDTO
	r, err := e.do(req, &res)
	if err != nil {
		return nil, err
	}
	if r.StatusCode != http.StatusOK {
		return nil, statusError(r.StatusCode)
	}
	return res.Data.Items, nil

}

//Item represent EFI print job
type Item struct {
	File    string `json:"title"`
	Printed yesNo  `json:"print status"`
}

// "data": {
// 	"totalItems": 4,
// 	"kind": "FieryCutSheetJobs",
// 	"items": [

type listData struct {
	TotalItems int    `json:"totalItems"`
	Items      []Item `json:"items"`
}
type listDTO struct {
	Data listData `json:"data"`
}

type yesNo bool

/*
"title": "1504660-2-blok001.pdf",
"username": "Fiery Hot Folders",
"status": "done printing",
"state": "completed",
"print status": "OK",
*/

//UnmarshalJSON  Unmarshal custom date format
func (yn *yesNo) UnmarshalJSON(b []byte) error {
	//TODO not work, so unused
	//unmarshal to string??
	y := yesNo(string(b) == "OK")
	yn = &y
	return nil
}

type efiError struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Errors  []efiError `json:"errors"`
}
