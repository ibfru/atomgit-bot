module github.com/opensourceways/robot-atomgit-access

go 1.20

require (
	github.com/opensourceways/community-robot-lib v0.0.0-20211025094652-e48b92d2df4f
	github.com/opensourceways/go-atomgit v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.9.3
	k8s.io/apimachinery v0.25.3
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/stretchr/testify v1.8.1 // indirect
	golang.org/x/oauth2 v0.12.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace github.com/opensourceways/community-robot-lib v0.0.0-20211025094652-e48b92d2df4f => ../community-robot-lib

replace github.com/opensourceways/go-atomgit v0.0.0-00010101000000-000000000000 => ../go-atomgit
