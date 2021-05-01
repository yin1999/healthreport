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

type selfTime struct {
	hour   int
	minute int
}

// Config config struct
type Config struct {
	MaxNumberOfAttempts Attempts `json:"maxNumberOfAttempts"`
	PunchTime           selfTime `json:"punchTime"`
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
	return json.Unmarshal(data, cfg)
}

// Time get punch time
func (cfg *Config) Time() (hour, minute int) {
	return cfg.PunchTime.hour, cfg.PunchTime.minute
}

// Check check config
func (cfg *Config) Check() error {
	if cfg.MaxNumberOfAttempts <= 0 || cfg.MaxNumberOfAttempts > 120 {
		return ErrOutOfRange
	}
	if cfg.PunchTime.hour < 0 ||
		cfg.PunchTime.hour >= 24 ||
		cfg.PunchTime.minute < 0 ||
		cfg.PunchTime.minute >= 60 {
		return ErrWrongFormat
	}

	return nil
}

// Show return configuration
func (cfg *Config) Show(logger Printer) {
	logger.Printf("Maximum number of attempts: %d\n", cfg.MaxNumberOfAttempts)
	logger.Printf("Time set: %02d:%02d\n", cfg.PunchTime.hour, cfg.PunchTime.minute)
}

// GetFromStdin 从Stdin获取配置信息
func (cfg *Config) GetFromStdin() {
	var (
		inputString string
		err         error
		n           int
	)

	fmt.Print("请输入每天最大尝试打卡的次数，默认为\"36\"\n")
	for {
		fmt.Print("请输入(1~120):")
		if n, err = fmt.Scanln(&inputString); err == io.EOF {
			return
		}

		if n == 0 {
			cfg.MaxNumberOfAttempts = 36
			break
		} else {
			n, _ := strconv.Atoi(inputString)
			if n > 0 && n <= 120 {
				cfg.MaxNumberOfAttempts = Attempts(n)
				break
			}
		}
	}

	fmt.Print("请输入每天定时运行的时间，默认为当前时间\n")
	for {
		fmt.Print("时间(HH:MM, 00:00-23:59):")
		if n, err = fmt.Scanln(&inputString); err == io.EOF {
			return
		}

		if n == 0 {
			timeNow := time.Now()
			cfg.PunchTime.hour = timeNow.Hour()
			cfg.PunchTime.minute = timeNow.Minute()
		} else {
			st := selfTime{}
			if err = st.UnmarshalText([]byte(inputString)); err != nil {
				continue
			} else {
				cfg.PunchTime = st
			}
		}

		if err = cfg.Check(); err == nil {
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
func (t selfTime) MarshalText() (data []byte, err error) {
	data = []byte(fmt.Sprintf("%02d:%02d", t.hour, t.minute))
	return
}

// UnmarshalText interface of json.Unmarshal
func (t *selfTime) UnmarshalText(text []byte) error {
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

	t.hour = hour
	t.minute = minute

	return err
}
