package main

import (
	"flag"
	"os"

	"github.com/opensourceways/community-robot-lib/atomgitclient"

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

	err := fs.Parse(args)
	if err != nil {
		return options{}
	}
	return o
}

func main() {
	logrusutil.ComponentInit(botName)

	o := gatherOptions(flag.NewFlagSet(os.Args[0], flag.ExitOnError), os.Args[1:]...)
	if err := o.Validate(); err != nil {
		logrus.WithError(err).Fatal("Invalid options")
	}

	secretAgent := new(secret.Agent)
	if err := secretAgent.Start([]string{o.atomgit.TokenPath}); err != nil {
		logrus.WithError(err).Fatal("Error starting secret agent.")
	}

	defer secretAgent.Stop()

	c := atomgitclient.NewClient(secretAgent.GetTokenGenerator(o.atomgit.TokenPath))

	r := newRobot(c)

	framework.Run(r, o.service, o.atomgit)
}
