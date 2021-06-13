package job

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
)

var efiURL string
var efiKey string

func initCheckPrinted(j *baseJob) error {
	efiURL = viper.GetString("efi.url")
	if efiURL == "" {
		return fmt.Errorf("initCheckPrinted error: efi.url not set")
	}
	efiKey = viper.GetString("efi.key")
	if efiKey == "" {
		return fmt.Errorf("initCheckPrinted error: efi.key not set")
	}
	//TODO check some else

	return nil
}

func checkPrinted(ctx context.Context, j *baseJob) error {

	pgs, err := j.repo.GetPrintPostedEFI(ctx)

	return nil
}
