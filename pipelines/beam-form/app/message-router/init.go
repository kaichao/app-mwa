package main

import (
	"beamform/internal/cache"
	"beamform/internal/picker"
	"os"
	"strconv"
)

var (
	targetPicker = picker.NewWeightedPicker("target")
	sourcePicker = picker.NewWeightedPicker("source")

	appID int
)

func init() {
	os.Setenv("REDIS_HOST", os.Getenv("GRPC_SERVER"))

	jobID, _ := strconv.Atoi(os.Getenv("JOB_ID"))
	appID = cache.GetAppIDByJobID(jobID)
}
