package sdk

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/opensourceways/server-common-lib/utils"

	"github.com/opensourceways/repo-file-cache/models"
)

func NewSDK(endpoint string, maxRetries int) *SDK {
	slash := "/"
	if !strings.HasSuffix(endpoint, slash) {
		endpoint += slash
	}

	return &SDK{
		hc:              utils.NewHttpClient(maxRetries),
		endpoint:        endpoint,
		summaryEndpoint: endpoint + "%s?branch=%s&summary=%t",
	}
}

type SDK struct {
	hc              utils.HttpClient
	endpoint        string
	summaryEndpoint string
}

func (cli *SDK) SaveFiles(opt models.FileUpdateOption) error {
	payload, err := utils.JsonMarshal(&opt)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, cli.endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	return cli.forwardTo(req, nil)
}

func (cli *SDK) GetFiles(b models.Branch, fileName string, summary bool) (models.FilesInfo, error) {
	endpoint := fmt.Sprintf(cli.summaryEndpoint, fileName, models.BranchToKey(&b), summary)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return models.FilesInfo{}, err
	}

	var v struct {
		Data models.FilesInfo `json:"data"`
	}

	if err = cli.forwardTo(req, &v); err != nil {
		return models.FilesInfo{}, err
	}

	return v.Data, nil
}

func (cli *SDK) DeleteFiles(opt models.FileDeleteOption) error {
	payload, err := utils.JsonMarshal(&opt)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodDelete, cli.endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	return cli.forwardTo(req, nil)
}

func (cli *SDK) forwardTo(req *http.Request, jsonResp interface{}) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "repo-file-cache-sdk")

	_, err := cli.hc.ForwardTo(req, jsonResp)

	return err
}
