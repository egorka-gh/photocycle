package api

import (
	"fmt"
	"net/http"
	"testing"
)

func TestGroupBoxes(t *testing.T) {
	//http://fotokniga.by/api/?appkey=91b06dc1105454167c8aad18a96c4572&action=fk:get_group_boxes&id=43314
	client, err := NewClient(http.DefaultClient, "http://fotokniga.by/", "91b06dc1105454167c8aad18a96c4572")
	if err != nil {
		t.Errorf("Error create client %q", err.Error())
		return
	}
	b, err := client.GetBoxes(nil, 43314)
	if err != nil {
		t.Errorf("Error %q", err.Error())
		return
	}
	fmt.Printf("Boxes:  %v\n", b)
	//wrong url
	client, err = NewClient(http.DefaultClient, "http://fotoknigGa.by/", "91b06dc1105454167c8aad18a96c4572")
	if err != nil {
		t.Errorf("Error create client %q", err.Error())
		return
	}
	_, err = client.GetBoxes(nil, 43314)
	if err == nil {
		t.Error("Expect error (wrong url) but got nil")
		return
	}
	fmt.Printf("Error:  %v\n", err)
	//wrong url
	client, err = NewClient(http.DefaultClient, "http://fotoknigGa.by/", "91b06dc1105454167c8aad18a96c4572")
	if err != nil {
		t.Errorf("Error create client %q", err.Error())
		return
	}
	_, err = client.GetBoxes(nil, 43314)
	if err == nil {
		t.Error("Expect error (wrong url) but got nil")
		return
	}
	fmt.Printf("Error:  %v\n", err)
	//wrong key
	client, err = NewClient(http.DefaultClient, "http://fotokniga.by/", "wrong_app_key")
	if err != nil {
		t.Errorf("Error create client %q", err.Error())
		return
	}
	_, err = client.GetBoxes(nil, 43314)
	if err == nil {
		t.Error("Expect error (wrong app key) but got nil")
		return
	}
	fmt.Printf("Error:  %v\n", err)
	//wrong id
	client, err = NewClient(http.DefaultClient, "http://fotokniga.by/", "91b06dc1105454167c8aad18a96c4572")
	if err != nil {
		t.Errorf("Error create client %q", err.Error())
		return
	}
	_, err = client.GetBoxes(nil, -11111)
	if err == nil {
		t.Error("Expect error (wrong group id) but got nil")
		return
	}
	fmt.Printf("Error:  %v\n", err)
}
