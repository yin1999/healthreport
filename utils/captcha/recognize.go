package captcha

import (
	"sync"

	"github.com/otiai10/gosseract/v2"
)

var (
	client *gosseract.Client
	mux    = sync.Mutex{}
)

func Init() {
	client = gosseract.NewClient()
	client.SetLanguage("digits")
	client.SetWhitelist("0123456789")
}

func Close() error {
	return client.Close()
}

func Recognize(data []byte) (text string, err error) {
	mux.Lock()
	defer mux.Unlock()
	client.SetImageFromBytes(data)
	text, err = client.Text()
	return
}
