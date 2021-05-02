package serve

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"
)

type Sender interface {
	Send(nickName, subject, body string) error
}

type Logger interface {
	Printf(format string, v ...interface{})
	Print(v ...interface{})
}

type Time struct {
	Hour     int
	Minute   int
	TimeZone *time.Location
}

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

func (cfg Config) PunchServe(ctx context.Context, account [2]string) {
	if ctx.Err() != nil {
		return
	}

	cfg.Logger.Print("Pausing...\n")
	select {
	case <-pause(cfg.Time.Hour, cfg.Time.Minute, cfg.Time.TimeZone):
		break
	case <-ctx.Done():
		return
	}

	ticker := time.NewTicker(24 * time.Hour)

	cfg.Logger.Print("Punch on a 24-hour cycle\n")

	go cfg.PunchRoutine(ctx, account)

	r := rand.New(rand.NewSource(time.Now().Unix()))

	for {
		select {
		case <-ticker.C:
			select {
			case <-time.After(time.Duration(r.Intn(int(time.Minute) * 10))):
				go cfg.PunchRoutine(ctx, account)
			case <-ctx.Done():
				return
			}
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

// PunchRoutine please call this function with go routine
func (cfg Config) PunchRoutine(ctx context.Context, account [2]string) {
	cfg.Logger.Print("Start punch routine\n")
	var punchCount uint8 = 1
	var err error
	cfg.Logger.Print(punchStart)
	err = cfg.PunchFunc(ctx, account, cfg.Timeout)

	switch err {
	case nil:
		cfg.Logger.Print(punchFinish)
		return
	case context.Canceled:
		cfg.Logger.Print(contextCancel)
		return
	}

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

func pause(hour, minute int, tz *time.Location) <-chan time.Time { // 暂停到第二天的指定时刻
	year, month, day := time.Now().In(tz).Date()
	t := time.Date(year, month, day+1, hour, minute, 0, 0, tz)
	return time.After(time.Until(t))
}
