package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/venturasr/ll-analyzer/analyzer"
	"github.com/venturasr/ll-analyzer/config"
	"github.com/venturasr/ll-analyzer/files"
	"github.com/venturasr/ll-analyzer/reports"

	"github.com/pkg/errors"
	"github.com/venturasr/ll-analyzer/tools"
)

const workers = 20

var engineer *analyzer.Engineer

func main() {
	start := time.Now()
	log.Println("| [main.go][main] start=", start)
	defer func() {
		//report published before method returns. Everything ends here.
		engineer.Publish()
		fmt.Println("")
		log.Println("| [main.go][main] elapsed=", time.Since(start))
	}()

	err := setup()
	tools.ExitIfError(err)

	err = starts(getCurrentPath())
	if err != nil {
		fmt.Printf("%+v\n", err)
	}

}

func setup() (error) {

	//Configuration
	pathToConfig := getPathToConfig()
	path, err := filepath.Abs(pathToConfig)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("can't find path %s", path))
	}

	c, err := config.NewConfig(path)
	if err != nil {
		return errors.Wrap(err, "failed to create configuration")
	}

	//Reports registry
	r, err := reports.NewRegistry(c)
	if err != nil {
		return errors.Wrap(err, "failed to create registry")
	}

	//Engineer (engineer)
	eng, err := analyzer.NewEngineer(c, r)
	if err != nil {
		return errors.Wrap(err, "failed to create engineer")
	}

	engineer = eng
	fmt.Println("")
	return nil
}

func starts(path string) error {

	done := make(chan struct{})
	numberFilesChan := make(chan int)
	var totalFiles int
	defer func() {
		close(done)
		log.Printf("| [llanalyzer.go][starts] number of files loaded: %v", totalFiles)
	}()

	filesChan, errChan := collectFiles(done, path)

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			iterateFiles(done, numberFilesChan, filesChan)
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(numberFilesChan)
		//log.Printf("| [llanalyzer.go][starts] clousing number file channel -> 4 \n") //TODO DELETE IT
	}()

	for nf := range numberFilesChan {
		//log.Printf("| [llanalyzer.go][starts] counting files -> 2 \n") //TODO DELETE IT
		totalFiles += nf
	}

	if err := <-errChan; err != nil {
		return err
	}
	return nil
}

// collectFiles starts goroutines to walk the given root directory to collect log files.
func collectFiles(done <-chan struct{}, directory string) (<-chan *files.LogFile, <-chan error) {
	filesChan := make(chan *files.LogFile) //It sends log files paths
	errChan := make(chan error, 1)         //buffered channel

	go func() {
		var wg sync.WaitGroup
		err := filepath.Walk(directory, func(pathToFile string, info os.FileInfo, err error) error {

			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return nil
			}

			wg.Add(1)
			go func() {

				lf := files.NewLogFile(info, pathToFile)
				select {
				case filesChan <- lf:
				case <-done:
				default:
				}
				wg.Done()

			}()
			// Abort the walk if analysisDoneChan is closed.
			select {
			case <-done:
				return nil
			default:
			}

			return nil
		})

		// here the Walk function has returned, so all calls to wg.Add are done.
		// it starts a goroutine to close filesChan once all the sends are done.
		go func() {
			wg.Wait()
			close(filesChan)
		}()

		// No select needed here, since errChan is buffered.
		errChan <- err
	}()

	return filesChan, errChan
}

func iterateFiles(done <-chan struct{}, numberFilesChan chan<- int, filesChan <-chan *files.LogFile) {
	for lf := range filesChan {
		engineer.Analyze(lf)
		select {
		case numberFilesChan <- 1:
			//log.Printf("| [llanalyzer.go][iterateFiles] returning from engineer.Analyze() -> 1 \n") //TODO DELETE IT
		case <-done:
			return
		}
	}
}

func getCurrentPath() (currentpath string) {
	currentpath, err := tools.GetCurrentDir()
	tools.ExitIfError(err)
	log.Printf("| [main.go][getCurrentPath] execution directory - %s", currentpath)

	return currentpath
}

func getPathToConfig() (pathToConfig string) {
	homeDir, err := tools.GetHomeDir()
	tools.ExitIfError(err)
	pathToConfig = fmt.Sprintf("%s%sconfig.toml", homeDir, tools.PathSeparator)
	log.Printf("| [main.go][GetPathToConfig] config file location - %s", pathToConfig)

	return pathToConfig
}
