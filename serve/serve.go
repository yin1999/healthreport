package serve

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"
)

// Sender send a message when punch failed
type Sender interface {
	Send(nickName, subject, body string) error
}

// Logger interface for log
type Logger interface {
	Printf(format string, v ...interface{})
	Print(v ...interface{})
}

// Time time info for punch
type Time struct {
	Hour     int
	Minute   int
	TimeZone *time.Location
}

// Config punch information configuration
type Config struct {
	Sender       Sender
	Logger       Logger
	MaxAttempts  uint8
	Time         Time
	MailNickName string
	Timeout      time.Duration
	RetryAfter   time.Duration
	PunchFunc    func(ctx context.Context, account [2]string, timeout time.Duration) error
}

// PunchServe universal punch service.
// When it is called, it will call the punch function immediately,
// and then call the punch function daily.
func (cfg Config) PunchServe(ctx context.Context, account [2]string) {
	if ctx.Err() != nil {
		return
	}

	cfg.Logger.Print("Punch on a 24-hour cycle\n")

	var nextTime time.Time
	{
		year, month, day := time.Now().In(cfg.Time.TimeZone).Date()
		nextTime = time.Date(year, month, day+1, // next day
			cfg.Time.Hour,
			cfg.Time.Minute-5, // rand in [-5, +5) minutes
			0, 0, cfg.Time.TimeZone,
		)
	}

	r := rand.New(rand.NewSource(time.Now().Unix()))

	timer := time.NewTimer(time.Until(nextTime) + time.Duration(r.Int63())%(time.Minute*10))
	for {
		go cfg.PunchRoutine(ctx, account)

		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return
		}
		nextTime = nextTime.Add(24 * time.Hour)
		timer.Reset(time.Until(nextTime) + time.Duration(r.Int63())%(time.Minute*10))
	}
}

// PunchRoutine punch until successed or max attempts reached
func (cfg Config) PunchRoutine(ctx context.Context, account [2]string) {
	cfg.Logger.Print("Start punch routine\n")
	var err error

	var timer *time.Timer
	for punchCount := uint8(1); punchCount <= cfg.MaxAttempts; punchCount++ {
		cfg.Logger.Print("Start punch\n")
		err = cfg.PunchFunc(ctx, account, cfg.Timeout)

		// error handling
		switch err {
		case nil:
			cfg.Logger.Print("Punch finished\n")
			return
		case context.Canceled:
			return
		default:
			cfg.Logger.Printf("Tried %d times. Retry after %v\n", punchCount, cfg.RetryAfter)
		}

		// waiting
		if timer == nil {
			timer = time.NewTimer(cfg.RetryAfter)
		} else {
			timer.Reset(cfg.RetryAfter)
		}
		select {
		case <-timer.C: // try again after cfg.RetryAfter.
		case <-ctx.Done():
			timer.Stop()
			return
		}
	}
	// error handling
	if cfg.Sender != nil {
		e := cfg.Sender.Send(cfg.MailNickName,
			fmt.Sprintf("??????????????????-%s", time.Now().In(cfg.Time.TimeZone).Format("2006-01-02")),
			fmt.Sprintf("?????????%s ????????????(err: %s)", account[0], err.Error()))
		if e != nil {
			cfg.Logger.Printf("Send message failed, err: %s\n", e.Error())
		}
	}
	cfg.Logger.Printf("Maximum attempts: %d reached. Last error: %s\n", cfg.MaxAttempts, err.Error())
	os.Exit(1)
}
