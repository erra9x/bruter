package scanner

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

type Scanner struct {
	Delay          time.Duration
	OutputFileName string
	OutputFile     *os.File
	Usernames      []string
	Passwords      []string
	Targets        []net.IP
	Port           int
	Threads        sync.WaitGroup
	Mutex          sync.Mutex
}

func (s *Scanner) Stop() {
	_ = s.OutputFile.Close()
}

func (s *Scanner) RunClickHouse(targets string, port int) error {
	err := s.ImportTargets(targets, port)
	if err != nil {
		return err
	}

	// the program creates a separate thread for each target
	for _, target := range s.Targets {
		probe, secure, err := ProbeClickHouse(target, port)
		if err != nil {
			log.Println("[!] Failed to probe, maybe the target is not a ClickHouse server:", target, err)
			continue
		}
		if probe {
			log.Println("[+] Successfully logged with default username and empty password:", target)
			if s.OutputFile != nil {
				s.Mutex.Lock()
				_, _ = s.OutputFile.WriteString(fmt.Sprintf("[clickhouse] %s %s %s\n", target, "default", ""))
				s.Mutex.Unlock()
			}
			continue
		}
		s.Threads.Add(1)
		go ThreadClickHouse(&s.Threads, &s.Mutex, s.OutputFile, target, port, secure, &s.Usernames, &s.Passwords, &s.Delay)
	}
	s.Threads.Wait()

	return nil
}

func (s *Scanner) ImportTargets(filename string, port int) error {
	var targets []net.IP

	if !(port >= 1 && port <= 65535) {
		return errors.New("invalid port number")
	}
	s.Port = port

	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			test, err := ParseIPOrCIDR(filename)
			if err != nil {
				return err
			}
			for _, target := range test {
				targets = append(targets, net.ParseIP(target))
			}
			s.Targets = targets
			return nil
		}
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		test, err := ParseIPOrCIDR(scanner.Text())
		if err != nil {
			return err
		}
		for _, target := range test {
			targets = append(targets, net.ParseIP(target))
		}
	}
	s.Targets = targets

	return nil
}
