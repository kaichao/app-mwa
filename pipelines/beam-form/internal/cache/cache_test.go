package cache_test

import (
	"beamform/internal/cache"
	"fmt"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func init() {
	os.Setenv("PGHOST", "10.0.6.100")
}
func TestGetAppIDByModuleID(t *testing.T) {
	appID := cache.GetAppIDByModuleID(5)
	fmt.Println("app-id:", appID)
}
