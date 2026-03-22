package utils

import (
	"bytes"
	"errors"
	"io"
	"net"
	"testing"
	"time"
)

// mockReader delivers data in configurable chunks for testing ReadFull.
type mockReader struct {
	data  []byte
	pos   int
	chunk int // max bytes per Read call; 0 = all at once
}

func (m *mockReader) Read(buf []byte) (int, error) {
	if m.pos >= len(m.data) {
		return 0, io.EOF
	}

	remaining := m.data[m.pos:]
	n := len(remaining)
	if m.chunk > 0 && n > m.chunk {
		n = m.chunk
	}
	if n > len(buf) {
		n = len(buf)
	}

	copy(buf, remaining[:n])
	m.pos += n
	return n, nil
}

// errorAfterReader returns data then an error after N bytes.
type errorAfterReader struct {
	data []byte
	pos  int
	err  error
}

func (m *errorAfterReader) Read(buf []byte) (int, error) {
	if m.pos >= len(m.data) {
		return 0, m.err
	}

	remaining := m.data[m.pos:]
	n := len(remaining)
	if n > len(buf) {
		n = len(buf)
	}
	copy(buf, remaining[:n])
	m.pos += n
	return n, nil
}

// testConn wraps an io.Reader to satisfy net.Conn.
type testConn struct {
	reader io.Reader
}

func (c *testConn) Read(b []byte) (int, error)         { return c.reader.Read(b) }
func (c *testConn) Write(b []byte) (int, error)        { return len(b), nil }
func (c *testConn) Close() error                       { return nil }
func (c *testConn) LocalAddr() net.Addr                { return nil }
func (c *testConn) RemoteAddr() net.Addr               { return nil }
func (c *testConn) SetDeadline(t time.Time) error      { return nil }
func (c *testConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *testConn) SetWriteDeadline(t time.Time) error { return nil }

func TestReadFull_ExactRead(t *testing.T) {
	data := []byte("hello world")
	conn := &testConn{reader: &mockReader{data: data}}
	buf := make([]byte, len(data))
	n, err := ReadFull(conn, buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(data) {
		t.Errorf("got n=%d, want %d", n, len(data))
	}
	if !bytes.Equal(buf, data) {
		t.Errorf("got %q, want %q", buf, data)
	}
}

func TestReadFull_ShortReads(t *testing.T) {
	data := []byte("abcdefghij")
	conn := &testConn{reader: &mockReader{data: data, chunk: 2}}
	buf := make([]byte, len(data))
	n, err := ReadFull(conn, buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(data) {
		t.Errorf("got n=%d, want %d", n, len(data))
	}
	if !bytes.Equal(buf, data) {
		t.Errorf("got %q, want %q", buf, data)
	}
}

func TestReadFull_EmptyBuffer(t *testing.T) {
	conn := &testConn{reader: &mockReader{data: []byte("data")}}
	buf := make([]byte, 0)
	n, err := ReadFull(conn, buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("got n=%d, want 0", n)
	}
}

func TestReadFull_ErrorMidStream(t *testing.T) {
	testErr := errors.New("connection reset")
	data := []byte("abc")
	conn := &testConn{reader: &errorAfterReader{data: data, err: testErr}}
	buf := make([]byte, 10)
	n, err := ReadFull(conn, buf)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if n != len(data) {
		t.Errorf("got n=%d, want %d partial bytes", n, len(data))
	}
}

func TestReadFull_ImmediateEOF(t *testing.T) {
	conn := &testConn{reader: &mockReader{data: []byte{}}}
	buf := make([]byte, 5)
	n, err := ReadFull(conn, buf)
	if err != io.EOF {
		t.Errorf("expected io.EOF, got %v", err)
	}
	if n != 0 {
		t.Errorf("got n=%d, want 0", n)
	}
}

func TestReadFull_SingleByteChunks(t *testing.T) {
	data := []byte("xyz")
	conn := &testConn{reader: &mockReader{data: data, chunk: 1}}
	buf := make([]byte, len(data))
	n, err := ReadFull(conn, buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(data) {
		t.Errorf("got n=%d, want %d", n, len(data))
	}
	if !bytes.Equal(buf, data) {
		t.Errorf("got %q, want %q", buf, data)
	}
}
