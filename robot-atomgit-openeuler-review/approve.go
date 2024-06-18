package main

import (
	"fmt"
	"regexp"
)

const approvedLabel = "approved"

var (
	regAddApprove    = regexp.MustCompile(`(?mi)^/approve\s*$`)
	regRemoveApprove = regexp.MustCompile(`(?mi)^/approve cancel\s*$`)
)

func (bot *robot) handleApprove(p *parameter) error {

	if regAddApprove.MatchString(p.commentContent) {
		return bot.AddApprove(p)
	}

	if regRemoveApprove.MatchString(p.commentContent) {
		return bot.removeApprove(p)
	}

	return nil
}

func (bot *robot) AddApprove(p *parameter) error {

	isBranchKeeper, IsSetBranch, err := bot.CheckBranchKeeper(p)
	if err != nil {
		return err
	}

	if IsSetBranch && !isBranchKeeper {
		return bot.cli.CreatePRComment(p.prArg, fmt.Sprintf(commentNoPermissionForLabel, p.commentator, "add", approvedLabel))
	}

	if !IsSetBranch {
		v, e := bot.hasPermission(p, false)

		if e != nil {
			return e
		}

		if !v {
			return bot.cli.CreatePRComment(p.prArg, fmt.Sprintf(commentNoPermissionForLabel, p.commentator, "add", approvedLabel))
		}

	}

	if err = bot.cli.AddPRLabel(p.prArg, approvedLabel); err != nil {
		return err
	}

	if err = bot.cli.CreatePRComment(p.prArg, fmt.Sprintf(commentAddLabel, approvedLabel, p.commentator)); err != nil {
		p.log.Error(err)
	}

	return bot.tryMerge(p, false)
}

func (bot *robot) removeApprove(p *parameter) error {
	isBranchKeeper, IsSetBranch, err := bot.CheckBranchKeeper(p)
	if err != nil {
		return err
	}

	if IsSetBranch && !isBranchKeeper {
		return bot.cli.CreatePRComment(p.prArg, fmt.Sprintf(commentNoPermissionForLabel, p.commentator, "remove", approvedLabel))
	}

	if !IsSetBranch {
		v, err := bot.hasPermission(p, false)

		if err != nil {
			return err
		}

		if !v {
			return bot.cli.CreatePRComment(p.prArg, fmt.Sprintf(commentNoPermissionForLabel, p.commentator, "remove", approvedLabel))
		}

	}

	err = bot.cli.RemovePRLabel(p.prArg, approvedLabel)
	if err != nil {
		return err
	}

	return bot.cli.CreatePRComment(p.prArg, fmt.Sprintf(commentRemovedLabel, approvedLabel, p.commentator))
}
