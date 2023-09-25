package group

import (
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"
)

var (
	db         *sql.DB
	mapDataset = make(map[string]*DataSet)
)

// DataSet ...
type DataSet struct {
	datasetID string
	// "H" / "V"
	groupType string
	// for tyep "H"
	horizontalSize int
	// for type "V"
	verticalStart int
	verticalSize  int
	// vertical group length
	groupLength int
}

func init() {
	var err error
	logrus.SetReportCaller(true)
	// set database connection
	if db, err = sql.Open("sqlite3", "/tmp/my.db"); err != nil {
		logrus.Fatalln("Unable to open sqlite3 database:", err)
	}
	sqlText := `
		DROP TABLE IF EXISTS t_entity;
		CREATE TABLE t_entity (
			id INTEGER PRIMARY KEY autoincrement,
			name TEXT,
			dataset_id TEXT,
			num_hrzn INTEGER,
			num_vert INTEGER,
			flag TEXT
		);
		CREATE INDEX i_entity_1 ON t_entity(dataset_id);
		CREATE INDEX i_entity_2 ON t_entity(dataset_id,num_hrzn);
	`

	if _, err = db.Exec(sqlText); err != nil {
		logrus.Fatal(err)
	}
}

// AddEntity ...
func (dataset *DataSet) AddEntity(name string, numHorizontal int, numVertical int) bool {
	sqlText := `
		INSERT INTO t_entity(name,dataset_id,num_hrzn,num_vert)
		VALUES($1,$2,$3,$4)
	`
	_, err := db.Exec(sqlText, name, dataset.datasetID, numHorizontal, numVertical)
	if err != nil {
		logrus.Errorf("add entity, err=%v\n", err)
		return false
	}

	var cnt int
	if dataset.groupType == "H" {
		sqlText = `
			SELECT count(*)
			FROM t_entity
			WHERE dataset_id=$1 AND num_vert=$2
		`
		err = db.QueryRow(sqlText, dataset.datasetID, numVertical).Scan(&cnt)
		if err != nil {
			logrus.Errorf("sum entity, err=%v\n", err)
			return false
		}
		if cnt == dataset.horizontalSize {
			return true
		}
	} else {
		sqlText = `
			SELECT count(*)
			FROM t_entity
			WHERE dataset_id=$1 AND (num_vert BETWEEN $2 AND $3)
		`
		length := dataset.groupLength
		n0 := numVertical - (numVertical-dataset.verticalStart)%dataset.groupLength
		n1 := numVertical - (numVertical-dataset.verticalStart)%dataset.groupLength + dataset.groupLength - 1
		if n1 > dataset.verticalStart+dataset.verticalSize-1 {
			n1 = dataset.verticalStart + dataset.verticalSize - 1
			length = n1 - n0 + 1
		}
		err = db.QueryRow(sqlText, dataset.datasetID, n0, n1).Scan(&cnt)
		fmt.Printf("n=%d,low=%d,high=%d,count=%d\n", numVertical, n0, n1, cnt)
		if err != nil {
			logrus.Errorf("sum entity, err=%v\n", err)
			return false
		}
		if cnt == length {
			return true
		}
	}
	return false
}

// NewDataSet ...
func NewDataSet(datasetID string, groupType string, horizontalSize int,
	verticalSize int, verticalStart int, groupLength int) *DataSet {
	dataset := &DataSet{
		datasetID:      datasetID,
		groupType:      groupType,
		horizontalSize: horizontalSize,
		verticalSize:   verticalSize,
		verticalStart:  verticalStart,
		groupLength:    groupLength,
	}
	mapDataset[datasetID] = dataset
	return dataset
}

// GetDataSet ...
func GetDataSet(datasetID string) *DataSet {
	return mapDataset[datasetID]
}
