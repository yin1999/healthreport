package config

import (
	"errors"
	"flag"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrOutOfRange number out of range
	ErrOutOfRange = errors.New("number: out of range")
	// ErrWrongFormat time format is not correct
	ErrWrongFormat = errors.New("time: wrong format")
)

// Time time config for scheduler
type Time struct {
	Hour   int
	Minute int
}

// Config config struct
type Config struct {
	MaxAttempts uint8 `json:"maxAttempts"`
	PunchTime   Time  `json:"punchTime"`
}

// Printer interface
type Printer interface {
	Printf(format string, v ...interface{})
}

// SetFlag load config from args
func (cfg *Config) SetFlag(flag *flag.FlagSet) {
	if field := reflect.ValueOf(cfg.PunchTime); field.IsZero() {
		now := time.Now()
		cfg.PunchTime.Hour = now.Hour()
		cfg.PunchTime.Minute = now.Minute()
	}
	flag.Func("t", "set punch time(default: now)", func(s string) error {
		return cfg.PunchTime.parse(s)
	})
	if cfg.MaxAttempts == 0 {
		cfg.MaxAttempts = 16
	}
	flag.Func("c", "set maximum retry attempts when punch failed(default: 16)", func(s string) error {
		return parseAttempts(&cfg.MaxAttempts, s)
	})
}

// Check check config
//
// Deprecated: use SetFlag to initialize config
func (cfg Config) Check() error {
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
func (cfg Config) Show(logger Printer) {
	logger.Printf("Maximum number of attempts: %d\n", cfg.MaxAttempts)
	logger.Printf("Time set: %02d:%02d\n", cfg.PunchTime.Hour, cfg.PunchTime.Minute)
}

func parseAttempts(t *uint8, text string) (err error) {
	var n uint64
	n, err = strconv.ParseUint(text, 10, 8)
	if err != nil {
		return
	}
	if n == 0 || n > 120 {
		return ErrOutOfRange
	}
	*t = uint8(n)
	return
}

func (t *Time) parse(text string) error {
	index := strings.IndexByte(text, ':')
	if index <= 0 {
		return ErrWrongFormat
	}

	hour, err := strconv.Atoi(text[:index])
	if err != nil || hour < 0 || hour >= 24 {
		return ErrWrongFormat
	}

	var minute int

	minute, err = strconv.Atoi(text[index+1:])
	if err != nil || minute < 0 || minute >= 60 {
		return ErrWrongFormat
	}

	t.Hour = hour
	t.Minute = minute
	return err
}
