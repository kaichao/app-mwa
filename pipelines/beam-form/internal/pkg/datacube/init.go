package datacube

import (
	"os"

	"github.com/kaichao/scalebox/pkg/misc"
	"github.com/sirupsen/logrus"
	// _ "github.com/jackc/pgx/v5/stdlib"
)

// n0000 .. n9999
// g00n00 .. g00n23 ... g99n00 .. g99n23
var (
	nodeNames []string
	nodeIPs   []string
)

func init() {
	loadNodeNames()
}

func loadNodeNames() {
	nodes := os.Getenv("NODES")
	if nodes == "" {
		nodes = "n.+"
	}
	clusterName := os.Getenv("CLUSTER")
	sqlText := `
		SELECT hostname,ip_addr 
		FROM t_host
		WHERE cluster=$1
			AND hostname ~ $2
		ORDER BY 1
	`
	rows, err := misc.GetDB().Query(sqlText, clusterName, nodes)
	defer rows.Close()
	if err != nil {
		logrus.Errorf("query t_app error: %v\n", err)
	}
	for rows.Next() {
		var hostname, ipAddr string
		if err := rows.Scan(&hostname, &ipAddr); err != nil {
			logrus.Errorf("Scan hostname, err-info:%v\n", err)
		}
		nodeNames = append(nodeNames, hostname)
		nodeIPs = append(nodeIPs, ipAddr)
	}

	// 检查 rows 是否有错误
	if err := rows.Err(); err != nil {
		logrus.Errorf("Query Resultset, err-info:%v\n", err)
	}
}
