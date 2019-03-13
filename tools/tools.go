package tools

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

var PathSeparator = fmt.Sprintf("%c", filepath.Separator)

var Lock = sync.RWMutex{}

type FileChunks struct {
	NumOfChunks int
	NumOfRows   int
	Reminder    int
	SliceChunks [][]string
	AllLines    []string
}

func ExitIfError(err error) {
	if err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
}

func CompileRegEx(expression string) (*regexp.Regexp) {
	return regexp.MustCompile(expression)
}

func MatchStringInLine(line string, regex *regexp.Regexp, Separator string) string {

	if &regex != nil {
		rtSrt := regex.FindString(line)
		rtSrt = strings.Replace(rtSrt, Separator, "", strings.Count(rtSrt, Separator))
		return rtSrt
	}

	return ""
}

func LineCounter(path string) (int, error) {
	count := 0

	ioreader, err := os.Open(path)
	if err != nil {
		return count, errors.Wrap(err, fmt.Sprintf("file, %s, can't be open", path))

	}
	defer ioreader.Close()

	buf := make([]byte, 32*1024)

	lineSep := []byte{'\n'}

	for {
		c, err := ioreader.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}

}

func FirstLine(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("file, %s, can't be open", path))
	}
	bf := bufio.NewReader(f)

	line, isPrefix, err := bf.ReadLine()
	if err == io.EOF {
		return "", errors.Wrap(err, fmt.Sprintf("end of line reached when reading file, %s", path))
	}
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("file, %s, can't be readed", path))
	}

	if isPrefix {
		log.Fatal("Error: Unexpected long line reading", f.Name())
	}

	lineLen := len(line)
	lineAfterTrim := strings.TrimSpace(string(line))

	if lineLen == 0 || lineAfterTrim == "" {
		return "", errors.New(fmt.Sprintf("log file %s can't be identified as a log. First line is empty", path))
	}
	return string(line), nil

}

func CalcNumChunks(nlines int) (nChunks int) {
	nChunks = 0

	if nlines <= 100 {
		nChunks = 1
		return
	}

	linesStr := fmt.Sprint(nlines)
	for _, strDigit := range linesStr {
		digit, _ := strconv.ParseInt(string(strDigit), 10, 0)
		nChunks += int(digit)
	}

	return nChunks
}

func (fchunks *FileChunks) chunksSlice() (sliceChunks [][]string) {

	idxLines := 0
	sliceChunks = make([][]string, fchunks.NumOfChunks, fchunks.NumOfChunks)
	for index1 := range sliceChunks {
		if index1 == fchunks.NumOfChunks-1 && fchunks.Reminder > 0 {
			fchunks.NumOfRows = fchunks.Reminder
		}
		sliceChunks[index1] = make([]string, fchunks.NumOfRows, fchunks.NumOfRows)
		for index2 := range sliceChunks[index1] {
			sliceChunks[index1][index2] = fchunks.AllLines[idxLines]
			//fmt.Printf("\n[index1=%v][index2=%v]", index1, index2)
			idxLines++
		}
	}

	return sliceChunks
}

func ReadMap(someMap map[string]string, someKey string) (string, bool) {
	Lock.RLock()
	defer Lock.RUnlock()
	value, ok := someMap[someKey]
	return value, ok
}

func WriteMap(someMap map[string]string, someKey string, someValue string) {
	Lock.Lock()
	defer Lock.Unlock()

	someMap[someKey] = someValue
}

func GetCurrentDir() (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return pwd, nil
}

func GetHomeDir() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	homeDir := filepath.Dir(ex)
	return homeDir, nil
}
