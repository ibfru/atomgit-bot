module github.com/opensourceways/robot-atomgit-cla

go 1.20

require (
	github.com/huaweicloud/golangsdk v0.0.0-20210831081626-d823fe11ceba
	github.com/opensourceways/community-robot-lib v0.0.0-20220117111729-62e2fe1e7b9e
	github.com/opensourceways/go-atomgit v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.9.3
	k8s.io/apimachinery v0.25.3
)

require (
	github.com/google/go-querystring v1.1.0 // indirect
	golang.org/x/oauth2 v0.21.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace (
	github.com/opensourceways/community-robot-lib v0.0.0-20220117111729-62e2fe1e7b9e => ../community-robot-lib
	github.com/opensourceways/go-atomgit v0.0.0-00010101000000-000000000000 => ../go-atomgit
)
