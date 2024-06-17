package sdk

import (
	"net/http"
	"strings"

	"github.com/opensourceways/server-common-lib/utils"
)

func NewSDK(endpoint string, maxRetries int) *SDK {
	slash := "/"
	if !strings.HasSuffix(endpoint, slash) {
		endpoint += slash
	}

	return &SDK{
		hc:       utils.NewHttpClient(maxRetries),
		endpoint: endpoint,
	}
}

type SDK struct {
	hc       utils.HttpClient
	endpoint string
}

func (cli *SDK) GetSigInfo(urlPath string) (string, error) {

	req, err := http.NewRequest(http.MethodGet, cli.endpoint+urlPath, nil)
	if err != nil {
		return "", err
	}

	var v struct {
		Data map[string]string `json:"data"`
	}

	if err = cli.forwardTo(req, &v); err != nil {
		return "", err
	}

	return v.Data["111"], nil
}

func (cli *SDK) forwardTo(req *http.Request, jsonResp interface{}) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "sig-info-cache-sdk")

	_, err := cli.hc.ForwardTo(req, jsonResp)

	return err
}
