package utils

import "net"

// ReadFull reads exactly len(buf) bytes from conn.
// It retries short reads until the buffer is full or an error occurs.
func ReadFull(conn net.Conn, buf []byte) (int, error) {
	total := 0
	for total < len(buf) {
		n, err := conn.Read(buf[total:])
		total += n
		if err != nil {
			return total, err
		}
	}
	return total, nil
}
