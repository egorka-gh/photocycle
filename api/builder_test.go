package api

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/egorka-gh/photocycle/repo"
	_ "github.com/go-sql-driver/mysql"
)

func TestBuildPackage(t *testing.T) {
	rep, err := repo.New("root:3411@tcp(127.0.0.1:3306)/fotocycle_202005?parseTime=true", false)
	if err != nil {
		t.Errorf("Error create repository %q", err.Error())
		return
	}

	fm, err := rep.GetJSONMaps(context.Background())
	if err != nil {
		t.Errorf("Error get JSONMaps from repository %q", err.Error())
		return
	}
	dm, err := rep.GetDeliveryMaps(context.Background())
	if err != nil {
		t.Errorf("Error get GetDeliveryMaps from repository %q", err.Error())
		return
	}

	client, err := NewClient(http.DefaultClient, "https://fabrika-fotoknigi.ru/", "")
	if err != nil {
		t.Errorf("Error create client %q", err.Error())
		return
	}
	g, err := client.GetGroup(nil, 349141)
	if err != nil {
		t.Errorf("Error get group  %q", err.Error())
		return
	}
	fmt.Printf("Group:  %v\n", g)

	builder := &Builder{
		jmap:            fm,
		deliveryMapping: dm,
	}
	p, err := builder.BuildPackage(11, g)
	if err != nil {
		t.Errorf("Error build package  %q", err.Error())
		return
	}
	fmt.Printf("Result:  %v\n", p)

}
