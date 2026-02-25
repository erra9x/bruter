package modules

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/vflame6/bruter/utils"
)

// AsteriskHandler is an implementation of ModuleHandler for the Asterisk Manager Interface (AMI).
// Connects on port 5038, reads the banner, sends Action: Login, and parses the response block.
func AsteriskHandler(ctx context.Context, dialer *utils.ProxyAwareDialer, timeout time.Duration, target *Target, credential *Credential) (bool, error) {
	addr := target.Addr()

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return false, err
	}
	defer func() { _ = conn.Close() }()

	_ = conn.SetDeadline(time.Now().Add(timeout))

	reader := bufio.NewReader(conn)

	// Read banner line — must contain "Asterisk".
	banner, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	if !strings.Contains(banner, "Asterisk") {
		return false, fmt.Errorf("not an AMI server: %q", strings.TrimSpace(banner))
	}

	// Send login action (blank line terminates the block).
	_, err = fmt.Fprintf(conn, "Action: Login\r\nUsername: %s\r\nSecret: %s\r\n\r\n",
		credential.Username, credential.Password)
	if err != nil {
		return false, err
	}

	// Read response lines until blank line.
	var lines []string
	for {
		line, err := reader.ReadString('\n')
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			lines = append(lines, trimmed)
		}
		if err != nil || trimmed == "" {
			break
		}
	}
	response := strings.Join(lines, " ")

	switch {
	case strings.Contains(response, "Response: Success"):
		return true, nil
	case strings.Contains(response, "Response: Error"):
		return false, nil
	default:
		return false, fmt.Errorf("unexpected AMI response: %s", response)
	}
}
