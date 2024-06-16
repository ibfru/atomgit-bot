module github.com/opensourceways/community-robot-lib

go 1.20

require (
	github.com/Shopify/sarama v1.34.1
	github.com/antihax/optional v1.0.0
	github.com/google/go-github/v36 v36.0.0
	github.com/google/uuid v1.3.0
	github.com/opensourceways/go-atomgit v0.0.0-00010101000000-000000000000
	github.com/opensourceways/go-gitee v0.0.0-20240305060727-0df28a4f60c0
	github.com/sirupsen/logrus v1.9.3
	github.com/xanzy/go-gitlab v0.68.0
	golang.org/x/oauth2 v0.21.0
	k8s.io/apimachinery v0.25.3
	sigs.k8s.io/yaml v1.3.0
)

replace github.com/opensourceways/go-atomgit v0.0.0-00010101000000-000000000000 => ../go-atomgit

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/eapache/go-resiliency v1.4.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230731223053-c322873962e3 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.1 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/klauspost/compress v1.17.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.18 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	golang.org/x/crypto v0.14.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
