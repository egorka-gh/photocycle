package job

import (
	"context"
	"fmt"

	"github.com/egorka-gh/photocycle/infrastructure/api"
	"github.com/spf13/viper"
)

func initCheckPrinted(j *baseJob) error {
	//TODO check some
	j.debug = viper.GetBool("efi.debug")
	return nil
}

func checkPrinted(ctx context.Context, j *baseJob) error {

	//get printgroups in state printpost
	pgs, err := j.repo.GetPrintPostedEFI(ctx)
	if err != nil {
		return err
	}
	if len(pgs) == 0 {
		//nothig process
		return nil
	}

	e, err := api.NewEFI()
	if err != nil {
		return err
	}

	err = e.Login(ctx)
	if err != nil {
		return err
	}

	for _, p := range pgs {
		mask := fmt.Sprintf("%s*", p.PrintgroupID)
		itms, err := e.List(ctx, mask)
		if err != nil {
			return err
		}
		if j.debug {
			j.logger.Log("debug", fmt.Sprintf("mask %s,responce %+v", mask, itms))
			continue
		}

		m := make(map[string]bool)
		for _, it := range itms {
			m[it.File] = true
			// if it.Printed {
			// 	m[it.File] = true
			// }
		}
		if len(m) == p.FilesCount {
			//all files printed
			//mark in database
			err = j.repo.SetPrintedEFI(ctx, p.PrintgroupID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
