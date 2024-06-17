package main

import (
	"flag"
	"net/url"
	"time"

	cache "github.com/opensourceways/atomgit-sig-file-cache/sdk"
	"github.com/opensourceways/community-robot-lib/atomgitclient"
	"github.com/opensourceways/community-robot-lib/logrusutil"
	liboptions "github.com/opensourceways/community-robot-lib/options"
	framework "github.com/opensourceways/community-robot-lib/robot-atomgit-framework"
	"github.com/opensourceways/community-robot-lib/secret"
	"github.com/sirupsen/logrus"
)

type options struct {
	service       liboptions.ServiceOptions
	atomgit       liboptions.AtomGitOptions
	cacheEndpoint string
	maxRetries    int
}

func (o *options) Validate() error {
	if _, err := url.ParseRequestURI(o.cacheEndpoint); err != nil {
		return err
	}

	if err := o.service.Validate(); err != nil {
		return err
	}

	return o.atomgit.Validate()
}

func gatherOptions(fs *flag.FlagSet, args ...string) options {
	var o options

	o.atomgit.AddFlags(fs)
	o.service.AddFlags(fs)
	retry := 3
	fs.StringVar(&o.cacheEndpoint, "cache-endpoint", "", "The endpoint of repo file cache")
	fs.IntVar(&o.maxRetries, "max-retries", retry, "The number of failed retry attempts to call the cache api")

	_ = fs.Parse(args)
	return o
}

func main() {
	logrusutil.ComponentInit(botName)

	//o := gatherOptions(flag.NewFlagSet(os.Args[0], flag.ExitOnError), os.Args[1:]...)
	o := options{
		service: liboptions.ServiceOptions{
			Port:        8833,
			ConfigFile:  "D:\\Project\\github\\ibfru\\atomgit-bot\\robot-atomgit-openeuler-welcome\\local\\config.yaml",
			GracePeriod: 300 * time.Second,
		},
		atomgit: liboptions.AtomGitOptions{
			TokenPath:     "D:\\Project\\github\\ibfru\\atomgit-bot\\robot-atomgit-openeuler-welcome\\local\\token",
			RepoCacheDir:  "",
			CacheRepoOnPV: true,
		},
		cacheEndpoint: "http://localhost:8888/v1/file",
		maxRetries:    1,
	}
	if err := o.Validate(); err != nil {
		logrus.WithError(err).Fatal("Invalid options")
	}

	secretAgent := new(secret.Agent)
	if err := secretAgent.Start([]string{o.atomgit.TokenPath}); err != nil {
		logrus.WithError(err).Fatal("Error starting secret agent.")
	}

	defer secretAgent.Stop()

	c := atomgitclient.NewClient(secretAgent.GetTokenGenerator(o.atomgit.TokenPath))
	s := cache.NewSDK(o.cacheEndpoint, o.maxRetries)

	p := newRobot(c, s)

	framework.Run(p, o.service, o.atomgit)
}
