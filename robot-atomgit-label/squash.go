package main

import (
	"github.com/opensourceways/go-atomgit/atomgit"
)

func (bot *robot) handleSquashLabel(e *atomgit.PullRequestEvent, commits uint, cfg SquashConfig) error {
	if cfg.unableCheckingSquash() {
		return nil
	}

	action := e.GetAction()
	if action != atomgit.ActionStateSynchronized {
		return nil
	}

	//labels := e.GetLabel()
	//hasSquashLabel := false
	////hasSquashLabel := labels.se(cfg.SquashCommitLabel)
	//exceeded := commits > cfg.CommitsThreshold
	//org, repo := e.GetRepo().GetOrgAndRepo()
	//number := e.GetNumber()
	//
	//if exceeded && !hasSquashLabel {
	//	return bot.cli.AddPRLabel(atomgitclient.BuildPRIssue(org, repo, number), cfg.SquashCommitLabel)
	//}
	//
	//if !exceeded && hasSquashLabel {
	//	return bot.cli.RemovePRLabel(atomgitclient.BuildPRIssue(org, repo, number), cfg.SquashCommitLabel)
	//}

	return nil
}
