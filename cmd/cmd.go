package cmd

import (
	"bufio"
	"github.com/vflame6/bruter/scanner"
	"os"
	"time"
)

func CreateScanner(timeout int, output string, parallel, threads, delay int, username, password string) (*scanner.Scanner, error) {
	var outputFile *os.File
	var passwords []string

	if output != "" {
		var err error
		outputFile, err = os.Create(output)
		if err != nil {
			return nil, err
		}
	}

	usernames, err := ParseUsernames(username)
	if err != nil {
		return nil, err
	}

	if CheckIfFileExists(password) {
		passwordFile, err := os.Open(password)
		if err != nil {
			return nil, err
		}
		defer passwordFile.Close()
		sc := bufio.NewScanner(passwordFile)
		for sc.Scan() {
			passwords = append(passwords, sc.Text())
		}
	} else {
		passwords = []string{password}
	}

	options := scanner.Options{
		Timeout:        time.Duration(timeout) * time.Second,
		Threads:        threads,
		Delay:          time.Duration(delay) * time.Millisecond,
		OutputFileName: output,
		OutputFile:     outputFile,
		Usernames:      usernames,
		Passwords:      passwords,
	}

	s := scanner.Scanner{
		Opts:     &options,
		Parallel: parallel,
	}

	return &s, nil
}

func ParseUsernames(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{filename}, nil
		}
		return nil, err
	}
	defer file.Close()

	sc := bufio.NewScanner(file)
	var usernames []string
	for sc.Scan() {
		usernames = append(usernames, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return usernames, nil
}

func CheckIfFileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
