package tool

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

func FetchSelfPublicIP() (string, error) {
	cli := resty.New().SetTimeout(1 * time.Second).SetRetryCount(3).SetRetryWaitTime(1 * time.Second).GetClient()
	response, err := cli.Get("https://ifconfig.me")
	if err != nil {
		return "", err
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http status code is not 200, body: %s", string(body))
	}
	return string(body), nil
}
