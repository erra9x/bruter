package scanner

import (
	"errors"
	"net"
	"os"
	"strconv"
	"strings"
)

func ParseTarget(target string, defaultPort int) (*Target, error) {
	testTarget := strings.Split(target, ":")

	if len(testTarget) == 2 {
		ip := net.ParseIP(testTarget[0])
		if ip == nil {
			return nil, errors.New("invalid IP address")
		}

		port, err := strconv.Atoi(testTarget[1])
		if err != nil {
			return nil, err
		}
		if !(port >= 1 && port <= 65535) {
			return nil, errors.New("invalid port number, format 1-65535")
		}

		return &Target{IP: ip, Port: port, Encryption: false}, nil
	}
	if len(testTarget) == 1 {
		ip := net.ParseIP(testTarget[0])
		if ip == nil {
			return nil, errors.New("invalid ip address")
		}
		return &Target{IP: ip, Port: defaultPort, Encryption: false}, nil
	}
	return nil, errors.New("target is not a valid IP, IP:PORT or filename")
}

// IsFileExists checks if a file exists at the given path.
func IsFileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if err == nil {
		return true // File exists
	}
	if errors.Is(err, os.ErrNotExist) {
		return false // File does not exist
	}
	return false
}
