package validate

import (
	"context"
	"encoding/binary"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func TestCheckUsesBoundedConcurrency(t *testing.T) {
	var current int32
	var maximum int32

	results, err := Check(context.Background(), []string{"a", "b", "c", "d", "e"}, Options{
		ProbeURL:    "http://example.test",
		Timeout:     time.Second,
		Concurrency: 2,
		Probe: func(ctx context.Context, address, probeURL string, timeout time.Duration) Result {
			active := atomic.AddInt32(&current, 1)
			for {
				seen := atomic.LoadInt32(&maximum)
				if active <= seen || atomic.CompareAndSwapInt32(&maximum, seen, active) {
					break
				}
			}
			time.Sleep(40 * time.Millisecond)
			atomic.AddInt32(&current, -1)
			return Result{Address: address, Reachable: true}
		},
	})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	if got, want := len(results), 5; got != want {
		t.Fatalf("len(results) = %d, want %d", got, want)
	}
	if got := atomic.LoadInt32(&maximum); got > 2 {
		t.Fatalf("maximum concurrency = %d, want <= 2", got)
	}
}

func TestCheckProbesThroughSocks5Proxy(t *testing.T) {
	probeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer probeServer.Close()

	proxyAddress, shutdown := startSOCKS5Server(t)
	defer shutdown()

	results, err := Check(context.Background(), []string{proxyAddress}, Options{
		ProbeURL:    probeServer.URL,
		Timeout:     2 * time.Second,
		Concurrency: 1,
	})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if !results[0].Reachable {
		t.Fatalf("Reachable = false, error = %q", results[0].Error)
	}
	if results[0].StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", results[0].StatusCode, http.StatusOK)
	}
}

func startSOCKS5Server(t *testing.T) (string, func()) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen failed: %v", err)
	}

	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go handleSOCKS5Conn(conn)
		}
	}()

	return listener.Addr().String(), func() {
		_ = listener.Close()
		<-stopped
	}
}

func handleSOCKS5Conn(client net.Conn) {
	defer client.Close()

	header := make([]byte, 2)
	if _, err := io.ReadFull(client, header); err != nil {
		return
	}
	methods := make([]byte, int(header[1]))
	if _, err := io.ReadFull(client, methods); err != nil {
		return
	}
	if _, err := client.Write([]byte{0x05, 0x00}); err != nil {
		return
	}

	requestHeader := make([]byte, 4)
	if _, err := io.ReadFull(client, requestHeader); err != nil {
		return
	}
	if requestHeader[1] != 0x01 {
		_, _ = client.Write([]byte{0x05, 0x07, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}

	target, err := readSOCKS5Address(client, requestHeader[3])
	if err != nil {
		return
	}

	upstream, err := net.Dial("tcp", target)
	if err != nil {
		_, _ = client.Write([]byte{0x05, 0x05, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}
	defer upstream.Close()

	if _, err := client.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0}); err != nil {
		return
	}

	go func() {
		_, _ = io.Copy(upstream, client)
		_ = upstream.Close()
	}()
	_, _ = io.Copy(client, upstream)
}

func readSOCKS5Address(conn net.Conn, atyp byte) (string, error) {
	var host string
	switch atyp {
	case 0x01:
		addr := make([]byte, 4)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return "", err
		}
		host = net.IP(addr).String()
	case 0x03:
		length := make([]byte, 1)
		if _, err := io.ReadFull(conn, length); err != nil {
			return "", err
		}
		name := make([]byte, int(length[0]))
		if _, err := io.ReadFull(conn, name); err != nil {
			return "", err
		}
		host = string(name)
	case 0x04:
		addr := make([]byte, 16)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return "", err
		}
		host = net.IP(addr).String()
	default:
		return "", io.ErrUnexpectedEOF
	}

	portBytes := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBytes); err != nil {
		return "", err
	}
	port := binary.BigEndian.Uint16(portBytes)
	return net.JoinHostPort(host, strconv.Itoa(int(port))), nil
}
