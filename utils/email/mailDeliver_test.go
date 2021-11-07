package email

import (
	"testing"
)

var (
	config *Config
)

func TestWrapper(t *testing.T) {
	testLoadConfig(t)
	testLoginTest(t)
	testSend(t)
}

func testLoadConfig(t *testing.T) {
	var err error
	config, err = LoadConfig("not/exist.json")
	if config != nil || err == nil {
		t.Fail()
	}
	config, err = LoadConfig("../../email.json")
	if config == nil || err != nil {
		t.Fatal(err)
	}
}

func testLoginTest(t *testing.T) {
	err := config.LoginTest()
	if err != nil {
		t.Fatal(err)
	}
}

func testSend(t *testing.T) {
	err := config.Send("测试邮件",
		"测试",
		"这是一封测试邮件",
	)
	if err != nil {
		t.Fatal(err)
	}
}
