package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	client "github.com/yin1999/healthreport/httpclient"
	"github.com/yin1999/healthreport/serve"
	"github.com/yin1999/healthreport/utils/config"
	"github.com/yin1999/healthreport/utils/email"
	"github.com/yin1999/healthreport/utils/log"
	"github.com/yin1999/healthreport/utils/object"
	"golang.org/x/term"
)

// build info
var (
	BuildTime       = "Not Provided."
	ProgramCommitID = "Not Provided."
	ProgramVersion  = "Not Provided."
)

const (
	mailNickName    = "打卡状态推送"
	mailConfigPath  = "email.json"
	logPath         = "log"         // 日志存储目录
	configPath      = "config.json" // 配置文件
	accountFilename = "account"     // 账户信息存储文件名

	retryAfter   = 5 * time.Minute
	punchTimeout = 30 * time.Second
)

func main() {
	logger, err := log.New(logPath, log.DefaultLayout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	defer logger.Close()

	defer logger.Print("Exit\n")
	logger.Print("Start program\n")

	cfg := &config.Config{} // config
	if err = cfg.Load(configPath); err == nil {
		logger.Printf("Got config from file: '%s'\n", configPath)
	} else {
		cfg.GetFromStdin()
		if err = cfg.Store(configPath); err != nil {
			logger.Printf("Cannot save config, err: %s\n", err.Error())
		}
	}
	cfg.Show(logger)

	var emailCfg *email.Config
	emailCfg, err = email.LoadConfig(mailConfigPath)
	if err == nil {
		logger.Print("Email deliver enabled\n")
	}

	var account [2]string // 账户信息
	if err = object.Load(&account, accountFilename); err == nil {
		logger.Printf("Got account info from file '%s'\n", accountFilename)
	} else {
		if err = getAccount(&account); err != nil {
			logger.Printf("Err: %s\n", err.Error())
			os.Exit(1)
		}

		logger.Print("Got account info from 'Stdin'\n")

		if err = object.Store(account, accountFilename); err != nil {
			logger.Printf("Cannot save account info, err: %s\n", err.Error())
		}
	}

	fmt.Print("Ctrl+C可退出程序\n")

	ctx, cc := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cc()

	logger.Print("正在验证账号密码\n")
	err = client.LoginConfirm(ctx, account, punchTimeout)
	if err != nil {
		logger.Printf("验证密码失败(Err: %s)\n", err.Error())
		return
	}
	logger.Print("账号密码验证成功，将在5秒后开始打卡\n")

	serveCfg := &serve.Config{
		Sender:      emailCfg,
		Logger:      logger,
		MaxAttempts: uint8(cfg.MaxAttempts),
		Time: serve.Time{
			Hour:     cfg.PunchTime.Hour,
			Minute:   cfg.PunchTime.Minute,
			TimeZone: time.FixedZone("CST", 8*3600), // China Standard Time Zone,
		},
		MailNickName: mailNickName,
		Timeout:      punchTimeout,
		RetryAfter:   retryAfter,
		PunchFunc:    client.Punch,
	}

	{
		timer := time.NewTimer(5 * time.Second)
		select {
		case <-timer.C:
			break
		case <-ctx.Done():
			timer.Stop()
			return
		}
	}

	if err = serveCfg.PunchServe(ctx, account); err != nil && err != context.Canceled {
		logger.Fatalln(err.Error())
	}
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
			cfg, err := email.LoadConfig(mailConfigPath)
			if err == nil {
				err = cfg.LoginTest()
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
}

func getAccount(account *[2]string) (err error) {
	var data string
	for len(data) == 0 && err != io.EOF { // avoid expect new line error
		fmt.Print("输入用户名:")
		_, err = fmt.Scanln(&data)
	}

	var passwd []byte
	for len(passwd) == 0 && err == nil {
		fmt.Print("输入密码:")
		passwd, err = term.ReadPassword(int(syscall.Stdin))
		fmt.Print("\n") // print in new line
	}
	if err == nil {
		account[0], account[1] = data, string(passwd)
	}
	return
}
