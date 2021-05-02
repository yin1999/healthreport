package log

import (
	"errors"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

const (
	// DefaultLayout using to name log file
	DefaultLayout = "2006-01.log"
)

// ErrInvalidSymbol shows that the layout contains invalid symbol
var ErrInvalidSymbol = errors.New("log: layout contains invalid symbol")

// Logger struct for log
type Logger struct {
	*log.Logger
	file *os.File
	done chan struct{}
}

// New create a new logger
func New(dir, layout string) (*Logger, error) {
	// check
	if strings.ContainsRune(layout, '\\') || strings.ContainsRune(layout, '/') || strings.ContainsRune(layout, ':') {
		return nil, ErrInvalidSymbol
	}

	// init config
	if len(dir) != 0 && !os.IsPathSeparator(dir[len(dir)-1]) {
		dir = string(append([]byte(dir), os.PathSeparator))
	}

	w, f, err := newWriter(dir, layout)
	if err != nil {
		return nil, err
	}
	logger := &Logger{
		Logger: log.New(w, "", log.LstdFlags),
		file:   f,
		done:   make(chan struct{}, 1),
	}

	go logger.serve(dir, layout)
	return logger, nil
}

// Close close logMaintainer and log file,
// set (*log.Logger)'s output to nil
func (l *Logger) Close() error {
	close(l.done)
	l.Logger.SetOutput(nil)
	return l.file.Close()
}

func newWriter(dir, layout string) (w io.Writer, file *os.File, err error) {
	fileName := dir + time.Now().Format(layout)
	if err = os.MkdirAll(dir, 0755); err != nil {
		return
	}
	file, err = os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	w = io.MultiWriter(os.Stderr, file)
	return
}

func (l *Logger) serve(dir, layout string) {
	var (
		year  int
		month time.Month
		next  time.Time

		w    io.Writer
		file *os.File
		err  error
	)

	for {
		year, month, _ = time.Now().Date()
		month++
		next = time.Date(year, month, 1, 0, 0, 0, 0, time.Local)

		select {
		case <-time.After(time.Until(next)): // 暂停到下个月创建日志文件
			w, file, err = newWriter(dir, layout)
			if err != nil {
				if l.Logger != nil {
					l.Printf("Create new log file failed: %s\n", err.Error())
				} else {
					log.Printf("Create new log file failed: %s\n", err.Error())
				}
				continue
			}

			l.Logger.SetOutput(w)
			l.file.Close()
			l.file = file
		case <-l.done:
			return
		}
	}
}
