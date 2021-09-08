package systemd

import (
	"net"
	"os"
)

const (
	// Ready notify the init system that the program is started up
	Ready = "READY=1"
	// Stopping notify the init system that the program is stopping
	Stopping = "STOPPING=1"
	// Reloading notify the init sysyem that the program is reloading
	Reloading = "RELOADING=1"
)

// Notify notify the init system about status changes
func Notify(state string) error {
	socketAddr := &net.UnixAddr{
		Name: os.Getenv("NOTIFY_SOCKET"),
		Net:  "unixgram",
	}
	if socketAddr.Name == "" {
		return nil
	}

	conn, err := net.DialUnix(socketAddr.Net, nil, socketAddr)
	if err != nil {
		return err
	}
	_, err = conn.Write([]byte(state))
	conn.Close()
	return err
}
