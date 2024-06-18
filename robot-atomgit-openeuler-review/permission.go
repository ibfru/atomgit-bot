package main

import (
	"encoding/base64"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/yaml"
)

const ownerFile = "OWNERS"
const sigInfoFile = "sig-info.yaml"

func (bot *robot) hasPermission(p *parameter, needCheckSig bool) (bool, error) {

	prm, err := bot.cli.GetUserPermissionOfRepo(p.prArg.Org, p.prArg.Repo, p.commentator)
	if err != nil {
		return false, err
	}

	if *prm.Permission == "admin" || *prm.Permission == "write" {
		return true, nil
	}

	if needCheckSig {
		return bot.isOwnerOfSig(p)
	}

	return false, nil
}

func (bot *robot) isOwnerOfSig(p *parameter) (bool, error) {
	changes, err := bot.cli.GetPullRequestChanges(p.prArg)
	if err != nil || len(changes) == 0 {
		return false, err
	}

	paths := sets.NewString()
	for _, file := range changes {
		if !p.bcf.regSigDir.MatchString(*file.Filename) || strings.Count(*file.Filename, "/") > 2 {
			return false, nil
		}

		paths.Insert(filepath.Dir(*file.Filename))
	}

	// ownerFile sigInfoFile
	// org repo => get userid from cache compare to commentator
	return false, nil
}

func decodeSigInfoFile(content string, log *logrus.Entry) sets.String {
	owners := sets.NewString()

	c, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		log.WithError(err).Error("decode file")

		return owners
	}

	var m SigInfo

	if err = yaml.Unmarshal(c, &m); err != nil {
		log.WithError(err).Error("code yaml file")

		return owners
	}

	for _, v := range m.Maintainers {
		owners.Insert(strings.ToLower(v.GiteeID))
	}

	return owners
}

func decodeOwnerFile(content string, log *logrus.Entry) sets.String {
	owners := sets.NewString()

	c, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		log.WithError(err).Error("decode file")

		return owners
	}

	var m struct {
		Maintainers []string `yaml:"maintainers"`
		Committers  []string `yaml:"committers"`
	}

	if err = yaml.Unmarshal(c, &m); err != nil {
		log.WithError(err).Error("code yaml file")

		return owners
	}

	for _, v := range m.Maintainers {
		owners.Insert(strings.ToLower(v))
	}

	for _, v := range m.Committers {
		owners.Insert(strings.ToLower(v))
	}

	return owners
}
