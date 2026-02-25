package modules

import (
	"context"
	"strings"
	"time"

	"github.com/mitchellh/go-vnc"
	"github.com/vflame6/bruter/utils"
)

// VNCHandler is an implementation of ModuleHandler for VNC RFB password authentication.
// The credential.Username is ignored — VNC uses only a password.
func VNCHandler(ctx context.Context, dialer *utils.ProxyAwareDialer, timeout time.Duration, target *Target, credential *Credential) (bool, error) {
	addr := target.Addr()

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return false, err
	}
	defer func() { _ = conn.Close() }()

	_ = conn.SetDeadline(time.Now().Add(timeout))

	config := &vnc.ClientConfig{
		Auth: []vnc.ClientAuth{
			&vnc.PasswordAuth{Password: credential.Password},
		},
	}

	client, err := vnc.Client(conn, config)
	if err == nil {
		_ = client.Close()
		return true, nil
	}

	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "authentication failed") ||
		strings.Contains(msg, "too many authentication failures") ||
		strings.Contains(msg, "auth") {
		return false, nil
	}
	return false, err
}
