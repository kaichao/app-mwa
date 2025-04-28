package main

import (
	"beamform/internal/pkg/cache"
	"os"
	"strconv"
)

var appID int

func init() {
	jobID, _ := strconv.Atoi(os.Getenv("JOB_ID"))
	appID = cache.GetAppIDByJobID(jobID)
}
