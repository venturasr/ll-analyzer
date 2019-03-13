package reports

import (
	"fmt"
	"log"
	"sync"

	"github.com/pkg/errors"
	"github.com/venturasr/ll-analyzer/config"
)

type Registry struct {
	sync.RWMutex
	Reports map[string]*Report
}

type Register interface {
	register(confLogs config.Logs) error
}

func NewRegistry(c *config.Config) (*Registry, error) {

	if c == nil {
		return nil, errors.New("there is no configuration, it is nil")
	}

	r := new(Registry)
	r.Reports = make(map[string]*Report)

	if len(c.Logs) > 0 {
		for _, confLogs := range c.Logs {
			r.register(confLogs)
		}
	}

	return r, nil
}

//
func (reg *Registry) register(confLogs config.Logs) error {

	mapR := reg.Reports
	if _, ok := mapR[confLogs.LogType]; !ok {
		//rpt := new(Report)
		rpt := &Report{}
		rpt.ReportType = confLogs.LogType
		rpt.Issues = make(map[string]interface{}) //DELETE IF USE Issues sync.Map
		rpt.Template = Template{TemplateFile: confLogs.TemplateFile, ReportFileName: confLogs.ReportFileName}

		if err := registerReportFunctions(rpt); err != nil {
			return errors.Wrap(err, fmt.Sprintf("report % cannot be registered", rpt.ReportType))
		}

		mapR[rpt.ReportType] = rpt
		log.Printf("| [register.go][register] report %s registered", rpt.ReportType)
	}

	return nil
}

func registerReportFunctions(rpt *Report) (error) {

	errMsg := fmt.Sprintf("there is not 'FindIssue()' function for this report type %s", rpt.ReportType)

	switch rpt.ReportType {

	case "jdbc":
		rpt.FindIssue = FindIssueJdbs

	case "console":
		rpt.FindIssue = FindIssueConsole

	case "thread-dump":
		return errors.New(errMsg)

	default:
		return errors.New(errMsg)
	}

	return nil
}

func (reg *Registry) ReadMapReports(reporType string) (*Report, bool) {

	reg.Lock()
	defer reg.Unlock()

	rpt, ok := reg.Reports[reporType]
	return rpt, ok
}

func (reg *Registry) IterateReports() (<-chan *Report) {

	c := make(chan *Report)

	f := func() {
		reg.Lock()
		defer reg.Unlock()

		for _, rpt := range reg.Reports {
			c <- rpt
		}
		close(c)
	}
	go f()

	return c
}
