package job

import (
	"context"
	"net/url"
)

func initCheckPrinted(j *baseJob) error {
	//TODO check some
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

	///live/api/v5/jobs?title=1504660-2-blok001.pdf
	u := e.eURL.ResolveReference(&url.URL{Path: "live/api/v5/jobs"})
	data = url.Values{}
	for _, pg := range pgs {

	}

	return nil
}
