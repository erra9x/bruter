package modules

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql" // register mysql driver
	"github.com/vflame6/bruter/utils"
)

// MySQLHandler is an implementation of ModuleHandler for MySQL service.
func MySQLHandler(ctx context.Context, dialer *utils.ProxyAwareDialer, timeout time.Duration, target *Target, credential *Credential) (bool, error) {
	addr := net.JoinHostPort(target.IP.String(), strconv.Itoa(target.Port))

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/", credential.Username, credential.Password, addr)
	if target.Encryption {
		dsn += "?tls=skip-verify"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return false, err
	}
	defer func() { _ = db.Close() }()

	db.SetConnMaxLifetime(timeout)

	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err = db.PingContext(pingCtx); err == nil {
		return true, nil
	}

	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "access denied") || strings.Contains(msg, "1045") {
		return false, nil
	}
	return false, err
}
