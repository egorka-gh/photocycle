package api

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/egorka-gh/photocycle"
	"github.com/egorka-gh/photocycle/repo"
	_ "github.com/go-sql-driver/mysql"
)

func TestBuildPackage(t *testing.T) {
	rep, err := repo.New("root:3411@tcp(127.0.0.1:3306)/fotocycle_202005?parseTime=true", false)
	if err != nil {
		t.Errorf("Error create repository %q", err.Error())
		return
	}
	//http://fotokniga.by/apiclient.php?cmd=group&args[number]=44059
	client, err := NewClient(http.DefaultClient, "http://fotokniga.by/", "")
	if err != nil {
		t.Errorf("Error create client %q", err.Error())
		return
	}
	g, err := client.GetGroup(nil, 44059)
	if err != nil {
		t.Errorf("Error get group  %q", err.Error())
		return
	}
	fmt.Printf("Group:  %v\n", g)

	builder, err := CreateBuilder(rep)
	if err != nil {
		t.Errorf("Error create builder  %q", err.Error())
		return
	}
	p, err := builder.BuildPackage(8, g)
	if err != nil {
		t.Errorf("Error build package  %q", err.Error())
		return
	}
	fmt.Printf("Result:  %v\n", p)
	ps := []*photocycle.Package{p}
	rep.PackageAddWithBoxes(context.Background(), ps)

}
