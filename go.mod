go 1.13

module go.thethings.network/lorawan-stack

// Use our fork of grpc-gateway.
replace github.com/grpc-ecosystem/grpc-gateway => github.com/TheThingsIndustries/grpc-gateway v1.11.1-gogo

// Use our fork of otto.
replace github.com/robertkrimen/otto => github.com/TheThingsIndustries/otto v0.0.0-20181129100957-6ddbbb60554a

// github.com/blang/semver doesn't have a v3 semantic import.
replace github.com/blang/semver v3.5.1+incompatible => github.com/blang/semver v0.0.0-20190414182527-1a9109f8c4a1

// github.com/goreleaser/goreleaser uses invalid syntax for dependency version.
replace github.com/go-macaron/cors v0.0.0-20190309005821-6fd6a9bfe14e9 => github.com/go-macaron/cors v0.0.0-20190418220122-6fd6a9bfe14e

// github.com/goreleaser/goreleaser depends on version of github.com/Azure/go-autorest, which has broken module management.
// See https://github.com/Azure/go-autorest/issues/414.
replace github.com/Azure/go-autorest v12.0.0+incompatible => github.com/Azure/go-autorest v11.1.2+incompatible

// Ensure only one version of golang.org/x/crypto is required.
replace golang.org/x/crypto => golang.org/x/crypto v0.0.0-20190911031432-227b76d455e7

// Ensure only one version of golang.org/x/image is required.
replace golang.org/x/image => golang.org/x/image v0.0.0-20190910094157-69e4b8554b2a // indirect

// Ensure only one version of golang.org/x/net is required.
replace golang.org/x/net => golang.org/x/net v0.0.0-20190909003024-a7b16738d86b

// Ensure only one version of golang.org/x/oauth2 is required.
replace golang.org/x/oauth2 => golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45

// Ensure only one version of golang.org/x/sys is required.
replace golang.org/x/sys => golang.org/x/sys v0.0.0-20190910064555-bbd175535a8b // indirect

// Ensure only one version of golang.org/x/tools is required.
replace golang.org/x/tools => golang.org/x/tools v0.0.0-20190911022129-16c5e0f7d110

// Ensure only one version of golang.org/x/xerrors is required.
replace golang.org/x/xerrors => golang.org/x/xerrors v0.0.0-20190717185122-a985d3407aa7

require (
	cloud.google.com/go/pubsub v1.0.1 // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.1.0
	github.com/Azure/azure-pipeline-go v0.2.2 // indirect
	github.com/Azure/azure-storage-blob-go v0.8.0 // indirect
	github.com/PuerkitoBio/purell v1.1.1
	github.com/TheThingsIndustries/magepkg v0.0.0-20190214092847-6c0299b7c3ed
	github.com/TheThingsIndustries/mystique v0.0.0-20190516134627-66efd81c68ea
	github.com/TheThingsIndustries/release-notes v0.1.0
	github.com/TheThingsNetwork/go-cayenne-lib v1.0.0
	github.com/alecthomas/chroma v0.6.6 // indirect
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20190910110746-680d30ca3117 // indirect
	github.com/apex/log v1.1.1 // indirect
	github.com/aws/aws-sdk-go v1.23.19
	github.com/bep/tmc v0.5.1 // indirect
	github.com/blakesmith/ar v0.0.0-20190502131153-809d4375e1fb // indirect
	github.com/blang/semver v3.5.1+incompatible
	github.com/certifi/gocertifi v0.0.0-20190905060710-a5e0173ced67 // indirect
	github.com/client9/misspell v0.3.4
	github.com/disintegration/imaging v1.6.1
	github.com/dlclark/regexp2 v1.2.0 // indirect
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/envoyproxy/protoc-gen-validate v0.2.0-java
	github.com/fsnotify/fsnotify v1.4.7
	github.com/getsentry/raven-go v0.2.0
	github.com/go-redis/redis v6.15.5+incompatible
	github.com/gobuffalo/envy v1.7.0 // indirect
	github.com/gobwas/glob v0.2.3
	github.com/gogo/protobuf v1.3.0
	github.com/gohugoio/hugo v0.58.1
	github.com/golang/gddo v0.0.0-20190904175337-72a348e765d2
	github.com/golang/protobuf v1.3.2
	github.com/google/go-github/v25 v25.1.3 // indirect
	github.com/goreleaser/goreleaser v0.117.2
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/websocket v1.4.1
	github.com/gotnospirit/makeplural v0.0.0-20180622080156-a5f48d94d976 // indirect
	github.com/gotnospirit/messageformat v0.0.0-20190719172517-c1d0bdacdea2
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.11.1
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/howeyc/gopass v0.0.0-20170109162249-bf9dde6d0d2c
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/jacobsa/crypto v0.0.0-20190317225127-9f44e2d11115
	github.com/jacobsa/oglematchers v0.0.0-20150720000706-141901ea67cd // indirect
	github.com/jacobsa/oglemock v0.0.0-20150831005832-e94d794d06ff // indirect
	github.com/jacobsa/ogletest v0.0.0-20170503003838-80d50a735a11 // indirect
	github.com/jacobsa/reqtrace v0.0.0-20150505043853-245c9e0234cb // indirect
	github.com/jarcoal/httpmock v1.0.4
	github.com/jaytaylor/html2text v0.0.0-20190408195923-01ec452cbe43
	github.com/jdkato/prose v1.1.1 // indirect
	github.com/jinzhu/gorm v1.9.10
	github.com/kamilsk/retry/v4 v4.3.1 // indirect
	github.com/kr/pretty v0.1.0
	github.com/kyokomi/emoji v2.1.0+incompatible // indirect
	// Do not upgrade Echo beyond v4.1.2 - see https://github.com/TheThingsNetwork/lorawan-stack/issues/977 .
	github.com/labstack/echo/v4 v4.1.2
	github.com/labstack/gommon v0.3.0
	github.com/lib/pq v1.2.0
	github.com/magefile/mage v1.8.1-0.20190718165527-e1fda1a0ffba
	github.com/markbates/inflect v1.0.4 // indirect
	github.com/mattn/go-ieproxy v0.0.0-20190805055040-f9202b1cfdeb // indirect
	github.com/mattn/go-isatty v0.0.9
	github.com/mattn/go-zglob v0.0.1 // indirect
	github.com/mattn/goveralls v0.0.2
	github.com/mdempsky/unconvert v0.0.0-20190325185700-2f5dc3378ed3
	github.com/mgechev/revive v0.0.0-20190910172647-84deee41635a
	github.com/mitchellh/mapstructure v1.1.2
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/muesli/smartcrop v0.3.0 // indirect
	github.com/nats-io/nats-server/v2 v2.0.4
	github.com/nats-io/nats.go v1.8.1
	github.com/oklog/ulid/v2 v2.0.2
	github.com/openshift/osin v1.0.1
	github.com/opentracing/opentracing-go v1.1.0 // indirect
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.1.0
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4 // indirect
	github.com/prometheus/procfs v0.0.4 // indirect
	github.com/robertkrimen/otto v0.0.0-20180617131154-15f95af6e78d
	github.com/rogpeppe/go-internal v1.3.2 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/sendgrid/rest v2.4.1+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.5.0+incompatible
	github.com/smartystreets/assertions v1.0.1
	github.com/soheilhy/cmux v0.1.4
	github.com/spf13/cast v1.3.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.4.0
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/tdewolff/minify/v2 v2.5.2 // indirect
	github.com/xanzy/go-gitlab v0.20.1 // indirect
	go.opencensus.io v0.22.1
	go.thethings.network/lorawan-stack-legacy v0.0.0-20190118141410-68812c833a78
	gocloud.dev v0.17.0
	gocloud.dev/pubsub/natspubsub v0.16.0
	golang.org/x/crypto v0.0.0-20190911031432-227b76d455e7
	golang.org/x/image v0.0.0-20190910094157-69e4b8554b2a // indirect
	golang.org/x/net v0.0.0-20190909003024-a7b16738d86b
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/sys v0.0.0-20190910064555-bbd175535a8b // indirect
	golang.org/x/tools v0.0.0-20190911022129-16c5e0f7d110
	golang.org/x/xerrors v0.0.0-20190717185122-a985d3407aa7
	google.golang.org/api v0.10.0
	google.golang.org/appengine v1.6.2 // indirect
	google.golang.org/genproto v0.0.0-20190905072037-92dd089d5514
	google.golang.org/grpc v1.23.0
	// Do not upgrade go-sqlmock beyond v1.3.0 until https://github.com/heptiolabs/healthcheck/issues/23 is resolved
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
	gopkg.in/mail.v2 v2.3.1
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/square/go-jose.v2 v2.3.1
	gopkg.in/yaml.v2 v2.2.2
)
