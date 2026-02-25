package modules

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/vflame6/bruter/utils"
)

// CiscoEnableHandler is an implementation of ModuleHandler for Cisco IOS enable-mode
// password brute-force. It first logs in to reach user-mode ">", then sends "enable"
// and tries credential.Password as the enable secret.
//
// credential.Username = the login username (used to reach user-mode)
// credential.Password = the enable secret to test
func CiscoEnableHandler(_ context.Context, dialer *utils.ProxyAwareDialer, timeout time.Duration, target *Target, credential *Credential) (bool, error) {
	addr := net.JoinHostPort(target.IP.String(), strconv.Itoa(target.Port))

	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return false, err
	}
	defer func() { _ = conn.Close() }()

	_ = conn.SetDeadline(time.Now().Add(timeout))

	reader := bufio.NewReader(conn)

	// Step 1 — login to reach user mode ">".
	initial, err := readUntilPrompt(reader, []string{"Username:", "login:", "Password:"})
	if err != nil {
		return false, err
	}

	lower := strings.ToLower(initial)
	if strings.Contains(lower, "username:") || strings.Contains(lower, "login:") {
		// Send username.
		if _, err = fmt.Fprintf(conn, "%s\r\n", credential.Username); err != nil {
			return false, err
		}
		// Wait for password prompt.
		if _, err = readUntilPrompt(reader, []string{"Password:"}); err != nil {
			return false, err
		}
	}

	// Send login password (same as username — common default).
	if _, err = fmt.Fprintf(conn, "%s\r\n", credential.Username); err != nil {
		return false, err
	}

	// Read until user-mode or failure.
	loginResp, err := readUntilPrompt(reader, []string{">", "#", "invalid", "failed", "Authentication failed"})
	if err != nil {
		return false, err
	}
	if !strings.Contains(loginResp, ">") && !strings.Contains(loginResp, "#") {
		// Could not reach user-mode — login failed, cannot test enable.
		return false, nil
	}

	// Step 2 — try enable.
	if _, err = fmt.Fprintf(conn, "enable\r\n"); err != nil {
		return false, err
	}

	if _, err = readUntilPrompt(reader, []string{"Password:"}); err != nil {
		return false, err
	}

	if _, err = fmt.Fprintf(conn, "%s\r\n", credential.Password); err != nil {
		return false, err
	}

	enableResp, err := readUntilPrompt(reader, []string{"#", "% Access denied", "% Bad passwords", "% No password set"})
	if err != nil {
		return false, err
	}

	if strings.Contains(enableResp, "#") {
		return true, nil
	}
	return false, nil
}
