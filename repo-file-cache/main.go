package main

import (
	"context"
	"os"
	"time"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/opensourceways/server-common-lib/interrupts"

	_ "github.com/opensourceways/repo-file-cache/routers"

	"github.com/opensourceways/repo-file-cache/config"
	"github.com/opensourceways/repo-file-cache/dbmodels"
	"github.com/opensourceways/repo-file-cache/mongodb"
)

func main() {
	configFile, err := beego.AppConfig.String("appconf")
	if err != nil {
		fatal(err)
	}

	if err := config.InitAppConfig(configFile); err != nil {
		fatal(err)
	}
	cfg := config.AppConfig

	clears := []func(){}
	defer func() {
		for _, f := range clears {
			f()
		}
		logs.Info("server exits.")
	}()

	f, err := startMongoService(&cfg.Mongodb)
	if err != nil {
		fatal(err)
	}
	clears = append(clears, f)

	run()
}

func fatal(err error) {
	logs.Error(err)
	os.Exit(1)
}

func startMongoService(cfg *config.MongodbConfig) (func(), error) {
	c, err := mongodb.Initialize(cfg)
	if err != nil {
		return nil, err
	}
	dbmodels.RegisterDB(c)

	return func() {
		logs.Info("closing mongodb ...")
		if err := dbmodels.GetDB().Close(); err != nil {
			logs.Error(err)
		}
	}, nil
}

func run() {
	defer interrupts.WaitForGracefulShutdown()

	interrupts.OnInterrupt(func() {
		shutdown()
	})

	beego.Run()
}

func shutdown() {
	logs.Info("server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := beego.BeeApp.Server.Shutdown(ctx); err != nil {
		logs.Error("error to shut down server, err:", err.Error())
	}
	cancel()
}
