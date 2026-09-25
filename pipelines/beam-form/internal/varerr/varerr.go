// Package varerr 统一判定 scalebox 平台「变量不存在」的错误语义。
//
// 变量不存在是正常语义（尚未分配），调用方应据此走初始化分支，而不是当成致命错误。
//
// 平台升级后，variable 后端由 Redis 改为 gRPC + SQL：server 端查无记录时返回
// codes.Internal，sql.ErrNoRows 只留在错误消息里，错误链中不再含 sql.ErrNoRows
// 实例，原先的 errors.Is(err, sql.ErrNoRows) 判断失效。
package varerr

import (
	"database/sql"
	"errors"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// noRowsMsg 是平台 server 查无记录时返回的错误原文。
const noRowsMsg = "sql: no rows in result set"

// IsNotFound 判断 err 是否表示「变量不存在」，覆盖三种形态：
//   - 错误链中含 sql.ErrNoRows：平台透传，或直连 SQL 后端
//   - gRPC code 为 codes.NotFound：平台按规范返回时
//   - 错误消息含 noRowsMsg：当前 server 以 codes.Internal 返回时
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, sql.ErrNoRows) {
		return true
	}
	if status.Code(err) == codes.NotFound {
		return true
	}
	// status.Convert 对非 gRPC 错误返回 codes.Unknown + err.Error()，
	// 故此处对普通错误同样生效。
	return strings.Contains(status.Convert(err).Message(), noRowsMsg)
}
