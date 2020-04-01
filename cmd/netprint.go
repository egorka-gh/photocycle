package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/egorka-gh/photocycle/api"
)

func main() {
	client, err := api.NewClient(http.DefaultClient, "https://fabrika-fotoknigi.ru/api/", "e5ea49c386479f7c30f60e52e8b9107b")
	if err != nil {
		fmt.Println(err)
		return
	}
	t := time.Now().Add(-time.Hour * 3)
	fmt.Println(t)

	g, err := client.GetGroups(context.Background(), []int{30, 40}, t.Unix())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(g)
	fmt.Println(len(g))
}
