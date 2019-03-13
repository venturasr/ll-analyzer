package analyzer

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/venturasr/ll-analyzer/config"
	"github.com/venturasr/ll-analyzer/files"
	"github.com/venturasr/ll-analyzer/reports"
)

type Engineer struct {
	*config.Config
	*reports.Registry
}

type Analyzer interface {
	Analyze(lf *files.LogFile) (int, error)
}

func (e *Engineer) Analyze(lf *files.LogFile) (error) {

	fmt.Println()
	log.Printf("| [engineer.go][Analyze]a file %s found in directory", lf.Name)

	err := lf.TypeOfLog(lf.Path, e.Config)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("there is no pattern to match with this file %s. No log type found", lf.Name))
	}

	rpt, data, err := e.setup(lf)
	if err != nil {
		return errors.Wrap(err, "file can't be analyzed")
	}

	err = e.run(rpt, data)
	if err != nil {
		return errors.Wrap(err, "file can't be analyzed")
	}
	return nil
}

// It runs a set of goroutines to open and read lines from a file.
// Each line analyzed by a goroutine which call a FindIssue function to look for issues.
func (e *Engineer) run(rpt *reports.Report, data *reports.Dataset) (error) {

	logType, logName, logPath, nLines := data.LogFile.Type, data.LogFile.Name, data.LogFile.Path, data.LogFile.NLines

	log.Printf("| [engineer.go][run] type of log, %s\r", logType)
	log.Printf("| [engineer.go][run] number of lines %d", nLines)

	file, errO := os.Open(logPath)
	if errO != nil {
		return errors.Wrap(errO, fmt.Sprintf("file, %s, can't be open", logPath))
	}
	defer file.Close()

	go timer(10 * time.Millisecond)

	reader := bufio.NewReaderSize(file, 8192)
	line, err := reader.ReadString('\n')
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("file, %s, can't be read", logPath))
	}

	lCount := 0
	var wg sync.WaitGroup
	for {
		lCount++
		// engineer working
		go func(wg *sync.WaitGroup, nextLine string, data *reports.Dataset, rpt *reports.Report, c int) {
			wg.Add(1)
			defer wg.Done()
			rpt.FindIssue(wg, nextLine, data, rpt)
		}(&wg, line, data, rpt, lCount)
		line, err = reader.ReadString('\n')
		if err != nil {
			log.Printf("| [engineer.go][run] ERROR %s", err)
			defer func() {
				fmt.Println("")
				log.Printf("| [engineer.go][run] %d lines read from %s", lCount, logName)
				fmt.Println("")
			}()
			break
		}

	}

	go func(wg *sync.WaitGroup) {
		time.Sleep(2 * time.Second)
		wg.Wait()
	}(&wg)

	rpt.NumberOfFiles++

	return nil
}

func timer(delay time.Duration) {
	c := time.Tick(1 * time.Millisecond)
	for {
		for now := range c {
			fmt.Printf("\r%v", now)
			time.Sleep(delay)
		}
	}
	fmt.Println("")
}

func (e *Engineer) setup(lf *files.LogFile) (rpt *reports.Report, data *reports.Dataset, err error) {

	rpt, ok := e.Registry.Reports[lf.Type]
	if !ok {
		return nil, nil, errors.New(fmt.Sprintf("there is not report of type %s to be used, the report is nil", lf.Type))
	}

	mapRegEx, err := e.RetrieveFieldsRegExp(lf.Type)
	if err != nil {
		return nil, nil, errors.Wrap(err, "problem retrieving regular expressions for file analysis")
	}

	data = &reports.Dataset{LogFile: lf, MapRegEx: mapRegEx}

	return
}

func (e *Engineer) Publish() {
	log.Println("| [engineer.go][Publish]")
	reports := e.IterateReports()

	for rpt := range reports {
		if rpt.IssuesLength() != 0 {
			rpt.Publish()
		}

	}

}

func NewEngineer(c *config.Config, r *reports.Registry) (*Engineer, error) {

	if c == nil {
		return nil, errors.New("there is no configuration, it is nil")
	}
	if r == nil {
		return nil, errors.New("there is no registry, it is nil")
	}

	e := &Engineer{c, r}

	log.Println("| [engineer.go] [NewEngineer] created new engineer ")
	return e, nil
}
