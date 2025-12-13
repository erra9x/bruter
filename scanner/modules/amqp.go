package modules

import (
	"crypto/tls"
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/vflame6/bruter/utils"
	"net"
	"time"
)

// AMQPHandler is an implementation of ModuleHandler for AMQP service
func AMQPHandler(dialer *utils.ProxyAwareDialer, timeout time.Duration, target *Target, credential *Credential) (bool, error) {
	var conn *amqp.Connection
	var endpoint string
	var err error

	if target.Encryption {
		endpoint = fmt.Sprintf("amqps://%s:%s@%s:%d/", credential.Username, credential.Password, target.IP.String(), target.Port)
		conn, err = amqp.DialConfig(endpoint, amqp.Config{
			Dial: func(network, addr string) (net.Conn, error) {
				tlsConfig := utils.GetTLSConfig()
				c, err := dialer.Dial(network, addr)
				if err != nil {
					return nil, err
				}
				return tls.Client(c, tlsConfig), nil
			},
		})
	} else {
		endpoint = fmt.Sprintf("amqp://%s:%s@%s:%d/", credential.Username, credential.Password, target.IP.String(), target.Port)
		conn, err = amqp.DialConfig(endpoint, amqp.Config{
			Dial: dialer.Dial,
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

	// successful authentication
	_ = conn.Close()
	return true, nil
}
