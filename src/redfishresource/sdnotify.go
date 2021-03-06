package domain

import (
	"errors"
	"io"
	"net"
	"os"
	"strconv"
)

// Sdnotify sends a specified string to the systemd notification socket.
type Sdnotify struct {
	conn io.WriteCloser
}

type Simulated struct{}

func SimulateSdnotify() *Simulated {
	return &Simulated{}
}
func (s *Simulated) Close() {}
func (s *Simulated) Notify(state string) error {
	//fmt.Printf("SD_NOTIFY not available, simulating: %s\n", state)
	return nil
}
func (s *Simulated) GetIntervalUsec() int {
	return GetIntervalUsec()
}

func GetIntervalUsec() int {
	intervalStr := os.Getenv("WATCHDOG_USEC")
	interval := 0
	if intervalStr != "" {
		i, err := strconv.Atoi(intervalStr)
		if err == nil {
			interval = i
		}
	}
	return interval
}

type closeNotifier interface {
	GetIntervalUsec() int
	Notify(string) error
	Close()
}

func NewSdnotify() (*Sdnotify, error) {

	name := os.Getenv("NOTIFY_SOCKET")
	if name == "" {
		return nil, errors.New("NOTIFY_SOCKET environment variable not set")
	}

	conn, err := net.DialUnix("unixgram", nil, &net.UnixAddr{Name: name, Net: "unixgram"})
	if err != nil {
		return nil, err
	}

	return &Sdnotify{
		conn: conn,
	}, nil
}

func (s *Sdnotify) GetIntervalUsec() int {
	return GetIntervalUsec()
}

func (s *Sdnotify) Notify(state string) error {
	_, err := s.conn.Write([]byte(state))
	return err
}

func (s *Sdnotify) Close() {
	s.conn.Close()
}
