package scanner

import (
	"fmt"
	"github.com/vflame6/bruter/logger"
	"github.com/vflame6/bruter/scanner/modules"
	"github.com/vflame6/bruter/utils"
	"os"
	"strings"
)

func SendTargets(targets chan *modules.Target, defaultPort int, filename string) {
	for line := range utils.ParseFileByLine(filename) {
		t := strings.TrimSpace(line)
		if t == "" {
			// skip empty lines
			continue
		}
		target, err := ParseTarget(line, defaultPort)
		if err != nil {
			logger.Debugf("can't parse line %s as host or host:port, ignoring", line)
			continue
		}
		targets <- target
	}

	close(targets)
}

func SendCredentials(credentials chan *modules.Credential, usernames, passwords string) {
	for linePwd := range utils.ParseFileByLine(passwords) {
		for lineUsername := range utils.ParseFileByLine(usernames) {
			credentials <- &modules.Credential{Username: lineUsername, Password: linePwd}
		}
	}

	close(credentials)
}

func GetResults(results chan *Result, outputFile *os.File) {
	for {
		result, ok := <-results
		if !ok {
			return
		}

		successString := fmt.Sprintf("[%s] %s:%d [%s] [%s]", result.Command, result.IP, result.Port, result.Username, result.Password)

		logger.Successf(successString)

		if outputFile != nil {
			_, _ = outputFile.WriteString(successString + "\n")
		}
	}
}
