# Module Review: bruter vs hydra + Go Libraries

**Date:** 2026-02-26  
**Task:** #111  
**Reviewer:** Erra

## Summary

Reviewed all 38 modules. Overall quality is solid — most use well-maintained Go libraries with proper error classification. Key gaps vs hydra are in SASL auth method diversity for mail protocols and some edge cases.

## Legend

- ✅ Correct, no action needed
- ⚠️ Minor improvement possible
- 🔧 Fix recommended

---

## Module Reviews

### FTP ✅
- **Library:** github.com/jlaffaye/ftp (well-maintained)
- **vs hydra:** Equivalent — both do USER/PASS. Hydra also has FTP Bounce, but that's a different attack.
- **Action:** None

### SSH ✅
- **Library:** golang.org/x/crypto/ssh (official)
- **vs hydra:** Good — includes insecure algorithms for legacy servers, classifies "method not allowed" errors. Hydra also supports keyboard-interactive; bruter only does password auth.
- **Action:** None (keyboard-interactive is edge case)

### SSHKey ✅
- **Library:** golang.org/x/crypto/ssh
- **vs hydra:** Equivalent. Supports raw PEM and file paths. Good.
- **Action:** None

### SMTP ⚠️
- **Library:** net/smtp (stdlib)
- **vs hydra:** Hydra supports PLAIN, LOGIN, CRAM-MD5, DIGEST-MD5, NTLM. We only do PLAIN via `smtp.PlainAuth`. LOGIN is very common on legacy servers.
- **Go lib:** emersion/go-sasl can do LOGIN, CRAM-MD5
- **Action:** Add LOGIN auth fallback. CRAM-MD5/NTLM are lower priority.

### POP3 ✅
- **Library:** Raw TCP (correct for POP3)
- **vs hydra:** Both do USER/PASS. Hydra also does APOP, CRAM-MD5, DIGEST-MD5, NTLM.
- **Action:** None — USER/PASS covers 99% of POP3 servers

### IMAP ⚠️
- **Library:** Raw TCP
- **vs hydra:** We do LOGIN only. Hydra supports PLAIN, CRAM-MD5, DIGEST-MD5, NTLM.
- **Note:** Our IMAP LOGIN doesn't escape quotes in username/password. If a password contains `"`, the LOGIN command breaks.
- **Action:** Escape `"` and `\` in LOGIN arguments. Consider adding PLAIN auth.

### Redis ✅
- **Library:** github.com/redis/go-redis/v9 (official, well-maintained)
- **vs hydra:** Equivalent. We additionally support ACL (username + password).
- **Action:** None

### MySQL ✅
- **Library:** github.com/go-sql-driver/mysql (standard)
- **vs hydra:** Equivalent. Both use native MySQL auth. Hydra doesn't support mysql_native_password vs caching_sha2 explicitly — the Go driver handles it automatically.
- **Note:** Proxy dialer not integrated — using default sql.Open. Low priority since MySQL rarely uses proxies.
- **Action:** None

### PostgreSQL ✅
- **Library:** github.com/lib/pq (mature)
- **vs hydra:** Equivalent. Both do md5/password auth.
- **Action:** None

### MSSQL ✅
- **Library:** github.com/microsoft/go-mssqldb (official Microsoft driver)
- **vs hydra:** Equivalent.
- **Action:** None

### MongoDB ✅
- **Library:** go.mongodb.org/mongo-driver/v2 (official)
- **vs hydra:** Better — we use the official driver which handles SCRAM-SHA-1/256 automatically. Hydra does raw SCRAM.
- **Action:** None

### LDAP ⚠️
- **Library:** github.com/go-ldap/ldap/v3 (standard Go LDAP lib)
- **vs hydra:** Equivalent for simple bind. Hydra also attempts CRAM-MD5 and DIGEST-MD5.
- **Note:** Proxy dialer not used — uses net.Dialer directly instead of ProxyAwareDialer.
- **Action:** Wire ProxyAwareDialer through LDAP connection (currently hardcoded net.Dialer).

### SMB ✅
- **Library:** github.com/hirochachacha/go-smb2
- **vs hydra:** We do SMB2/3 NTLM auth. Hydra supports SMB1 too (LM/NTLMv1). SMB1 is extremely rare now.
- **Note:** Domain is hardcoded to empty string. Could accept domain from a flag/credential field.
- **Action:** None (domain support would be a feature request)

### VNC ✅
- **Library:** Custom RFB implementation (pure Go, no external dep)
- **vs hydra:** Equivalent — both do DES challenge-response for RFB 3.3/3.7/3.8. Well implemented.
- **Action:** None

### SNMP ✅
- **Library:** github.com/gosnmp/gosnmp (de facto standard)
- **vs hydra:** We do v2c community string brute. Hydra supports v1/v2c/v3. SNMPv3 with USM would be a good addition but is complex.
- **Action:** None (v3 is a separate feature)

### AMQP ✅
- **Library:** github.com/rabbitmq/amqp091-go (official RabbitMQ)
- **vs hydra:** No hydra module for AMQP — we're ahead.
- **Action:** None

### Asterisk (AMI) ✅
- **Library:** Raw TCP (correct — AMI is simple line protocol)
- **vs hydra:** Hydra's hydra-asterisk.c does the same: banner → Action:Login → parse Response.
- **Action:** None

### Cisco (Telnet) ✅
- **Library:** Raw TCP with shared telnet_util
- **vs hydra:** Equivalent. Both handle Username: or direct Password: prompts.
- **Action:** None

### Cisco Enable ✅
- **Library:** Raw TCP
- **vs hydra:** Same approach — login first, then `enable`, send secret.
- **Note:** Login password hardcoded to `credential.Username` (line: `Fprintf(conn, "%s\r\n", credential.Username)`). This is a design choice — uses username as login password. Should be documented.
- **Action:** None (document the behavior)

### ClickHouse ✅
- **Library:** github.com/ClickHouse/clickhouse-go/v2 (official)
- **vs hydra:** No hydra module — we're ahead.
- **Action:** None

### Cobalt Strike ✅
- **Library:** Custom raw protocol (pure Go)
- **vs hydra:** Matches hydra-cobaltstrike.c exactly. Same magic bytes, same packet layout.
- **Action:** None

### etcd ✅
- **Library:** go.etcd.io/etcd/client/v3 (official)
- **vs hydra:** No hydra module — we're ahead.
- **Action:** None

### HTTP Basic ✅
- **Library:** net/http (stdlib)
- **vs hydra:** Equivalent for Basic auth. Hydra also does digest auth and form-based. Those are separate modules (we don't have http-form yet).
- **Action:** None

### rexec ✅
- **Library:** Raw TCP (correct — simple null-delimited protocol)
- **vs hydra:** Matches. Protocol: \0 stderr-port\0 user\0 pass\0 cmd\0, read status byte.
- **Action:** None

### rlogin ✅
- **Library:** Raw TCP
- **vs hydra:** Matches. Note: rlogin is host-based trust, not password-based. Brute-forcing usernames that might be trusted.
- **Action:** None

### rsh ✅
- **Library:** Raw TCP
- **vs hydra:** Matches.
- **Action:** None

### RTSP ✅
- **Library:** Raw TCP (correct — RTSP is HTTP-like, simple enough)
- **vs hydra:** Both do Basic auth. Hydra also supports Digest auth for RTSP.
- **Action:** Consider adding Digest auth (many IP cameras use it).

### SMPP ✅
- **Library:** github.com/linxGnu/gosmpp (maintained)
- **vs hydra:** No hydra module — we're ahead. Good error classification with ESME codes.
- **Action:** None

### TeamSpeak ✅
- **Library:** Raw TCP (correct — ServerQuery is line protocol)
- **vs hydra:** Same approach. `login user pass` → parse `error id=`.
- **Action:** None

### Telnet ✅
- **Library:** Raw TCP with shared telnet_util
- **vs hydra:** Both handle login:/password: prompts. Hydra has more prompt patterns but ours covers the common ones.
- **Action:** None

### Vault ✅
- **Library:** net/http (correct — Vault uses HTTP API)
- **vs hydra:** No hydra module — we're ahead.
- **Note:** Uses deprecated `ioutil.ReadAll` — should be `io.ReadAll`.
- **Action:** Replace `ioutil.ReadAll` with `io.ReadAll` (minor cleanup).

### XMPP ⚠️
- **Library:** github.com/xmppo/go-xmpp
- **vs hydra:** Both do SASL auth. Our implementation doesn't use ProxyAwareDialer (hardcoded dialer in xmpp.Options).
- **Action:** Wire ProxyAwareDialer if possible (may need library support).

### SOCKS5 ✅ (not in output above but exists)
- **Library:** Raw TCP
- **vs hydra:** Both do RFC 1928 username/password auth (method 0x02).
- **Action:** None

---

## Priority Issues

### Must Fix
1. **IMAP:** Escape `"` and `\` in LOGIN credentials
2. **Vault:** Replace deprecated `ioutil.ReadAll`

### Should Fix
3. **SMTP:** Add LOGIN auth method fallback (many Exchange servers need it)
4. **LDAP:** Wire ProxyAwareDialer (currently bypasses proxy)
5. **RTSP:** Add Digest auth (common on IP cameras)

### Nice to Have
6. **XMPP:** Wire ProxyAwareDialer
7. **IMAP:** Add PLAIN auth method
8. **SMB:** Support domain parameter

## Modules Where We're Ahead of hydra
- AMQP, ClickHouse, Cobalt Strike, etcd, SMPP, Vault — no hydra equivalents
- MongoDB — using official driver with auto SCRAM negotiation
- Redis — ACL username support

## Go Libraries Not Currently Used But Worth Noting
- `emersion/go-sasl` — for adding LOGIN/CRAM-MD5 to mail protocols
- `github.com/Azure/go-ntlmssp` — for NTLM auth (SMTP, IMAP, HTTP)
- `github.com/masterzen/winrm` — for WinRM module (task #121)
