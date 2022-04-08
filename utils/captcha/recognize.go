package captcha

import (
	"strings"

	"github.com/otiai10/gosseract/v2"
)

func Recognize(data []byte) (text string, err error) {
	client := gosseract.NewClient()
	client.SetLanguage("digits")
	client.SetImageFromBytes(data)
	text, err = client.Text()
	if err == nil {
		text = strings.ReplaceAll(text, " ", "")
	}
	return
}
