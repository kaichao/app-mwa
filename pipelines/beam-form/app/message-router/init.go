package main

import (
	"os"
)

func init() {
	os.Setenv("REDIS_HOST", os.Getenv("GRPC_SERVER"))
}
