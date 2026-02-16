package node

import (
	"fmt"
	"os"
	"strings"

	"github.com/kaichao/scalebox/pkg/postgres"
	"github.com/sirupsen/logrus"
)

// n0000 .. n9999
// g00-00 .. g00-23

// Node ...
type Node struct {
	ID     int
	Name   string
	IPAddr string
	Group  string
}

// Nodes ...
var Nodes []*Node

func init() {
	if inTestMode() {
		return
	}
	loadNodeData()
}

func inTestMode() bool {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.") {
			return true
		}
	}
	return false
}
func loadNodeData() {
	nodesRegex := os.Getenv("NODES")
	// regex replace
	nodesRegex = strings.ReplaceAll(nodesRegex, `\|`, `|`)
	if nodesRegex == "" {
		nodesRegex = "n.+"
	}
	clusterName := os.Getenv("CLUSTER")
	sqlText := `
		SELECT id,hostname,ip_addr,group_id
		FROM t_host
		WHERE cluster=$1 AND status='ON'
			AND hostname ~ $2
		ORDER BY 2
	`
	rows, err := postgres.GetDB().Query(sqlText, clusterName, nodesRegex)
	if err != nil {
		logrus.Errorf("query t_cluster error in loadNodeData(), err: %v\n", err)
	}
	defer rows.Close()
	for rows.Next() {
		node := Node{}
		if err := rows.Scan(&node.ID, &node.Name, &node.IPAddr, &node.Group); err != nil {
			logrus.Errorf("Scan hostname, err-info:%v\n", err)
		}
		Nodes = append(Nodes, &node)
	}
	fmt.Printf("regex:%s,nodes:%v\n", nodesRegex, Nodes)
	// 检查 rows 是否有错误
	if err := rows.Err(); err != nil {
		logrus.Errorf("Query Resultset, err-info:%v\n", err)
	}

	if !isFactorOrMultipleOf24(len(Nodes)) {
		logrus.Errorf("node-regex=%s, the number of compute nodes is %d, which is not a multiple or a divisor of 24.\n",
			nodesRegex, len(Nodes))
		os.Exit(1)
	}

}

// isFactorOrMultipleOf24 判断 n 是否是 24 的约数或倍数
func isFactorOrMultipleOf24(n int) bool {
	if n == 0 {
		return false // 0 既不是倍数也不是约数
	}
	return 24%n == 0 || n%24 == 0
}
