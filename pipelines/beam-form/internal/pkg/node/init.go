package node

import (
	"fmt"
	"os"
	"strings"

	"github.com/kaichao/scalebox/pkg/postgres"
	"github.com/sirupsen/logrus"
)

// n0000 .. n9999
// g00n00 .. g00n23 ... g99n00 .. g99n23
var (
	NodeNames []string
	nodeIPs   []string
)

func init() {
	if IsInTest() {
		return
	}
	loadNodeNames()
}

func IsInTest() bool {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.") {
			return true
		}
	}
	return false
}
func loadNodeNames() {
	nodesRegex := os.Getenv("NODES")
	// regex replace
	nodesRegex = strings.ReplaceAll(nodesRegex, `\|`, `|`)
	if nodesRegex == "" {
		nodesRegex = "n.+"
	}
	clusterName := os.Getenv("CLUSTER")
	sqlText := `
		SELECT hostname,ip_addr 
		FROM t_host
		WHERE cluster=$1
			AND hostname ~ $2
		ORDER BY 1
	`
	rows, err := postgres.GetDB().Query(sqlText, clusterName, nodesRegex)
	defer rows.Close()
	if err != nil {
		logrus.Errorf("query t_app error: %v\n", err)
	}
	for rows.Next() {
		var hostname, ipAddr string
		if err := rows.Scan(&hostname, &ipAddr); err != nil {
			logrus.Errorf("Scan hostname, err-info:%v\n", err)
		}
		NodeNames = append(NodeNames, hostname)
		nodeIPs = append(nodeIPs, ipAddr)
	}

	fmt.Printf("regex:%s,nodes:%v\n", nodesRegex, NodeNames)
	// 检查 rows 是否有错误
	if err := rows.Err(); err != nil {
		logrus.Errorf("Query Resultset, err-info:%v\n", err)
	}
}
