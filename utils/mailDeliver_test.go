package utils

import (
	"testing"
)

var (
	config *EmailConfig
)

func TestWrapper(t *testing.T) {
	testLoadEmailConfig(t)
	testLoginTest(t)
	testSendMail(t)
}

func testLoadEmailConfig(t *testing.T) {
	var err error
	config, err = LoadEmailConfig("not/exist.json")
	if config != nil || err == nil {
		t.Fail()
	}
	config, err = LoadEmailConfig("../email.json")
	if config == nil || err != nil {
		t.Fatal(err)
	}

}

func testLoginTest(t *testing.T) {
	err := config.SMTP.LoginTest()
	if err != nil {
		t.Fatal(err)
	}
}

func testSendMail(t *testing.T) {
	err := config.SendMail("测试邮件",
		"测试",
		"这是一封测试邮件",
	)
	if err != nil {
		t.Fatal(err)
	}
}
