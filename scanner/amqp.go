package scanner

import (
	"crypto/tls"
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/vflame6/bruter/logger"
	"net"
	"time"
)

// AMQPChecker is an implementation of CheckerHandler for AMQP service
func AMQPChecker(target *Target, opts *Options) (bool, bool, error) {
	defaultUsername := "guest"
	defaultPassword := "guest"

	success := false
	secure := false

	// try with encryption first
	probe, err := ProbeAMQP(target.IP, target.Port, true, opts.Timeout, defaultUsername, defaultPassword)
	if err == nil {
		secure = true
		if probe {
			RegisterSuccess(opts.OutputFile, &opts.FileMutex, opts.Command, target, defaultUsername, defaultPassword)
			success = true
		}
	} else {
		logger.Debugf("(%s:%d) failed to connect to AMQP with encryption, trying plaintext", target.IP, target.Port)
		// connect via plaintext FTP
		probe, err = ProbeAMQP(target.IP, target.Port, false, opts.Timeout, defaultUsername, defaultPassword)
		if err == nil {
			if probe {
				RegisterSuccess(opts.OutputFile, &opts.FileMutex, opts.Command, target, defaultUsername, defaultPassword)
				success = true
			}
		} else {
			// if nothing succeeded, return error
			return false, false, err
		}
	}

	return success, secure, nil
}

// AMQPHandler is an implementation of CommandHandler for AMQP service
func AMQPHandler(opts *Options, target *Target, credential *Credential) (bool, bool) {
	probe, err := ProbeAMQP(target.IP, target.Port, target.Encryption, opts.Timeout, credential.Username, credential.Password)
	if err != nil {
		// not connected
		return false, false
	}

	// connected and authenticated or not
	return true, probe
}

func ProbeAMQP(ip net.IP, port int, encryption bool, timeout time.Duration, username, password string) (bool, error) {
	var conn *amqp.Connection
	var endpoint string
	var err error

	if encryption {
		endpoint = fmt.Sprintf("amqps://%s:%s@%s:%d/", username, password, ip.String(), port)
		conn, err = amqp.DialConfig(endpoint, amqp.Config{
			Dial: func(network, addr string) (net.Conn, error) {
				tlsConfig := &tls.Config{InsecureSkipVerify: true}
				conn, err := net.DialTimeout(network, addr, timeout)
				if err != nil {
					return nil, err
				}
				return tls.Client(conn, tlsConfig), nil
			},
		})
	} else {
		endpoint = fmt.Sprintf("amqp://%s:%s@%s:%d/", username, password, ip.String(), port)
		conn, err = amqp.DialConfig(endpoint, amqp.Config{
			Dial: func(network, addr string) (net.Conn, error) {
				return net.DialTimeout(network, addr, timeout)
			},
		})
	}

	if err != nil {
		if errors.Is(err, amqp.ErrCredentials) {
			// failed authentication
			return false, nil
		}
		// failed connection
		return false, err
	}

	defer conn.Close()
	// successful authentication or unusual error
	return true, nil
}
