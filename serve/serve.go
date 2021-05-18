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

var (
	punchStart    = "Start punch\n"
	punchFinish   = "Punch finished\n"
	contextCancel = "Context canceled\n"
)

// PunchServe universal punch service
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
	var timer *time.Timer

	for {
		timer = time.NewTimer(time.Until(nextTime) + time.Duration(r.Int63())%time.Minute*10)
		select {
		case <-timer.C:
			go cfg.PunchRoutine(ctx, account)
		case <-ctx.Done():
			timer.Stop()
			return
		}
		nextTime = nextTime.Add(24 * time.Hour)
	}
}

// PunchRoutine please call this function with go routine
func (cfg Config) PunchRoutine(ctx context.Context, account [2]string) {
	cfg.Logger.Print("Start punch routine\n")
	cfg.Logger.Print(punchStart)

	err := cfg.PunchFunc(ctx, account, cfg.Timeout)
	switch err {
	case nil:
		cfg.Logger.Print(punchFinish)
		return
	case context.Canceled:
		cfg.Logger.Print(contextCancel)
		return
	}

	punchCount := uint8(1)
	ticker := time.NewTicker(cfg.RetryAfter)

	for punchCount < cfg.MaxAttempts {
		cfg.Logger.Printf("Tried %d times. Retry after %v\n", punchCount, cfg.RetryAfter)
		select {
		case <-ticker.C: // try again after cfg.RetryAfter.
			cfg.Logger.Print(punchStart)
			err = cfg.PunchFunc(ctx, account, cfg.Timeout)
			switch err {
			case nil:
				ticker.Stop()
				cfg.Logger.Print(punchFinish)
				return
			case context.Canceled:
				ticker.Stop()
				cfg.Logger.Print(contextCancel)
				return
			}

		case <-ctx.Done():
			return
		}
		punchCount++
	}
	if cfg.Sender != nil {
		e := cfg.Sender.Send(cfg.MailNickName,
			fmt.Sprintf("打卡状态推送-%s", time.Now().In(cfg.Time.TimeZone).Format("2006-01-02")),
			fmt.Sprintf("账户：%s 打卡失败<br>error: %s", account[0], err.Error()))
		if e != nil {
			cfg.Logger.Printf("Send mail failed, err: %s\n", e.Error())
		}
	}
	cfg.Logger.Printf("Maximum attempts: %d reached. Last error: %s\n", punchCount, err.Error())
	os.Exit(1)
}
