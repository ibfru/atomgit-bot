package framework

import (
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/opensourceways/community-robot-lib/config"
	"github.com/opensourceways/community-robot-lib/interrupts"
	"github.com/opensourceways/community-robot-lib/options"
)

type HandlerRegister interface {
	RegisterAccessHandler(handler AccessHandler)
	RegisterIssueHandler(IssueHandler)
	RegisterPullRequestHandler(PullRequestHandler)
	RegisterPushEventHandler(PushEventHandler)
	RegisterIssueCommentHandler(IssueCommentHandler)
	RegisterReviewEventHandler(ReviewEventHandler)
	RegisterReviewCommentEventHandler(ReviewCommentEventHandler)
}

type Robot interface {
	NewConfig() config.Config
	RegisterEventHandler(HandlerRegister)
}

func Run(bot Robot, servOpt options.ServiceOptions, atomgitOpt options.AtomGitOptions) {
	agent := config.NewConfigAgent(bot.NewConfig)
	if err := agent.Start(servOpt.ConfigFile); err != nil {
		logrus.WithError(err).Errorf("start config:%s", servOpt.ConfigFile)
		return
	}

	h := handlers{}
	bot.RegisterEventHandler(&h)

	d := &dispatcher{agent: &agent, h: h, hmac: atomgitOpt.TokenGenerator}

	defer interrupts.WaitForGracefulShutdown()

	interrupts.OnInterrupt(func() {
		agent.Stop()
		d.Wait()
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// service's healthy check, do nothing
	})

	http.Handle("/atomgit-hook", d)

	httpServer := &http.Server{Addr: ":" + strconv.Itoa(servOpt.Port)}

	interrupts.ListenAndServe(httpServer, servOpt.GracePeriod)
}
