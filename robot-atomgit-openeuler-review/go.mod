module github.com/opensourceways/robot-atomgit-openeuler-review

go 1.20

require (
	github.com/opensourceways/atomgit-sig-file-cache v0.0.0-00010101000000-000000000000
	github.com/opensourceways/community-robot-lib v0.0.0-20220118064921-28924d0a1246
	github.com/opensourceways/go-atomgit v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.9.3
	k8s.io/apimachinery v0.29.1
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/opensourceways/server-common-lib v0.0.0-20240206030521-a8e5c88d6816 // indirect
	golang.org/x/oauth2 v0.21.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace (
	github.com/opensourceways/atomgit-sig-file-cache v0.0.0-00010101000000-000000000000 => ../atomgit-sig-info-cache
	github.com/opensourceways/community-robot-lib v0.0.0-20220118064921-28924d0a1246 => ../community-robot-lib
	github.com/opensourceways/go-atomgit v0.0.0-00010101000000-000000000000 => ../go-atomgit
)
