module github.com/opensourceways/robot-gitee-cla

go 1.16

require (
	github.com/huaweicloud/golangsdk v0.0.0-20210831081626-d823fe11ceba
	github.com/opensourceways/community-robot-lib v0.0.0-20220117111729-62e2fe1e7b9e
	github.com/opensourceways/go-atomgit v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.9.3
)

replace (
	github.com/opensourceways/community-robot-lib v0.0.0-20220117111729-62e2fe1e7b9e => ../community-robot-lib
	github.com/opensourceways/go-atomgit v0.0.0-00010101000000-000000000000 => ../go-atomgit
)
