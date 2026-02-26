package modules

import (
	"context"
	"net"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/vflame6/bruter/utils"
)

// LDAPHandler is an implementation of ModuleHandler for LDAP/LDAPS simple bind authentication.
// Supports plain LDAP (port 389) and LDAPS (port 636) with TLS.
func LDAPHandler(ctx context.Context, _ *utils.ProxyAwareDialer, timeout time.Duration, target *Target, credential *Credential) (bool, error) {
	addr := target.Addr()

	d := &net.Dialer{Timeout: timeout}

	var (
		conn *ldap.Conn
		err  error
	)
	if target.Encryption {
		conn, err = ldap.DialURL("ldaps://"+addr, ldap.DialWithDialer(d), ldap.DialWithTLSConfig(utils.GetTLSConfig()))
	} else {
		conn, err = ldap.DialURL("ldap://"+addr, ldap.DialWithDialer(d))
	}
	if err != nil {
		return false, err
	}
	defer func() { _ = conn.Close() }()

	// Close connection if context is cancelled.
	go func() {
		select {
		case <-ctx.Done():
			_ = conn.Close()
		}
	}()

	conn.SetTimeout(timeout)

	if err = conn.Bind(credential.Username, credential.Password); err == nil {
		return true, nil
	}

	if ldapErr, ok := err.(*ldap.Error); ok && ldapErr.ResultCode == ldap.LDAPResultInvalidCredentials {
		return false, nil
	}
	return false, err
}
