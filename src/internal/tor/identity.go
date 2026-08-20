package tor

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

const DefaultCookieAuthFile = "/var/run/tor/control.authcookie"
const DefaultControlTimeout = 10 * time.Second

var (
	ErrAnonSurfNotRunning  = errors.New("AnonSurf is not running")
	ErrControlPortNotFound = errors.New("Tor ControlPort not found")
)

type ControlConfig struct {
	Address        string
	CookieAuthFile string
}

type ControlDialer func(network, address string) (net.Conn, error)

type IdentityChanger struct {
	TorrcPath         string
	ReadFile          func(path string) ([]byte, error)
	Dial              ControlDialer
	IsAnonSurfRunning func() bool
	Timeout           time.Duration
}

func NewIdentityChanger() IdentityChanger {
	return IdentityChanger{
		TorrcPath:         TorrcPath,
		ReadFile:          os.ReadFile,
		IsAnonSurfRunning: IsAnonsurfRunning,
		Timeout:           DefaultControlTimeout,
	}
}

func ParseControlConfig(reader io.Reader) (ControlConfig, error) {
	config := ControlConfig{CookieAuthFile: DefaultCookieAuthFile}
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		fields := strings.Fields(stripTorrcComment(scanner.Text()))
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "ControlPort":
			config.Address = normalizeControlAddress(fields[1])
		case "CookieAuthFile":
			config.CookieAuthFile = fields[1]
		}
	}
	if err := scanner.Err(); err != nil {
		return config, err
	}
	if config.Address == "" {
		return config, ErrControlPortNotFound
	}
	return config, nil
}

func (c IdentityChanger) Change() error {
	if c.IsAnonSurfRunning != nil && !c.IsAnonSurfRunning() {
		return ErrAnonSurfNotRunning
	}

	readFile := c.ReadFile
	if readFile == nil {
		readFile = os.ReadFile
	}
	torrcPath := c.TorrcPath
	if torrcPath == "" {
		torrcPath = TorrcPath
	}

	torrcData, err := readFile(torrcPath)
	if err != nil {
		return fmt.Errorf("read torrc: %w", err)
	}
	config, err := ParseControlConfig(strings.NewReader(string(torrcData)))
	if err != nil {
		return fmt.Errorf("parse torrc control config: %w", err)
	}

	cookieData, err := readFile(config.CookieAuthFile)
	if err != nil {
		return fmt.Errorf("read Tor control cookie: %w", err)
	}

	dial := c.Dial
	timeout := c.Timeout
	if timeout <= 0 {
		timeout = DefaultControlTimeout
	}
	if dial == nil {
		dial = func(network, address string) (net.Conn, error) {
			return net.DialTimeout(network, address, timeout)
		}
	}
	conn, err := dial("tcp", config.Address)
	if err != nil {
		return fmt.Errorf("connect Tor control port: %w", err)
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return fmt.Errorf("set Tor control deadline: %w", err)
	}

	return requestNewIdentity(conn, hex.EncodeToString(cookieData))
}

func requestNewIdentity(conn net.Conn, cookieHex string) error {
	reader := bufio.NewReader(conn)
	if _, err := fmt.Fprintf(conn, "AUTHENTICATE %s\r\n", cookieHex); err != nil {
		return fmt.Errorf("authenticate with Tor control port: %w", err)
	}
	if err := readControlOK(reader, "authenticate"); err != nil {
		return err
	}

	if _, err := fmt.Fprint(conn, "SIGNAL NEWNYM\r\n"); err != nil {
		return fmt.Errorf("request new Tor identity: %w", err)
	}
	if err := readControlOK(reader, "signal newnym"); err != nil {
		return err
	}

	_, _ = fmt.Fprint(conn, "QUIT\r\n")
	return nil
}

func readControlOK(reader *bufio.Reader, context string) error {
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("%s response: %w", context, err)
		}
		line = strings.TrimRight(line, "\r\n")
		if len(line) < 4 {
			return fmt.Errorf("%s response: malformed Tor control response %q", context, line)
		}
		if !strings.HasPrefix(line, "250") {
			return fmt.Errorf("%s response: %s", context, line)
		}
		if line[3] == ' ' {
			return nil
		}
	}
}

func normalizeControlAddress(value string) string {
	if strings.Contains(value, ":") {
		return value
	}
	return "127.0.0.1:" + value
}

func stripTorrcComment(line string) string {
	if index := strings.Index(line, "#"); index >= 0 {
		return line[:index]
	}
	return line
}
