package cache

import (
	"time"

	"github.com/kaichao/gopkg/dbcache"
	"github.com/kaichao/scalebox/pkg/postgres"
	"github.com/sirupsen/logrus"
)

var (
	appIDCache *dbcache.DBCache[int]
)

func init() {
	appIDCache = dbcache.New[int](
		postgres.GetDB(), // *sql.DB connection
		"SELECT app FROM t_module WHERE id = $1",
		5*time.Minute,  // Cache expiration
		10*time.Minute, // Cleanup interval
		nil,            // Use default loader
	)
}

// GetAppIDByModuleID ...
func GetAppIDByModuleID(moduleID int) int {
	appID, err := appIDCache.Get(moduleID)
	if err != nil {
		logrus.Errorf("In GetAppIDByModuleID(),err-info:%v\n", err)
		return -1
	}
	return appID
}
