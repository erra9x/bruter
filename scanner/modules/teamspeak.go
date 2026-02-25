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

// TeamSpeakHandler is an implementation of ModuleHandler for TeamSpeak 3 ServerQuery.
// Connects on port 10011, reads the TS3 banner, logs in, and parses the error response.
func TeamSpeakHandler(_ context.Context, dialer *utils.ProxyAwareDialer, timeout time.Duration, target *Target, credential *Credential) (bool, error) {
	addr := net.JoinHostPort(target.IP.String(), strconv.Itoa(target.Port))

	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return false, err
	}
	defer func() { _ = conn.Close() }()

	_ = conn.SetDeadline(time.Now().Add(timeout))

	reader := bufio.NewReader(conn)

	// Read banner — first line must be "TS3".
	line1, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	if !strings.Contains(line1, "TS3") {
		return false, fmt.Errorf("not a TS3 server: %q", strings.TrimSpace(line1))
	}

	// Read second banner line (welcome message) — discard.
	_, _ = reader.ReadString('\n')

	// Send login command.
	if _, err = fmt.Fprintf(conn, "login %s %s\n", credential.Username, credential.Password); err != nil {
		return false, err
	}

	// Read response lines until we see "error id=".
	var respLines []string
	for {
		line, err := reader.ReadString('\n')
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			respLines = append(respLines, trimmed)
		}
		if strings.Contains(line, "error id=") {
			break
		}
		if err != nil {
			break
		}
	}
	response := strings.Join(respLines, " ")

	switch {
	case strings.Contains(response, "error id=0 msg=ok"):
		return true, nil
	case strings.Contains(response, "error id="):
		return false, nil
	default:
		return false, fmt.Errorf("unexpected TS3 response: %s", response)
	}
}
