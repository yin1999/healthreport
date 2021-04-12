package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	client "github.com/yin1999/healthreport/httpclient"
	"github.com/yin1999/healthreport/utils"
	"golang.org/x/term"
)

// build info
var (
	BuildTime       = "Not Provided."
	ProgramCommitID = "Not Provided."
	ProgramVersion  = "Not Provided."
)

const (
	accountFilename = "account" // 账户信息存储文件名

	configPath = "config.json" // 配置文件

	mailNickName   = "打卡状态推送"
	mailConfigPath = "email.json"

	logFilenameLayout = "2006-01.log" // using time layout
	logPath           = "log"         // 日志存储目录

	retryDelay = 5 // minute

	punchTimeout = 30 // second

	punchStart    = "Start punch\n"
	punchFinish   = "Punch finished\n"
	contextCancel = "Context canceled\n"
)

var (
	cfg      = &utils.Config{} // config
	emailCfg *utils.EmailConfig
	logger   *utils.Logger                   // multiLogger
	cstZone  = time.FixedZone("CST", 8*3600) // China Standard Time Zone
)

func main() {
	var (
		err     error
		account [2]string // 账户信息
	)

	logger, err = utils.NewLogger(logPath, logFilenameLayout)

	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}

	defer logger.Close()
	defer logger.Print("Exit\n")

	logger.Print("Start program\n")

	if err = utils.JSONReader(cfg, configPath); err == nil {
		logger.Printf("Got config from file: '%s'\n", configPath)
	} else {
		cfg.GetFromStdin()
		if err = utils.JSONWriter(*cfg, configPath); err != nil {
			logger.Printf("Cannot save config, err: %s\n", err.Error())
		}
	}

	for _, line := range cfg.Show() {
		logger.Println(line)
	}

	emailCfg, err = utils.LoadEmailConfig(mailConfigPath)
	if err == nil {
		logger.Print("Email deliver enabled\n")
	}

	if err = utils.DataLoad(&account, accountFilename); err == nil {
		logger.Printf("Got account info from file '%s'\n", accountFilename)
	} else {
		if account, err = getAccount(); err != nil {
			logger.Printf("Err: %s\n", err.Error())
			os.Exit(1)
		}

		logger.Print("Got account info from 'Stdin'\n")

		if err = utils.DataStore(account, accountFilename); err != nil {
			logger.Printf("Cannot save account info, err: %s\n", err.Error())
		}
	}

	fmt.Print("Ctrl+C可退出程序\n")

	ctx, cancel := context.WithCancel(context.Background())

	go signalListener(ctx, cancel)

	logger.Print("正在验证账号密码\n")
	err = client.LoginConfirm(ctx, account, punchTimeout*time.Second)
	switch err {
	case nil:
		break
	case context.Canceled:
		logger.Print(contextCancel)
		return
	default:
		logger.Printf("验证密码失败，请检查网络连接、账号密码(Err: %s)\n", err.Error())
		return
	}
	logger.Print("账号密码验证成功，将在5秒后开始打卡\n")

	select {
	case <-time.After(5 * time.Second):
		punchRoutine(ctx, account) // 当天打卡
	case <-ctx.Done():
		break
	}

	punchServe(ctx, account)
}

func init() {
	if len(os.Args) >= 2 {
		var (
			returnCode = 0
			version    bool
			checkEmail bool
		)

		flag.BoolVar(&version, "v", false, "show version and exit")
		flag.BoolVar(&checkEmail, "c", false, "check email")
		flag.Parse()

		if version {
			fmt.Printf("Program Version:        %s\n", ProgramVersion)
			fmt.Printf("Go Version:             %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
			fmt.Printf("Build Time:             %s\n", BuildTime)
			fmt.Printf("Program Commit ID:      %s\n", ProgramCommitID)
		}

		if checkEmail {
			cfg, err := utils.LoadEmailConfig(mailConfigPath)
			if err == nil {
				err = cfg.SMTP.LoginTest()
			}

			if err != nil {
				fmt.Fprintf(os.Stderr, "email check: failed, err: %s\n", err.Error())
				returnCode = 1
			} else {
				fmt.Print("email check: pass\n")
			}
		}

		os.Exit(returnCode)
	}

	client.SetTimeZone(cstZone)
}

func getAccount() (account [2]string, err error) {
	for len(account[0]) == 0 && err != io.EOF { // avoid expect new line error
		fmt.Print("输入用户名:")
		_, err = fmt.Scanln(&account[0])
	}

	var passwd []byte

	for len(passwd) == 0 && err == nil {
		fmt.Print("输入密码:")
		passwd, err = term.ReadPassword(int(syscall.Stdin))
		fmt.Print("\n") // print in new line
	}
	account[1] = string(passwd)
	return
}

func punchServe(ctx context.Context, account [2]string) {
	if ctx.Err() != nil {
		return
	}

	logger.Print("Pausing...")
	select {
	case <-pause(cfg.Time()):
		break
	case <-ctx.Done():
		return
	}

	ticker := time.NewTicker(24 * time.Hour)

	logger.Print("Punch on a 24-hour cycle\n")

	go punchRoutine(ctx, account)

	rand.Seed(time.Now().Unix())

	for {
		select {
		case <-ticker.C:
			select {
			case <-time.After(time.Duration(rand.Intn(int(time.Minute) * 10))):
				break
			case <-ctx.Done():
				return
			}

			go punchRoutine(ctx, account)
		case <-ctx.Done():
			return
		}
	}
}

func pause(hour, minute int) <-chan time.Time { // 暂停到第二天的指定时刻
	year, month, day := time.Now().In(cstZone).Date()
	t := time.Date(year, month, day+1, hour, minute, 0, 0, cstZone)
	return time.After(time.Until(t))
}

// punchRoutine please call this function with go routine
func punchRoutine(ctx context.Context, account [2]string) {
	logger.Print("Start punch routine\n")
	var punchCount utils.Attempts = 1
	var err error
	logger.Print(punchStart)
	err = client.Punch(ctx, account, punchTimeout*time.Second)

	switch err {
	case nil:
		logger.Print(punchFinish)
		return
	case context.Canceled:
		logger.Print(contextCancel)
		return
	}

	ticker := time.NewTicker(retryDelay * time.Minute)
	for punchCount < cfg.MaxNumberOfAttempts {
		logger.Printf("Tried %d times. Retry after %d minutes\n", punchCount, retryDelay)
		select {
		case <-ticker.C: // try again after $retryDelay minutes.
			logger.Print("Start punch\n")
			err = client.Punch(ctx, account, punchTimeout*time.Second)
			switch err {
			case nil:
				logger.Print(punchFinish)
				return
			case context.Canceled:
				logger.Print(contextCancel)
				return
			}

		case <-ctx.Done():
			return
		}
		punchCount++
	}
	if emailCfg != nil {
		e := emailCfg.SendMail(mailNickName,
			fmt.Sprintf("打卡状态推送-%s", time.Now().In(cstZone).Format("2006-01-02")),
			fmt.Sprintf("账户：%s 打卡失败<br>error: %s", account[0], err.Error()))
		if e != nil {
			logger.Printf("Send mail failed, err: %s\n", e.Error())
		}
	}
	logger.Fatalf("Maximum number of attempts reached: %d times. The error of the last time is: %v\n", punchCount, err)
}

func signalListener(ctx context.Context, cancel context.CancelFunc) {
	if ctx == nil || cancel == nil {
		panic("ctx or cancel is nil")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	select {
	case s := <-c:
		logger.Println("Got signal:", s)
		cancel()
	case <-ctx.Done():
		return
	}
}
