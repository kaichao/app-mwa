package varerr_test

import (
	"database/sql"
	"errors"
	"testing"

	"beamform/internal/varerr"

	gopkgerrors "github.com/kaichao/gopkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestIsNotFound(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"sql.ErrNoRows", sql.ErrNoRows, true},
		{"gopkg-wrapped sql.ErrNoRows", gopkgerrors.Wrap(sql.ErrNoRows, "query failed"), true},
		{"grpc NotFound", status.Error(codes.NotFound, "variable not found"), true},
		{
			// 当前 server 的实际行为：codes.Internal + 消息里带 SQL 原文
			"grpc Internal with no-rows message",
			status.Error(codes.Internal, "get variable: sql: no rows in result set"),
			true,
		},
		{
			// 真实调用链：variable.GetValue 的 WrapE 包裹
			"gopkg-wrapped grpc Internal with no-rows message",
			gopkgerrors.WrapE(
				status.Error(codes.Internal, "get variable: sql: no rows in result set"),
				"get variable", "app-id", 1302282040, "var-name", "member-path:preload:k"),
			true,
		},
		{"grpc Internal unrelated", status.Error(codes.Internal, "connection refused"), false},
		{"plain unrelated error", errors.New("dial tcp: connection refused"), false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, varerr.IsNotFound(c.err))
		})
	}
}
