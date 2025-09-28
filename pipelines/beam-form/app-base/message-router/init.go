package main

import (
	"beamform/internal/cache"
	"os"
	"strconv"
)

var appID int

func init() {
	moduleID, _ := strconv.Atoi(os.Getenv("MODULE_ID"))
	appID = cache.GetAppIDByModuleID(moduleID)
}
