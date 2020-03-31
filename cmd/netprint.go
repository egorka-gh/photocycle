package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/egorka-gh/photocycle/api"
)

func main() {
	client, err := api.NewClient(http.DefaultClient, "https://fabrika-fotoknigi.ru/api/", "e5ea49c386479f7c30f60e52e8b9107b")
	if err != nil {
		fmt.Println(err)
		return
	}
	//t :=time.Now()

	g, err := client.GetGroups(context.Background(), []int{30}, int64(1574334313))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(g)
}
