package email

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strings"
)

var (
	// ErrNotSupportAuth auth failed error
	ErrNotSupportAuth = errors.New("smtp: server doesn't support AUTH")
	// ErrNoReceiver reciver is empty error
	ErrNoReceiver = errors.New("mail: no receiver")
)

type emailAccount struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	TLS      bool   `json:"TLS"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Config smtp config
type Config struct {
	To   []string     `json:"to"`
	SMTP emailAccount `json:"SMTP"`
}

// LoginTest return nil, expect cannot login to the server
func (config *emailAccount) LoginTest() error {
	a := smtp.PlainAuth("",
		config.Username,
		config.Password,
		config.Host)

	c, err := newClient(config.Host, config.Port, config.TLS)
	if err != nil {
		return err
	}

	if ok, _ := c.Extension("AUTH"); !ok {
		return ErrNotSupportAuth
	}

	if err = c.Auth(a); err != nil {
		return err
	}
	c.Quit()
	return err
}

// SendMail send mail on STARTTLS/TLS port
func (config *Config) SendMail(nickName, subject, body string) error {
	if len(config.To) == 0 {
		return ErrNoReceiver
	}
	header := make(map[string]string)
	header["From"] = nickName + "<" + config.SMTP.Username + ">"
	header["To"] = strings.Join(config.To, ";")
	header["Subject"] = subject
	header["Content-Type"] = "text/html; charset=UTF-8"
	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body
	auth := smtp.PlainAuth(
		"",
		config.SMTP.Username,
		config.SMTP.Password,
		config.SMTP.Host,
	)
	client, err := newClient(config.SMTP.Host, config.SMTP.Port, config.SMTP.TLS)
	if err != nil {
		return err
	}
	return sendMail(client,
		auth,
		config.SMTP.Username,
		config.To,
		[]byte(message))
}

// LoadConfig if Email config exists, return email config
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	if err = json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, err
}

func newClient(host string, port int, TLS bool) (client *smtp.Client, err error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	var conn net.Conn
	if TLS {
		conn, err = tls.Dial("tcp",
			addr,
			&tls.Config{ServerName: host},
		)
	} else {
		conn, err = net.Dial("tcp", addr)
	}
	if err != nil {
		return
	}

	client, err = smtp.NewClient(conn, host)
	if err != nil {
		client.Close()
		return
	}
	if TLS {
		if err = client.Hello("localhost"); err != nil {
			client.Close()
			client = nil
		}
		return
	}
	if ok, _ := client.Extension("STARTTLS"); ok {
		config := &tls.Config{ServerName: host}
		err = client.StartTLS(config)
	}
	if err != nil {
		client.Close()
		client = nil
	}
	return
}

func sendMail(c *smtp.Client, a smtp.Auth, from string, to []string, msg []byte) error {
	var err error
	defer c.Close()
	if err = validateLine(from); err != nil {
		return err
	}
	for _, recp := range to {
		if err = validateLine(recp); err != nil {
			return err
		}
	}

	if a != nil {
		if ok, _ := c.Extension("AUTH"); !ok {
			return ErrNotSupportAuth
		}
		if err = c.Auth(a); err != nil {
			return err
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}

func validateLine(line string) error {
	if strings.ContainsAny(line, "\n\r") {
		return errors.New("smtp: A line must not contain CR or LF")
	}
	return nil
}
