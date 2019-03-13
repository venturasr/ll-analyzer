package reports

import (
	"bytes"
	"regexp"
	"strings"
	"sync"

	"github.com/venturasr/ll-analyzer/tools"
)

type ErrorIssue struct {
	ConsoleIssue
	Line    string
	Message string
	Thread  string
	Job     string
}

type ExceptionIssue struct {
	ConsoleIssue
	Line      string
	Message   string
	Exception string
	Causes    []*CausedByException
}

type CausedByException struct {
	CausedBy string
	Message  string
}

type ConsoleIssue struct {
	IssueType      string
	Timestamp      string
	Occurrences    int
	IssueFileNames map[string]string
}

func FindIssueConsole(wg *sync.WaitGroup, line string, d *Dataset, rpt *Report) (kv *KeyValue) {

	isAndIssue, issueField, mapValues := isConsoleIssue(line, d)
	if !isAndIssue {
		return
	}

	switch issueField {

	case "causedby":

		return addCausedByToException(mapValues, line, rpt)

	case "errorLine":

		return addErrorIssue(mapValues, rpt, d, line)

	case "exceptionLine":

		return addExceptionIssue(mapValues, rpt, d, line)

	default:

	}

	return

}

func addErrorIssue(mapValues map[string]string, rpt *Report, dataset *Dataset, line string) (kv *KeyValue) {

	var errorIssue ErrorIssue

	var buffer bytes.Buffer
	buffer.WriteString("ERROR_")
	buffer.WriteString(mapValues["job"])
	buffer.WriteString("_")
	buffer.WriteString(mapValues["message"])
	keyIssue := buffer.String()

	rpt.Lock()
	defer rpt.Unlock()
	issue, ok := rpt.Issues[keyIssue]
	if ok {

		errorIssue = issue.(ErrorIssue)

	} else {

		errorIssue = ErrorIssue{
			Line:         line,
			Thread:       mapValues["thread"],
			Job:          mapValues["job"],
			Message:      mapValues["message"],
			ConsoleIssue: ConsoleIssue{IssueType: "error", Timestamp: mapValues["timestamp"], IssueFileNames: make(map[string]string)},
		}
	}

	tools.WriteMap(errorIssue.IssueFileNames, dataset.GetLogFile().Name, dataset.GetLogFile().Name)
	errorIssue.Occurrences++
	kv = &KeyValue{}
	kv.Key = keyIssue
	kv.Value = errorIssue
	rpt.Issues[keyIssue] = errorIssue
	return kv

}

func addExceptionIssue(mapValues map[string]string, rpt *Report, dataset *Dataset, line string) (kv *KeyValue) {

	var exceptionIssue ExceptionIssue
	var buffer bytes.Buffer
	buffer.WriteString("EXCEP_")
	buffer.WriteString(mapValues["exception"])
	keyIssue := buffer.String()

	rpt.Lock()
	defer rpt.Unlock()
	issue, ok := rpt.Issues[keyIssue]
	if ok {

		exceptionIssue = issue.(ExceptionIssue)

	} else {

		exceptionIssue = ExceptionIssue{
			Line:         line,
			Exception:    mapValues["exception"],
			Message:      mapValues["message"],
			Causes:       make([]*CausedByException, 1, 2),
			ConsoleIssue: ConsoleIssue{IssueType: "exception", Timestamp: mapValues["timestamp"], IssueFileNames: make(map[string]string)},
		}

	}

	tools.WriteMap(exceptionIssue.IssueFileNames, dataset.GetLogFile().Name, dataset.GetLogFile().Name)
	exceptionIssue.Occurrences++
	kv = &KeyValue{}
	kv.Key = keyIssue
	kv.Value = exceptionIssue
	rpt.Issues[keyIssue] = exceptionIssue
	return kv

}

func addCausedByToException(mapValues map[string]string, line string, rpt *Report) (kv *KeyValue) {
	causedByExc := getCausedBy(mapValues, line)

	if causedByExc != nil {
		ts := mapValues["timestamp"]
		var exceptionIssue ExceptionIssue
		rpt.Lock()
		defer rpt.Unlock()
		for keyIssue, value := range rpt.Issues {

			if strings.Contains(keyIssue, "EXCEP_") {
				exceptionIssue = value.(ExceptionIssue)
				if exceptionIssue.Timestamp == ts && len(exceptionIssue.Causes) > 0 {
					exceptionIssue.Causes[len(exceptionIssue.Causes)-1] = causedByExc
					kv = &KeyValue{}
					kv.Key = keyIssue
					kv.Value = exceptionIssue
					rpt.Issues[keyIssue] = exceptionIssue
					return kv
				}
			}

		}
	}
	return nil
}

func getCausedBy(mapValues map[string]string, line string) (causedByExc *CausedByException) {
	var causedby string
	var message string

	if len(mapValues["causedby"]) != 0 {

		causedby = mapValues["causedby"]
		splitted := strings.SplitAfterN(line, "Exception", 2)

		if len(splitted) == 2 {
			message = strings.TrimLeft(strings.TrimPrefix(splitted[1], ":"), " ")
		}

		causedByExc = &CausedByException{CausedBy: causedby, Message: message}

	}

	return
}

func isConsoleIssue(line string, dataset *Dataset) (isAndIssue bool, issueField string, mapValues map[string]string) {
	isAndIssue = false

	if len(line) == 0 {
		isAndIssue = false
		return
	}

	causedby := matchField(line, dataset, "causedby")
	errorLine := matchField(line, dataset, "errorLine")
	exceptionLine := matchField(line, dataset, "exceptionLine")

	if len(causedby) == 0 && len(errorLine) == 0 && len(exceptionLine) == 0 {
		return false, "", nil
	}

	for issueField, regex := range dataset.GetMapRegEx() {
		switch issueField {

		case "causedby":

			if match := regex.FindString(line); match != "" {
				if regex, ok := dataset.ReadMapRegEx("causedbyLine"); ok {
					mapValues = matchAllField(regex, line)
					isAndIssue = true
					//log.Printf("| [console.go][isConsoleIssue - causedby]starting: %s\ntimestamp: %s\ncausedby: %s\nmessage: %s", mapValues["starting"], mapValues["timestamp"], mapValues["causedby"], mapValues["message"])
					return isAndIssue, issueField, mapValues
				}
			}

		case "causedbyLine":

			if rgcausedby, ok := dataset.ReadMapRegEx("causedby"); ok {
				if match := rgcausedby.FindString(line); match != "" {
					mapValues = matchAllField(regex, line)
					isAndIssue = true
					//log.Printf("| [console.go][isConsoleIssue - causedbyLine]starting: %s\ntimestamp: %s\ncausedby: %s\nmessage: %s", mapValues["starting"], mapValues["timestamp"], mapValues["causedby"], mapValues["message"])
					return isAndIssue, issueField, mapValues

				}

			}

		case "errorLine":

			if match := regex.FindString(line); match != "" {
				mapValues = matchAllField(regex, line)
				isAndIssue = true
				//log.Printf("| [console.go][isConsoleIssue - errorLine]starting: %s\ntimestamp: %s\nthread: %s\njob: %s\nmessage: %s", mapValues["starting"], mapValues["timestamp"], mapValues["thread"], mapValues["job"], mapValues["message"])
				return isAndIssue, issueField, mapValues
			}

		case "exceptionLine":

			if rgcausedby, ok := dataset.ReadMapRegEx("causedby"); ok {

				if match := rgcausedby.FindString(line); match == "" {

					if match := regex.FindString(line); match != "" {

						mapValues = matchAllField(regex, line)
						isAndIssue = true
						//log.Printf("| [console.go][isConsoleIssue - exceptionLine] starting: %s\ntimestamp: %s\nexception: %s\nmessage: %s", mapValues["starting"], mapValues["timestamp"], mapValues["exception"], mapValues["message"])
						return isAndIssue, issueField, mapValues

					}
				}
			}

		}

	}

	return false, "", nil
}

func matchField(line string, dataset *Dataset, field string) string {

	if match := tools.MatchStringInLine(line, dataset.MapRegEx[field], ""); match != "" {
		return match
	}

	return ""
}

func matchAllField(r *regexp.Regexp, str string) (map[string]string) {

	match := r.FindStringSubmatch(str)
	mapValues := make(map[string]string)

	for i, name := range r.SubexpNames() {
		if i != 0 && len(name) > 0 && i < len(match)-1 {
			mapValues[name] = match[i]
		}

		//else {
		//	fmt.Println()
		//	fmt.Printf("name, %s index %d\n", name, i)
		//	fmt.Printf("line,  %s\n", str)
		//	fmt.Printf("regex,  %s\n", r.String())
		//	fmt.Printf("match[i],  %s\n", match[i])
		//}
	}

	return mapValues
}
