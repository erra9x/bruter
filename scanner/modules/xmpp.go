package modules

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/vflame6/bruter/utils"
	"github.com/xmppo/go-xmpp"
)

// XMPPHandler is an implementation of ModuleHandler for XMPP SASL authentication.
// Supports plain XMPP (port 5222) and XMPP over TLS.
func XMPPHandler(ctx context.Context, _ *utils.ProxyAwareDialer, timeout time.Duration, target *Target, credential *Credential) (bool, error) {
	addr := target.Addr()

	// Use domain from OriginalTarget when available (for SASL JID construction).
	host := target.IP.String()
	if net.ParseIP(target.OriginalTarget) == nil {
		host = target.OriginalTarget
	}

	tlsCfg := utils.GetTLSConfig()
	tlsCfg.ServerName = host

	options := xmpp.Options{
		Host:                         addr,
		User:                         credential.Username + "@" + host,
		Password:                     credential.Password,
		NoTLS:                        !target.Encryption,
		InsecureAllowUnencryptedAuth: true,
		TLSConfig:                    tlsCfg,
		DialTimeout:                  timeout,
	}

	// Run NewClient in a goroutine so we can respect context cancellation.
	type result struct {
		client *xmpp.Client
		err    error
	}
	ch := make(chan result, 1)
	go func() {
		c, err := options.NewClient()
		ch <- result{c, err}
	}()

	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case r := <-ch:
		if r.err == nil {
			_ = r.client.Close()
			return true, nil
		}

		msg := strings.ToLower(r.err.Error())
		if strings.Contains(msg, "not-authorized") || strings.Contains(msg, "authentication failed") {
			return false, nil
		}
		return false, r.err
	}
}
