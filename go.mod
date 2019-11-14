module go.thethings.network/lorawan-stack

go 1.13

// Use our fork of grpc-gateway.
replace github.com/grpc-ecosystem/grpc-gateway => github.com/TheThingsIndustries/grpc-gateway v1.12.1-gogo

// Use our fork of otto.
replace github.com/robertkrimen/otto => github.com/TheThingsIndustries/otto v0.0.0-20181129100957-6ddbbb60554a

// github.com/blang/semver doesn't have a v3 semantic import.
replace github.com/blang/semver v3.5.1+incompatible => github.com/blang/semver v0.0.0-20190414182527-1a9109f8c4a1

// github.com/goreleaser/goreleaser uses invalid syntax for dependency version.
replace github.com/go-macaron/cors v0.0.0-20190309005821-6fd6a9bfe14e9 => github.com/go-macaron/cors v0.0.0-20190418220122-6fd6a9bfe14e

// github.com/goreleaser/goreleaser depends on version of github.com/Azure/go-autorest, which has broken module management.
// See https://github.com/Azure/go-autorest/issues/414.
replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.1.0+incompatible

// github.com/goreleaser/goreleaser depends on https://github.com/google/rpmpack/pull/33.
// See https://github.com/goreleaser/goreleaser/commit/ff31e29878dd336d34c11fc189ea01b553864c5b
replace github.com/google/rpmpack => github.com/caarlos0/rpmpack v0.0.0-20191106130752-24a815bfaee0

require (
	cloud.google.com/go v0.48.0 // indirect
	cloud.google.com/go/storage v1.3.0 // indirect
	code.gitea.io/sdk/gitea v0.0.0-20191106151626-e4082d89cc3b // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.1.0
	github.com/Azure/azure-pipeline-go v0.2.2 // indirect
	github.com/Azure/azure-sdk-for-go v36.1.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest/adal v0.8.0 // indirect
	github.com/Azure/go-autorest/autorest/azure/auth v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.2.0 // indirect
	github.com/Masterminds/semver/v3 v3.0.2 // indirect
	github.com/PuerkitoBio/purell v1.1.1
	github.com/TheThingsIndustries/magepkg v0.0.0-20190214092847-6c0299b7c3ed
	github.com/TheThingsIndustries/mystique v0.0.0-20190516134627-66efd81c68ea
	github.com/TheThingsNetwork/go-cayenne-lib v1.0.0
	github.com/alecthomas/chroma v0.6.9 // indirect
	github.com/aws/aws-sdk-go v1.25.35
	github.com/bep/tmc v0.5.1 // indirect
	github.com/blang/semver v3.5.1+incompatible
	github.com/certifi/gocertifi v0.0.0-20191021191039-0944d244cd40 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/chrj/smtpd v0.1.2
	github.com/client9/misspell v0.3.4
	github.com/disintegration/imaging v1.6.1
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/envoyproxy/protoc-gen-validate v0.2.0-java
	github.com/fatih/structtag v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.4.7
	github.com/getsentry/raven-go v0.2.0
	github.com/go-redis/redis v6.15.6+incompatible
	github.com/gobuffalo/envy v1.7.1 // indirect
	github.com/gobwas/glob v0.2.3
	github.com/gogo/protobuf v1.3.1
	github.com/gohugoio/hugo v0.59.1
	github.com/golang/gddo v0.0.0-20190904175337-72a348e765d2
	github.com/golang/groupcache v0.0.0-20191027212112-611e8accdfc9 // indirect
	github.com/golang/protobuf v1.3.2
	github.com/goreleaser/goreleaser v0.121.0
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/websocket v1.4.1
	github.com/gotnospirit/makeplural v0.0.0-20180622080156-a5f48d94d976 // indirect
	github.com/gotnospirit/messageformat v0.0.0-20190719172517-c1d0bdacdea2
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.12.1
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/howeyc/gopass v0.0.0-20190910152052-7cb4b85ec19c
	github.com/jacobsa/crypto v0.0.0-20190317225127-9f44e2d11115
	github.com/jacobsa/oglematchers v0.0.0-20150720000706-141901ea67cd // indirect
	github.com/jacobsa/oglemock v0.0.0-20150831005832-e94d794d06ff // indirect
	github.com/jacobsa/ogletest v0.0.0-20170503003838-80d50a735a11 // indirect
	github.com/jacobsa/reqtrace v0.0.0-20150505043853-245c9e0234cb // indirect
	github.com/jarcoal/httpmock v1.0.4
	github.com/jaytaylor/html2text v0.0.0-20190408195923-01ec452cbe43
	github.com/jdkato/prose v1.1.1 // indirect
	github.com/jinzhu/gorm v1.9.11
	github.com/jstemmer/go-junit-report v0.9.1 // indirect
	github.com/kr/pretty v0.1.0
	// Do not upgrade Echo beyond v4.1.2 - see https://github.com/TheThingsNetwork/lorawan-stack/issues/977 .
	github.com/labstack/echo/v4 v4.1.2
	github.com/labstack/gommon v0.3.0
	github.com/lib/pq v1.2.0
	github.com/magefile/mage v1.9.0
	github.com/markbates/inflect v1.0.4 // indirect
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-ieproxy v0.0.0-20191113090002-7c0f6868bffe // indirect
	github.com/mattn/go-isatty v0.0.10
	github.com/mattn/go-runewidth v0.0.6 // indirect
	github.com/mattn/goveralls v0.0.4
	github.com/mdempsky/unconvert v0.0.0-20190921185256-3ecd357795af
	github.com/mgechev/dots v0.0.0-20190921121421-c36f7dcfbb81 // indirect
	github.com/mgechev/revive v0.0.0-20191017201419-88015ccf8e97
	github.com/mitchellh/mapstructure v1.1.2
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/muesli/smartcrop v0.3.0 // indirect
	github.com/nats-io/jwt v0.3.2 // indirect
	github.com/nats-io/nats-server/v2 v2.1.0
	github.com/nats-io/nats.go v1.9.1
	github.com/niklasfasching/go-org v0.1.8 // indirect
	github.com/oklog/ulid/v2 v2.0.2
	github.com/olekukonko/tablewriter v0.0.2 // indirect
	github.com/openshift/osin v1.0.1
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.2.1
	github.com/prometheus/procfs v0.0.6 // indirect
	github.com/robertkrimen/otto v0.0.0-20181129100957-6ddbbb60554a
	github.com/rogpeppe/go-internal v1.5.0 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/sendgrid/rest v2.4.1+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.5.0+incompatible
	github.com/skip2/go-qrcode v0.0.0-20191027152451-9434209cb086
	github.com/smartystreets/assertions v1.0.1
	github.com/soheilhy/cmux v0.1.4
	github.com/spf13/cast v1.3.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.5.0
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/valyala/fasttemplate v1.1.0 // indirect
	github.com/xanzy/go-gitlab v0.22.0 // indirect
	go.opencensus.io v0.22.2
	go.thethings.network/lorawan-stack-legacy v0.0.0-20190118141410-68812c833a78
	gocloud.dev v0.18.0
	gocloud.dev/pubsub/natspubsub v0.18.0
	golang.org/x/crypto v0.0.0-20191112222119-e1110fd1c708
	golang.org/x/image v0.0.0-20191009234506-e7c1f5e7dbb8 // indirect
	golang.org/x/net v0.0.0-20191112182307-2180aed22343
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/sys v0.0.0-20191113165036-4c7a9d0fe056 // indirect
	golang.org/x/tools v0.0.0-20191114200427-caa0b0f7d508
	golang.org/x/xerrors v0.0.0-20191011141410-1b5146add898
	google.golang.org/api v0.13.0
	google.golang.org/genproto v0.0.0-20191114150713-6bbd007550de
	google.golang.org/grpc v1.25.1
	// Do not upgrade go-sqlmock beyond v1.3.0 until https://github.com/heptiolabs/healthcheck/issues/23 is resolved
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/mail.v2 v2.3.1
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/square/go-jose.v2 v2.4.0
	gopkg.in/yaml.v2 v2.2.5
)
