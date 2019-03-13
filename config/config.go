package config

import (
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"github.com/venturasr/ll-analyzer/tools"
)

type Config struct {
	Global       global
	Logs         []Logs
	FieldsRegExp fieldsRegExp
}

type global struct {
	ReportName string
}

type Logs struct {
	LogType          string
	TemplateFile     string
	ReportFileName   string
	Separator        string
	SplitSeparator   string
	LineMatchRegex   string
	FieldsMatchRegex [][]string
}

type fieldsRegExp struct {
	regExp map[string]map[string]*regexp.Regexp
}

func NewConfig(configFile string) (*Config, error) {
	c := new(Config)

	info, err := os.Stat(configFile)
	if os.IsNotExist(err) {
		return nil, errors.Wrap(err, fmt.Sprintf("can't find configuration file %s", configFile))
	}

	log.Printf("| [config.go][NewConfig] Configuration file found: %v", info.Name())

	if _, err := toml.DecodeFile(configFile, c); err != nil {
		return c, errors.Wrap(err, fmt.Sprintf("the content of the configuration file can't be decode. %s", configFile))
	}

	err = c.compileFieldsMatchRegex()
	if err != nil {
		return c, err
	}

	log.Println("| [config.go][NewConfig] created new configuration")

	return c, nil
}

func (conf *Config) compileFieldsMatchRegex() error {

	MapAllRegExp := make(map[string]map[string]*regexp.Regexp) // Contains all compiled regular expressions
	for _, configuredLog := range conf.Logs {

		mapRegEx := make(map[string]*regexp.Regexp)
		for _, regex := range configuredLog.FieldsMatchRegex {

			re := tools.CompileRegEx(regex[1])
			mapRegEx[regex[0]] = re

		}
		log.Printf("| [config.go][createRegExpressions] regular expressions for log type: %s", configuredLog.LogType)
		MapAllRegExp[configuredLog.LogType] = mapRegEx

	}

	conf.FieldsRegExp = fieldsRegExp{regExp: MapAllRegExp}
	return nil
}

func (c *Config) RetrieveFieldsRegExp(logType string) (map[string]*regexp.Regexp, error) {

	if m, ok := c.FieldsRegExp.regExp[logType]; ok {
		return m, nil
	}

	ret := fmt.Sprintf("there is not compiled regular expressions for log type, %s ", logType)

	return nil, errors.New(ret)

}
