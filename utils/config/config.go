package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var (
	// ErrOutOfRange number out of range
	ErrOutOfRange = errors.New("number: out of range")
	// ErrWrongFormat time format is not correct
	ErrWrongFormat = errors.New("time: wrong format")
)

// Attempts 尝试次数
type Attempts uint8

type SelfTime struct {
	Hour   int
	Minute int
}

// Config config struct
type Config struct {
	MaxAttempts Attempts `json:"maxAttempts"`
	PunchTime   SelfTime `json:"punchTime"`
}

// Printer interface
type Printer interface {
	Printf(format string, v ...interface{})
}

// Store write config to file
func (cfg *Config) Store(path string) error {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0744)
	if err != nil {
		return err
	}

	var data []byte
	data, err = json.MarshalIndent(cfg, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// Load read config from file
func (cfg *Config) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, cfg); err != nil {
		return err
	}
	return cfg.Check()
}

// Check check config
func (cfg *Config) Check() error {
	if cfg.MaxAttempts <= 0 || cfg.MaxAttempts > 120 {
		return ErrOutOfRange
	}
	if cfg.PunchTime.Hour < 0 ||
		cfg.PunchTime.Hour >= 24 ||
		cfg.PunchTime.Minute < 0 ||
		cfg.PunchTime.Minute >= 60 {
		return ErrWrongFormat
	}
	return nil
}

// Show return configuration
func (cfg *Config) Show(logger Printer) {
	logger.Printf("Maximum number of attempts: %d\n", cfg.MaxAttempts)
	logger.Printf("Time set: %02d:%02d\n", cfg.PunchTime.Hour, cfg.PunchTime.Minute)
}

// GetFromStdin 从Stdin获取配置信息
func (cfg *Config) GetFromStdin() {
	var (
		inputString string
		err         error
		n           int
	)

	fmt.Print("请输入每天最大尝试打卡的次数，默认为\"16\"\n")
	for n <= 0 || n > 120 {
		fmt.Print("请输入(1~120):")
		if n, err = fmt.Scanln(&inputString); err == io.EOF {
			return
		}

		if n == 0 {
			n = 16
		} else {
			n, _ = strconv.Atoi(inputString)
		}
	}
	cfg.MaxAttempts = Attempts(n)

	fmt.Print("请输入每天定时运行的时间，默认为当前时间\n")
	for {
		fmt.Print("时间(HH:MM, 00:00-23:59):")
		if n, err = fmt.Scanln(&inputString); err == io.EOF {
			return
		}

		if n == 0 {
			timeNow := time.Now()
			cfg.PunchTime.Hour = timeNow.Hour()
			cfg.PunchTime.Minute = timeNow.Minute()
			break
		}
		if err = cfg.PunchTime.UnmarshalText([]byte(inputString)); err == nil {
			break
		}
	}
}

// MarshalJSON interface of json.Marshal
func (t Attempts) MarshalJSON() (data []byte, err error) {
	data = []byte(strconv.Itoa(int(t)))
	return
}

// UnmarshalJSON interface of json.Unmarshal
func (t *Attempts) UnmarshalJSON(text []byte) (err error) {
	var n int
	n, err = strconv.Atoi(string(text))
	if err != nil {
		return
	}
	if n < 1 || n > 120 {
		return ErrOutOfRange
	}
	*t = Attempts(n)
	return
}

// MarshalText interface of json.Marshal
func (t SelfTime) MarshalText() (data []byte, err error) {
	data = []byte(fmt.Sprintf("%02d:%02d", t.Hour, t.Minute))
	return
}

// UnmarshalText interface of json.Unmarshal
func (t *SelfTime) UnmarshalText(text []byte) error {
	index := bytes.IndexByte(text, ':')
	if index <= 0 {
		return ErrWrongFormat
	}

	s := string(text)
	hour, err := strconv.Atoi(s[:index])
	if err != nil || hour < 0 || hour >= 24 {
		return ErrWrongFormat
	}

	var minute int

	minute, err = strconv.Atoi(s[index+1:])
	if err != nil || minute < 0 || minute >= 60 {
		return ErrWrongFormat
	}

	t.Hour = hour
	t.Minute = minute
	return err
}
