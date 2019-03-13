package reports

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/gobuffalo/packr"
	"github.com/pkg/browser"
	"github.com/venturasr/ll-analyzer/files"
	"github.com/venturasr/ll-analyzer/tools"
)

var lock = sync.RWMutex{}

type FindIssue func(wg *sync.WaitGroup, line string, data *Dataset, rpt *Report) (kv *KeyValue)

type Publisher interface {
	Publish()
}

type Report struct {
	sync.RWMutex
	ReportType    string
	NumberOfFiles int
	Issues        map[string]interface{}
	Template
	FindIssue
}

type KeyValue struct {
	Key   string
	Value interface{}
}

func (rpt *Report) IssuesLength() (int) {

	rpt.RLock()
	len := len(rpt.Issues)
	rpt.RUnlock()
	return len

}
func (rpt *Report) ReadIssue(key string) (interface{}, bool) {
	rpt.RLock()
	result, ok := rpt.Issues[key]
	rpt.RUnlock()
	return result, ok
}

func (rpt *Report) WriteIssue(key string, valueIssue interface{}) {
	rpt.Lock()
	rpt.Issues[key] = valueIssue
	rpt.Unlock()
}

func (rpt *Report) IterateIssues() (<-chan KeyValue) {

	c := make(chan KeyValue)
	rpt.RLock()
	defer rpt.RUnlock()

	f := func() {
		for key, value := range rpt.Issues {
			c <- KeyValue{key, value}
		}
	}

	go f()

	return c
}

type Dataset struct {
	sync.RWMutex
	Data     string
	LogFile  *files.LogFile
	MapRegEx map[string]*regexp.Regexp
}

func (data *Dataset) ReadData() (string) {

	lock.RLock()
	defer lock.RUnlock()

	value := data.Data
	return value

}

func (data *Dataset) WriteData(someValue string) {

	lock.Lock()
	defer lock.Unlock()

	data.Data = someValue

}

func (data *Dataset) ReadMapRegEx(keyRegex string) (*regexp.Regexp, bool) {

	data.Lock()
	defer data.Unlock()

	regexp, ok := data.MapRegEx[keyRegex]
	return regexp, ok

}

func (data *Dataset) GetLogFile() (*files.LogFile) {

	lock.RLock()
	defer lock.RUnlock()

	return data.LogFile

}

func (data *Dataset) GetMapRegEx() (map[string]*regexp.Regexp) {

	lock.RLock()
	defer lock.RUnlock()

	return data.MapRegEx

}

type Template struct {
	TemplateFile   string
	ReportFileName string
}

func (rpt *Report) Publish() {

	fmt.Println("")
	log.Printf("| [report.go][Publish] report Type -> %s, number of Issues  -> %d", rpt.ReportType, rpt.IssuesLength())
	log.Println()
	log.Println()

	rpt.Lock()
	defer rpt.Unlock()

	pathToTemplates := "../templates"
	homeDir, _ := tools.GetHomeDir()
	if packr.NewBox("../").Has(rpt.TemplateFile) {
		pathToTemplates = "../"
	} else if packr.NewBox("../templates").Has(rpt.TemplateFile) {
		pathToTemplates = "../templates"
	} else if packr.NewBox(homeDir).Has(rpt.TemplateFile) {
		pathToTemplates = homeDir
	} else {
		return
	}
	fmt.Println("\n")
	log.Printf("| [report.go][Publish] path to templates is %s", pathToTemplates)

	box := packr.NewBox(pathToTemplates)

	if rpt.TemplateFile != "" {

		log.Printf("| [report.go][Publish] report type %s, report file name %s, template file %s", rpt.ReportType, rpt.Template.ReportFileName, rpt.Template.TemplateFile)

		funcs := template.FuncMap{"uniqueCollapse": uniqueCollapse}

		tpl, err := template.New(rpt.ReportType).Funcs(funcs).Parse(box.String(rpt.TemplateFile))
		if err != nil {
			log.Println("| [report.go][Publish] error getting template file", err)
		}

		published, err := os.Create(rpt.ReportFileName)
		if err != nil {
			log.Printf("| [report.go][Publish] error creating file %s. %v", rpt.Template, err)
		}

		err = tpl.Execute(published, rpt)
		if err != nil {
			panic(err)
		}

		browser.OpenFile(getPathToReport(rpt.ReportFileName))
	} else {
		fmt.Println("\n")
		log.Printf("| [report.go][Publish] %s has not template file for report publishing", rpt.ReportType)
	}

	return
}

func getPathToReport(reportFileName string) (pathToConfig string) {
	homeDir, err := tools.GetCurrentDir()
	tools.ExitIfError(err)
	pathToConfig = fmt.Sprintf("%s%s%s", homeDir, tools.PathSeparator, reportFileName)

	return pathToConfig
}

func uniqueCollapse(index string) (int) {
	unique := int(100*time.Now().Nanosecond()) + len(index)
	return unique
}
