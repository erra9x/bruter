package scanner

import (
	"bufio"
	"errors"
	"github.com/vflame6/bruter/logger"
	"net"
	"os"
	"sync"
	"time"
)

var DefaultPorts = map[string]int{
	"clickhouse": 9000,
}

var CommandHandlers = map[string]CommandHandler{
	"clickhouse": ClickHouseHandler,
}

var CheckerHandlers = map[string]CheckerHandler{
	"clickhouse": ClickHouseChecker,
}

// CommandHandler is an interface for one bruteforcing thread
type CommandHandler func(wg *sync.WaitGroup, credentials <-chan *Credential, opts *Options, target *Target)

// CheckerHandler is an interface for service checker function
// the return values are:
// DEFAULT (bool) for test if the target has default credentials
// ENCRYPTION (bool) for test if the target is using encryption
// ERROR (error) for connection errors
// if checker could not be implemented for target service, the checker must return false, false, nil
type CheckerHandler func(target *Target, opts *Options) (bool, bool, error)

type Scanner struct {
	Opts     *Options
	Targets  []*Target
	Port     int
	Parallel int
}

type Options struct {
	Timeout        time.Duration
	Threads        int
	Delay          time.Duration
	OutputFileName string
	OutputFile     *os.File
	Usernames      []string
	Passwords      []string
	Mutex          sync.Mutex
}

type Target struct {
	IP         net.IP
	Port       int
	Encryption bool
}

type Credential struct {
	Username string
	Password string
}

func (s *Scanner) Stop() {
	_ = s.Opts.OutputFile.Close()
}

// Run method is used to handle parallel execution
func (s *Scanner) Run(command, targets string) error {
	// check if command is valid
	handler, ok := CommandHandlers[command]
	if !ok {
		return errors.New("unknown command")
	}
	checker, ok := CheckerHandlers[command]
	if !ok {
		return errors.New("unknown command")
	}

	// check if delay is set
	if s.Opts.Delay > 0 {
		s.Opts.Threads = 1
	}

	// import targets
	err := s.ImportTargets(command, targets)
	if err != nil {
		return err
	}

	var parallelWg sync.WaitGroup
	parallelTargets := make(chan *Target, 256)

	go func() {
		for _, target := range s.Targets {
			defaultCreds, encryption, err := checker(target, s.Opts)
			if err != nil {
				logger.Debug(err)
				continue
			}
			if encryption {
				target.Encryption = true
			}
			if defaultCreds {
				continue
			}

			parallelTargets <- target
		}
		close(parallelTargets)
	}()

	for i := 0; i < s.Parallel; i++ {
		parallelWg.Add(1)

		go s.ThreadedHandler(&parallelWg, parallelTargets, handler)
	}
	parallelWg.Wait()

	return nil
}

func (s *Scanner) ThreadedHandler(wg *sync.WaitGroup, targets <-chan *Target, handler CommandHandler) {
	defer wg.Done()

	for {
		target, ok := <-targets
		if !ok {
			break
		}

		credentials := make(chan *Credential, 256)
		go func() {
			for _, password := range s.Opts.Passwords {
				for _, username := range s.Opts.Usernames {
					credentials <- &Credential{Username: username, Password: password}
				}
			}
			close(credentials)
		}()

		var threadWg sync.WaitGroup
		for i := 0; i < s.Opts.Threads; i++ {
			threadWg.Add(1)
			go handler(&threadWg, credentials, s.Opts, target)
		}
		threadWg.Wait()
	}
}

func (s *Scanner) ImportTargets(command, filename string) error {
	defaultPort := DefaultPorts[command]

	var targets []*Target

	if IsFileExists(filename) {
		file, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			target, err := ParseTarget(line, defaultPort)
			if err != nil {
				logger.Debugf("can't parse line %s as host or host:port, ignoring", line)
				continue
			}
			targets = append(targets, target)
		}
	} else {
		target, err := ParseTarget(filename, defaultPort)
		if err != nil {
			return err
		}
		targets = append(targets, target)
	}

	if len(targets) == 0 {
		return errors.New("no targets found: " + filename)
	}
	s.Targets = targets
	return nil
}
