package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	scalebox "github.com/kaichao/scalebox/golang/misc"
)

// DataSet ...
type DataSet struct {
	// prefix ':' type ':' sub-id
	DatasetID string

	// for type "H", x-coord
	HorizontalWidth int

	// for type "V", y-coord
	VerticalStart  int
	VerticalHeight int
}

func parseDataSet(t string) *DataSet {
	var ds DataSet
	if err := json.Unmarshal([]byte(t), &ds); err != nil {
		// skip non-json format error
		if !strings.HasPrefix(err.Error(), "invalid character") {
			fmt.Printf("error parsing, err-info:%v\n", err)
		}
		// non-dataset definition
		return nil
	}
	return &ds
}

func initDataGrouping(dataset *DataSet) {

	fmtDatDataSet := ` {
		"datasetID":"dat:%s",
		"keyGroupRegex":"^([0-9]+)/[0-9]+_([0-9]+)_ch([0-9]{3}).dat$",
		"keyGroupIndex":"1,3,2",
		"sinkJob":"message-router-main",

		"groupType":"V",
		"verticalStart": %d,
		"verticalHeight": %d,
		"groupSize": %d
	}
	`
	numPerGroup := 30
	// Remove space characters
	format := regexp.MustCompile("\\s+").ReplaceAllString(fmtDatDataSet, "")
	s := fmt.Sprintf(format, dataset.DatasetID, dataset.VerticalStart, dataset.VerticalHeight, numPerGroup)
	scalebox.AppendToFile("/work/messages.txt", "data-grouping-main,"+s)

	fmtFitsDataSet := ` {
		"datasetID":"fits:%s",
		"keyGroupRegex":"^([0-9]+)/([0-9]+_[0-9]+/[0-9]+)/ch([0-9]{3}).fits$",
		"keyGroupIndex":"1,3,2",
		"sinkJob":"message-router-main",

		"groupType":"H",
		"horizontalWidth":%d
	}
	`
	// Remove space characters
	format = regexp.MustCompile("\\s+").ReplaceAllString(fmtFitsDataSet, "")
	s = fmt.Sprintf(format, dataset.DatasetID, dataset.HorizontalWidth)
	scalebox.AppendToFile("/work/messages.txt", "data-grouping-main,"+s)
}

func fromDataGroupingMain(message string, headers map[string]string) int {
	// for dat file
	//  input: 1257010784/1257010784_1257010790_ch132.dat,...,1257010784/1257010784_1257010799_ch132.dat
	//	output: 1257010784/1257010986_1257011185/132/00001_00003

	// for fits file
	//  input: 1257010784/1257010786_1257010795/00001/ch109.fits,...,1257010784/1257010786_1257010795/00001/ch132.fits
	//	output: 1257010784/1257010786_1257010815/00001

	if strings.HasSuffix(message, "dat") {
		ms := strings.Split(message, ",")
		// first plus last
		str := ms[0] + "," + ms[len(ms)-1]
		fmt.Println(str)

		datPattern := "([0-9]+)/[0-9]+_([0-9]+)_ch([0-9]{3}).dat"
		format := "^%s,%s$"
		reDat := regexp.MustCompile(fmt.Sprintf(format, datPattern, datPattern))
		ss := reDat.FindStringSubmatch(str)
		if ss == nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Not valid format, message:%s\n", message)
			return 99
		}
		ds := ss[1]
		start := ss[2]
		ch := ss[3]
		end := ss[5]
		m := fmt.Sprintf("%s/%s_%s/%s/%s", ds, start, end, ch, currentPointings)

		sinkJob := "beam-maker"
		n, _ := strconv.Atoi(ch)
		toHost := hosts[(n-109)%numNodesPerGroup]
		cmdTxt := fmt.Sprintf("scalebox task add --sink-job %s --to-ip %s %s", sinkJob, toHost, m)
		// scalebox.ExecShellCommand(cmdTxt)
		code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdTxt, 10)
		fmt.Printf("stdout for task-add:\n%s\n", stdout)
		fmt.Fprintf(os.Stderr, "stderr for task-add:\n%s\n", stderr)
		return code

		// scalebox.AppendToFile("/work/messages.txt", "beam-maker,"+m)
	} else if strings.HasSuffix(message, "fits") {
		str := strings.Split(message, ",")[0]
		fits1chPattern := "^([0-9]+/[0-9]+_[0-9]+/([0-9]{5}))/ch.+fits$"
		reFits1ch := regexp.MustCompile(fits1chPattern)
		ss := reFits1ch.FindStringSubmatch(str)
		if ss == nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Not valid format, message:%s\n", message)
			return 99
		}
		m := ss[1]
		pointing := ss[2]
		n, _ := strconv.Atoi(pointing)
		toHost := hosts[(n-1)%numNodesPerGroup]
		sinkJob := "fits-merger"
		cmdTxt := fmt.Sprintf("scalebox task add --sink-job %s --to-ip %s %s", sinkJob, toHost, m)
		code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdTxt, 10)
		fmt.Printf("stdout for task-add:\n%s\n", stdout)
		fmt.Fprintf(os.Stderr, "stderr for task-add:\n%s\n", stderr)
		return code
		// scalebox.AppendToFile("/work/messages.txt", "fits-merger,"+m)
	} else {
		fmt.Fprintf(os.Stderr, "[ERROR] Not valid format, message:%s\n", message)
		return 99
	}
}
