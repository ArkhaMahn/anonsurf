package tor

import (
	"bufio"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testIdentityChanger(t *testing.T, dial ControlDialer) IdentityChanger {
	t.Helper()
	dir := t.TempDir()
	torrcPath := filepath.Join(dir, "torrc")
	cookiePath := filepath.Join(dir, "control.authcookie")
	torrc := "ControlPort 127.0.0.1:9051\nCookieAuthFile " + cookiePath + "\n"
	if err := os.WriteFile(torrcPath, []byte(torrc), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cookiePath, []byte("cookie"), 0600); err != nil {
		t.Fatal(err)
	}
	return IdentityChanger{
		TorrcPath:         torrcPath,
		ReadFile:          os.ReadFile,
		Dial:              dial,
		IsAnonSurfRunning: func() bool { return true },
		Timeout:           50 * time.Millisecond,
	}
}

func TestIdentityChangerRequestsNewIdentity(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close()
	changer := testIdentityChanger(t, func(string, string) (net.Conn, error) {
		return client, nil
	})

	done := make(chan error, 1)
	go func() {
		reader := bufio.NewReader(server)
		for _, command := range []string{"AUTHENTICATE ", "SIGNAL NEWNYM"} {
			line, err := reader.ReadString('\n')
			if err != nil {
				done <- err
				return
			}
			if !strings.HasPrefix(line, command) {
				t.Errorf("command = %q, want prefix %q", line, command)
			}
			if _, err := server.Write([]byte("250 OK\r\n")); err != nil {
				done <- err
				return
			}
		}
		line, err := reader.ReadString('\n')
		if err != nil {
			done <- err
			return
		}
		if line != "QUIT\r\n" {
			t.Errorf("command = %q, want QUIT", line)
		}
		done <- nil
	}()

	if err := changer.Change(); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestIdentityChangerTimesOutControlResponse(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close()
	changer := testIdentityChanger(t, func(string, string) (net.Conn, error) {
		return client, nil
	})

	if err := changer.Change(); err == nil {
		t.Fatal("expected control response timeout")
	}
}
