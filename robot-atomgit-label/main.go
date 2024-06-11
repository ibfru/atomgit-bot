package main

import (
	"flag"
	atomgitclient "github.com/opensourceways/community-robot-lib/atomgitclient"
	_ "github.com/opensourceways/go-atomgit/atomgit"
	"time"

	//ss "../go-atomgit"
	"github.com/opensourceways/community-robot-lib/logrusutil"
	liboptions "github.com/opensourceways/community-robot-lib/options"
	framework "github.com/opensourceways/community-robot-lib/robot-atomgit-framework"
	"github.com/opensourceways/community-robot-lib/secret"
	"github.com/sirupsen/logrus"
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
	var o options

	o.atomgit.AddFlags(fs)
	o.service.AddFlags(fs)

	_ = fs.Parse(args)

	return o
}

func main() {
	logrusutil.ComponentInit(botName)

	//o := gatherOptions(flag.NewFlagSet(os.Args[0], flag.ExitOnError), os.Args[1:]...)
	o := options{
		service: liboptions.ServiceOptions{
			Port:        8863,
			ConfigFile:  "D:\\Project\\github\\opensourceways\\develop\\atomgit\\robot-gitee-label\\local\\config.yaml",
			GracePeriod: 300 * time.Second,
		},
		atomgit: liboptions.AtomGitOptions{
			TokenPath: "D:\\Project\\github\\opensourceways\\develop\\atomgit\\robot-gitee-label\\local\\token",
		},
	}
	if err := o.Validate(); err != nil {
		logrus.WithError(err).Fatal("Invalid options")
	}

	secretAgent := new(secret.Agent)
	if err := secretAgent.Start([]string{o.atomgit.TokenPath}); err != nil {
		logrus.WithError(err).Fatal("Error starting secret agent.")
	}

	defer secretAgent.Stop()

	//c := atomgitlib.NewClient(secretAgent.GetTokenGenerator(o.atomgit.TokenPath))
	c := atomgitclient.NewClient(secretAgent.GetTokenGenerator(o.atomgit.TokenPath))
	p := newRobot(c)

	framework.Run(p, o.service, o.atomgit)
}
