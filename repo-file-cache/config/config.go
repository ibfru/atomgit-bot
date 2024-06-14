package config

import (
	"errors"
	"fmt"

	"github.com/opensourceways/repo-file-cache/util"
)

var AppConfig *appConfig

func InitAppConfig(path string) error {
	v := new(appConfig)
	if err := util.LoadFromYaml(path, v); err != nil {
		return err
	}

	v.setDefault()

	if err := v.validate(); err != nil {
		return err
	}

	AppConfig = v
	return nil
}

type appConfig struct {
	Mongodb MongodbConfig `json:"mongodb" required:"true"`
}

func (cfg *appConfig) setDefault() {
}

func (cfg *appConfig) validate() error {
	if err := cfg.Mongodb.validate(); err != nil {
		return fmt.Errorf("error config for item: mongodb, err: %s", err.Error())
	}
	return nil
}

type MongodbConfig struct {
	MongodbConn     string `json:"mongodb_conn" required:"true"`
	DBName          string `json:"mongodb_db" required:"true"`
	FilesCollection string `json:"files_collection" required:"true"`
}

func (m MongodbConfig) validate() error {
	missing := func(k string) error {
		return errors.New("missing parameter: " + k)
	}

	if m.MongodbConn == "" {
		return missing("mongodb_conn")
	}

	if m.DBName == "" {
		return missing("mongodb_db")
	}

	if m.FilesCollection == "" {
		return missing("files_collection")
	}

	return nil
}
