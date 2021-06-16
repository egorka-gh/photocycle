package efi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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
func New() (*EFI, error) {
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
	u := e.baseURL.ResolveReference(&url.URL{Path: "job/"})
	data := url.Values{}
	data.Set("apikey", e.key)
	data.Set("username", e.user)
	data.Set("password", e.pass)
	r := strings.NewReader(data.Encode())
	req, err := http.NewRequestWithContext(ctx, "POST", u.String(), r)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

}

func (e *EFI) List(title string) ([]Item, error) {
	//https://localhost/live/api/v5/jobs?title=1504660-2-blok001.pdf

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
