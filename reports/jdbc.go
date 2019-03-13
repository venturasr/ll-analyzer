package reports

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/venturasr/ll-analyzer/files"
	"github.com/venturasr/ll-analyzer/tools"
)

type JDBCIssue struct {
	Statement         string
	Occurrences       int
	TotalTime         int
	HigherTimeMillis  int
	AverageTimeMillis int
	IssueFileNames    map[string]string
	Sqls              map[string]SqlIssue
}

type SqlIssue struct {
	Sql   string
	Trace []string
}

func FindIssueJdbs(wg *sync.WaitGroup, line string, d *Dataset, rpt *Report) (kv *KeyValue) {
	lf := d.LogFile
	mapRegEx := d.MapRegEx

	isAndIssue := isJDBCIssue(mapRegEx, line, lf.Separator)
	if !isAndIssue {
		return nil
	}

	mapStrValues := make(map[string]string)
	for k, v := range mapRegEx {
		if matchedField := tools.MatchStringInLine(line, v, lf.Separator); matchedField != "" {
			mapStrValues[k] = matchedField
		}
	}

	kv = addJDBCIssue(mapStrValues, lf, rpt)

	return kv

}

func isJDBCIssue(mapRegEx map[string]*regexp.Regexp, data string, separator string) (isAndIssue bool) {
	isAndIssue = true

	if len(data) == 0 {
		isAndIssue = false
		return
	}

	if matchedField := tools.MatchStringInLine(data, mapRegEx["executionTime"], separator); matchedField == "" {
		isAndIssue = false
		return
	}

	if matchedField := tools.MatchStringInLine(data, mapRegEx["statement"], separator); matchedField == "" {
		isAndIssue = false
		return
	}
	return

}

func addJDBCIssue(mapStrValues map[string]string, logFile *files.LogFile, rpt *Report) (kv *KeyValue) {
	var jdbcIssue JDBCIssue
	statement, _ := tools.ReadMap(mapStrValues, "statement")

	rpt.Lock()
	defer rpt.Unlock()
	issue, ok := rpt.Issues[statement]

	if ok {

		jdbcIssue = issue.(JDBCIssue)

	} else {

		jdbcIssue = JDBCIssue{}
		jdbcIssue.IssueFileNames = make(map[string]string)
		jdbcIssue.Sqls = make(map[string]SqlIssue)

	}

	trace, _ := tools.ReadMap(mapStrValues, "trace")
	sql, _ := tools.ReadMap(mapStrValues, "sql")

	if len(sql) == 0 {
		return
	}

	var classes []string
	if trace != "" {
		sql = strings.Split(sql, trace)[0]
		sql = formatSql(sql)
		classes = formatTrace(trace)
	}

	sqlIssue := SqlIssue{Sql: sql, Trace: classes}
	if _, ok := readSqlIssueMap(jdbcIssue.Sqls, statement); !ok {
		writeSqlIssueMap(jdbcIssue.Sqls, statement, sqlIssue)
	}

	tools.WriteMap(jdbcIssue.IssueFileNames, logFile.Name, logFile.Name)
	jdbcIssue.Statement = statement
	jdbcIssue.Occurrences++

	executionTime, _ := tools.ReadMap(mapStrValues, "executionTime")
	executionTimeIssue := formatExecutionTime(executionTime)
	jdbcIssue.TotalTime += executionTimeIssue
	if executionTimeIssue > jdbcIssue.HigherTimeMillis {
		jdbcIssue.HigherTimeMillis = executionTimeIssue
	}
	jdbcIssue.AverageTimeMillis = jdbcIssue.TotalTime / jdbcIssue.Occurrences

	kv = &KeyValue{}
	kv.Key = statement
	kv.Value = jdbcIssue
	rpt.Issues[statement] = jdbcIssue
	return

}

func formatTrace(trace string) ([]string) {

	splitted := strings.Split(trace, ":")
	formatted := make([]string, 0, 100)
	temp := make([]string, 2)
	var y int
	if len(splitted) > 0 {

		for x := 0; x < len(splitted); {

			y = x + 1

			if x+1 < len(splitted) {
				temp[0] = splitted[x]
				temp[1] = splitted[y]
				formatted = append(formatted, strings.Join(temp, ":"))
			}

			x += 2
		}
	}

	return formatted
}

func formatSql(sql string) (string) {
	replacewith := make(map[string]string)
	replacewith["("] = "(\n"
	replacewith[")"] = ")\n"
	replacewith[","] = ",\n"

	return replaceAll(sql, replacewith)
}

func replaceAll(original string, replacewith map[string]string) (replaced string) {

	if len(original) == 0 || len(replacewith) == 0 {
		return
	}

	replaced = original
	for old, new := range replacewith {
		replaced = strings.Replace(replaced, old, new, -1)
	}

	return
}

func formatExecutionTime(exeTime string) int {

	etInt, err := strconv.ParseInt(strings.TrimSpace(strings.TrimRight(exeTime, "ms")), 10, 0)
	if err != nil {
		fmt.Println("err: ", err)
		return 0
	}

	return int(etInt)
}

func writeSqlIssueMap(sqlMap map[string]SqlIssue, sqlKey string, sqlIssue SqlIssue) {
	lock.Lock()
	defer lock.Unlock()
	sqlMap[sqlKey] = sqlIssue
}

func readSqlIssueMap(sqlMap map[string]SqlIssue, sqlKey string) (sqlIssue SqlIssue, ok bool) {
	lock.Lock()
	defer lock.Unlock()

	sqlIssue, ok = sqlMap[sqlKey]
	return
}
