package modules

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/vflame6/bruter/utils"
	"golang.org/x/crypto/ssh"
)

// SSHKeyHandler is an implementation of ModuleHandler for SSH public key authentication.
//
// credential.Username = SSH username
// credential.Password = path to a PEM private key file, OR raw PEM content starting with "-----"
func SSHKeyHandler(ctx context.Context, dialer *utils.ProxyAwareDialer, timeout time.Duration, target *Target, credential *Credential) (bool, error) {
	addr := target.Addr()

	// Load private key — either raw PEM or a file path.
	pemData := []byte(credential.Password)
	if !bytes.HasPrefix(pemData, []byte("-----")) {
		var err error
		pemData, err = os.ReadFile(credential.Password)
		if err != nil {
			return false, err
		}
	}

	signer, err := ssh.ParsePrivateKey(pemData)
	if err != nil {
		// Try with empty passphrase (encrypted key with no passphrase set).
		signer, err = ssh.ParsePrivateKeyWithPassphrase(pemData, nil)
		if err != nil {
			return false, fmt.Errorf("invalid key: %w", err)
		}
	}

	supported := ssh.SupportedAlgorithms()
	insecure := ssh.InsecureAlgorithms()

	config := &ssh.ClientConfig{
		User: credential.Username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		Timeout:         timeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Config: ssh.Config{
			KeyExchanges: append(supported.KeyExchanges, insecure.KeyExchanges...),
			Ciphers:      append(supported.Ciphers, insecure.Ciphers...),
			MACs:         append(supported.MACs, insecure.MACs...),
		},
		HostKeyAlgorithms: append(supported.HostKeys, insecure.HostKeys...),
	}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return false, err
	}

	sshConn, _, _, err := ssh.NewClientConn(conn, addr, config)
	if err == nil {
		_ = sshConn.Close()
		_ = conn.Close()
		return true, nil
	}
	_ = conn.Close()

	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "permission denied") || strings.Contains(msg, "no supported") {
		return false, nil
	}
	return false, err
}
