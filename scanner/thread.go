package scanner

import (
	"fmt"
	"github.com/vflame6/bruter/logger"
	"github.com/vflame6/bruter/scanner/modules"
	"github.com/vflame6/bruter/utils"
	"os"
	"strings"
	"sync"
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

// SendCredentials sends credential pairs to the credentials channel.
// The done channel is closed by the caller when threads stop early,
// preventing this goroutine from leaking (Bug 3 fix).
func SendCredentials(credentials chan *modules.Credential, usernames, passwords string, done <-chan struct{}) {
	defer close(credentials)
	for linePwd := range utils.ParseFileByLine(passwords) {
		for lineUsername := range utils.ParseFileByLine(usernames) {
			select {
			case credentials <- &modules.Credential{Username: lineUsername, Password: linePwd}:
			case <-done:
				return
			}
		}
	}
}

// GetResults drains the results channel and writes each success to the log and output file.
// The wg WaitGroup is signalled Done when the channel is fully drained (Bug 2 fix).
func GetResults(results chan *Result, outputFile *os.File, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		result, ok := <-results
		if !ok {
			return
		}

		successString := fmt.Sprintf("[%s] %s:%d [%s] [%s]", result.Command, result.IP, result.Port, result.Username, result.Password)

		logger.Success(successString)

		if outputFile != nil {
			_, _ = outputFile.WriteString(successString + "\n")
		}
	}
}
