package scanner

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func ThreadClickHouse(wg *sync.WaitGroup, mutex *sync.Mutex, outputFile *os.File, address net.IP, port int, secure bool, usernames *[]string, passwords *[]string, sleep *time.Duration) {
	for _, password := range *passwords {
		for _, username := range *usernames {
			conn, err := GetClickHouseConnection(address, port, secure, username, password)
			if err != nil {
				time.Sleep(*sleep)
				continue
			}
			defer conn.Close()
			err = conn.Ping(context.Background())
			if err != nil {
				time.Sleep(*sleep)
				continue
			}
			log.Println("[+] [clickhouse]", address, username, password)
			if outputFile != nil {
				mutex.Lock()
				_, _ = outputFile.WriteString(fmt.Sprintf("[clickhouse] %s %s %s\n", address, username, password))
				mutex.Unlock()
			}
			time.Sleep(*sleep)
		}
	}
	wg.Done()
}

func GetClickHouseConnection(address net.IP, port int, secure bool, username, password string) (driver.Conn, error) {
	addr := net.JoinHostPort(address.String(), strconv.Itoa(port))

	opts := &clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: username,
			Password: password,
		},
		DialTimeout: 5 * time.Second,
	}

	if secure {
		opts.TLS = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func TestClickHouseConnection(address net.IP, port int, secure bool, username, password string) (driver.Conn, string, error) {
	addr := net.JoinHostPort(address.String(), strconv.Itoa(port))

	opts := &clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: username,
			Password: password,
		},
		DialTimeout: 5 * time.Second,
	}

	if secure {
		opts.TLS = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, classifyError(err), err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.Ping(ctx); err != nil {
		conn.Close()
		return nil, classifyError(err), err
	}

	return conn, "", nil
}

// ProbeClickHouse is a function to test login to ClickHouse with default credentials
// return values are SUCCESS, TLS, error
func ProbeClickHouse(address net.IP, port int) (success bool, secure bool, err error) {
	// Try TLS first
	conn, errType, err := TestClickHouseConnection(address, port, true, "default", "")
	if err == nil {
		defer conn.Close()
		return true, true, nil
	}

	// If it's an auth error, no point trying without TLS with same creds
	if errType == "auth_error" {
		return false, true, nil
	}

	// If it's a TLS error, try plaintext
	if errType == "tls_error" {
		conn, errType, err = TestClickHouseConnection(address, port, false, "default", "")
		if err == nil {
			defer conn.Close()
			return true, false, nil
		}

		if errType == "auth_error" {
			return false, false, nil
		}
	}

	return false, false, fmt.Errorf("connection failed: %w", err)
}

func classifyError(err error) string {
	if err == nil {
		return "no error"
	}

	// Check for ClickHouse protocol errors (including auth)
	var chErr *clickhouse.Exception
	if errors.As(err, &chErr) {
		// Error codes: https://github.com/ClickHouse/ClickHouse/blob/master/src/Common/ErrorCodes.cpp
		switch chErr.Code {
		case 516: // AUTHENTICATION_FAILED
			return "auth_error"
		case 192: // UNKNOWN_USER
			return "auth_error"
		case 193: // WRONG_PASSWORD
			return "auth_error"
		case 194: // REQUIRED_PASSWORD
			return "auth_error"
		default:
			return "clickhouse_error"
		}
	}

	// Check for TLS errors
	var tlsRecordErr tls.RecordHeaderError
	if errors.As(err, &tlsRecordErr) {
		return "tls_error"
	}

	// Check for certificate errors
	var certErr *tls.CertificateVerificationError
	if errors.As(err, &certErr) {
		return "tls_error"
	}

	// Some TLS errors come as plain errors with specific messages
	if strings.Contains(err.Error(), "tls:") ||
		strings.Contains(err.Error(), "first record does not look like a TLS handshake") {
		return "tls_error"
	}

	// Network errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return "timeout_error"
		}
		return "network_error"
	}

	// Connection refused
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return "connection_error"
	}

	return "unknown_error"
}
