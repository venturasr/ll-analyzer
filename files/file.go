package files

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/venturasr/ll-analyzer/config"
	"github.com/venturasr/ll-analyzer/tools"
)

type fileInfo interface {
	TypeOfLog(line string, c *config.Config) error
}

type LogFile struct {
	Type           string //jdbc, gc, console, thread-dump, etc.
	Path           string
	Name           string
	Size           int
	Separator      string
	SplitSeparator string
	NLines         int
	NChunks        int
	NRows          int
}

func NewLogFile(fi os.FileInfo, pathToFile string) (*LogFile) {
	lf := new(LogFile)

	if !fi.IsDir() {
		lf.Path = pathToFile
		lf.Name = fi.Name()
		lf.Size = int(fi.Size())
		nlines, _ := tools.LineCounter(lf.Path)
		lf.NLines = nlines
		lf.NChunks = tools.CalcNumChunks(nlines)
		if lf.NChunks > 0 && lf.NLines > 0 {
			lf.NRows = lf.NLines / lf.NChunks
		}

	}

	return lf

}

func (lf *LogFile) TypeOfLog(path string, c *config.Config) error {

	strLine, err := tools.FirstLine(path)
	if err != nil {
		return err
	}

	for _, configuredLog := range c.Logs {
		lmRegex := configuredLog.LineMatchRegex
		rg := tools.CompileRegEx(lmRegex)
		if rg.MatchString(strLine) {
			lf.Type = configuredLog.LogType
			lf.Separator = configuredLog.Separator
			lf.SplitSeparator = configuredLog.SplitSeparator
			break
		}
	}

	if lf.Type == "" {
		return errors.New(fmt.Sprintf("there is no pattern to match with this file %s. No log type found", lf.Name))
	}
	return nil
}
