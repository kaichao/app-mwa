package main

import (
	"beamform/internal/cache"
	"os"
	"strconv"
)

var (
	// targetPicker = picker.NewWeightedPickerByFile("target")
	// sourcePicker = picker.NewWeightedPickerByFile("source")

	appID int
)

func init() {
	os.Setenv("REDIS_HOST", os.Getenv("GRPC_SERVER"))

	jobID, _ := strconv.Atoi(os.Getenv("JOB_ID"))
	appID = cache.GetAppIDByJobID(jobID)
}
