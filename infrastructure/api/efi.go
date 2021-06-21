package api

import (
	"context"
	"fmt"
	"net/http"
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

	return &EFI{
		baseURL: u,
		key:     key,
		client:  &http.Client{Timeout: time.Second * 40},
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

/*
ok responce
{
    "time": "2021-06-16T23:01:40+03:00",
    "data": {
        "kind": "FieryCutSheetAuthentication",
        "item": {
            "authenticated": true,
            "fiery": true
        },
        "_links": [
            {
                "rel": "self",
                "href": "https://localhost/live/api/v5/login/"
            }
        ]
    }
}
err responce
{
    "error": {
        "code": 401,
        "message": "Unauthorized",
        "errors": [
            {
                "domain": "Harmony",
                "reason": "Harmony function returned non-zero code",
                "code": -2,
                "message": "Invalid username/password, or the permission set for this service do not allow you access to the service."
            }
        ]
    }
}
*/

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
	var res interface{}
	r, err := e.do(req, res)
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
	req, err := newRequest(ctx, "GET", u, data, false)
	if err != nil {
		return nil, err
	}
	var res []Item
	r, err := e.do(req, res)
	if err != nil {
		return nil, err
	}
	if r.StatusCode != http.StatusOK {
		return nil, statusError(r.StatusCode)
	}
	return res, nil

}

type yesNo bool

/*
"title": "1504660-2-blok001.pdf",
"username": "Fiery Hot Folders",
"status": "done printing",
"state": "completed",
"print status": "OK",
*/

//Item represent EFI print job
type Item struct {
	File    string `json:"title"`
	Printed yesNo  `json:"print status"`
}

//UnmarshalJSON  Unmarshal custom date format
func (yn *yesNo) UnmarshalJSON(b []byte) error {
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
