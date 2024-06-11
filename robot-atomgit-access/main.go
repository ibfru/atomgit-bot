package main

import (
	"flag"
	_ "github.com/opensourceways/community-robot-lib/config"
	_ "github.com/opensourceways/community-robot-lib/interrupts"
	"github.com/opensourceways/community-robot-lib/logrusutil"
	liboptions "github.com/opensourceways/community-robot-lib/options"
	framework "github.com/opensourceways/community-robot-lib/robot-atomgit-framework"
	"github.com/opensourceways/community-robot-lib/secret"
	_ "github.com/opensourceways/community-robot-lib/utils"
	"github.com/sirupsen/logrus"
	_ "net/http"
	_ "strconv"
	"time"
)

type options struct {
	service liboptions.ServiceOptions
	atomgit liboptions.AtomGitOptions
}

func (o *options) Validate() error {
	if err := o.service.Validate(); err != nil {
		return err
	}

	return o.atomgit.Validate()
}

func gatherOptions(fs *flag.FlagSet, args ...string) options {
	var opt options

	opt.atomgit.AddFlags(fs)
	opt.service.AddFlags(fs)

	_ = fs.Parse(args)

	return opt
}

func main() {
	logrusutil.ComponentInit(botName)

	//o := gatherOptions(flag.NewFlagSet(os.Args[0], flag.ExitOnError), os.Args[1:]...)
	opt := options{
		service: liboptions.ServiceOptions{
			Port:        8822,
			ConfigFile:  "D:\\Project\\github\\opensourceways\\develop\\atomgit\\robot-gitee-access\\local\\config.yaml",
			GracePeriod: 300 * time.Second,
		},
		atomgit: liboptions.AtomGitOptions{
			TokenPath: "D:\\Project\\github\\opensourceways\\develop\\atomgit\\robot-gitee-access\\local\\secret",
		},
	}

	if err := opt.Validate(); err != nil {
		logrus.WithError(err).Fatal("Invalid options")
	}

	secretAgent := new(secret.Agent)
	if err := secretAgent.Start([]string{opt.atomgit.TokenPath}); err != nil {
		logrus.WithError(err).Fatal("Error starting secret agent.")
	}
	defer secretAgent.Stop()

	// to replace

	p := newRobot()
	opt.atomgit.TokenGenerator = secretAgent.GetTokenGenerator(opt.atomgit.TokenPath)
	framework.Run(p, opt.service, opt.atomgit)
}
