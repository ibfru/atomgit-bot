package main

import (
	"k8s.io/apimachinery/pkg/util/sets"
)

type sigDiffBySha struct {
	Maintainers, Committers    sets.Set[string]
	Org, Repo, Name, Path, Sha string
	Size                       float64
}

func (bot *robot) getSigOfRepo(org, repo string) (string, error) {
	// TODO use service[sig-info-cache]
	//sigName, err := bot.findSigName(org, repo, cfg, true)
	//if err != nil {
	//	return sigName, err
	//}

	return "sigName", nil
}
