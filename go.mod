module go.thethings.network/lorawan-stack

replace github.com/grpc-ecosystem/grpc-gateway => github.com/TheThingsIndustries/grpc-gateway v1.9.4-gogo

replace github.com/robertkrimen/otto => github.com/TheThingsIndustries/otto v0.0.0-20181129100957-6ddbbb60554a

// Pin versions of golang.org/x modules, because one (or more) of our other deps
// is importing invalid versions. Also, we don't need 10 different versions of
// the same module.

replace golang.org/x/crypto => golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4

replace golang.org/x/image => golang.org/x/image v0.0.0-20190622003408-7e034cad6442

replace golang.org/x/lint => golang.org/x/lint v0.0.0-20190409202823-959b441ac422

replace golang.org/x/net => golang.org/x/net v0.0.0-20190628185345-da137c7871d7

replace golang.org/x/oauth2 => golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45

replace golang.org/x/sync => golang.org/x/sync v0.0.0-20190423024810-112230192c58

replace golang.org/x/sys => golang.org/x/sys v0.0.0-20190626221950-04f50cda93cb

replace golang.org/x/tools => golang.org/x/tools v0.0.0-20190702201734-44aeb8b7c377

require (
	contrib.go.opencensus.io/exporter/prometheus v0.1.0
	github.com/PuerkitoBio/purell v1.1.1
	github.com/TheThingsIndustries/magepkg v0.0.0-20190214092847-6c0299b7c3ed
	github.com/TheThingsIndustries/mystique v0.0.0-20190516134627-66efd81c68ea
	github.com/TheThingsIndustries/release-notes v0.0.2
	github.com/TheThingsNetwork/go-cayenne-lib v1.0.0
	github.com/aws/aws-sdk-go v1.20.14
	github.com/blang/semver v3.6.1+incompatible
	github.com/certifi/gocertifi v0.0.0-20190506164543-d2eda7129713 // indirect
	github.com/client9/misspell v0.3.4
	github.com/disintegration/imaging v1.6.0
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/envoyproxy/protoc-gen-validate v0.1.0
	github.com/fsnotify/fsnotify v1.4.7
	github.com/getsentry/raven-go v0.2.0
	github.com/go-logfmt/logfmt v0.4.0 // indirect
	github.com/go-mail/mail v2.3.1+incompatible
	github.com/go-redis/redis v6.15.3+incompatible
	github.com/gobuffalo/envy v1.7.0 // indirect
	github.com/gobwas/glob v0.2.3
	github.com/gogo/protobuf v1.2.1
	github.com/gohugoio/hugo v0.55.6
	github.com/golang/gddo v0.0.0-20190419222130-af0f2af80721
	github.com/golang/protobuf v1.3.1
	github.com/goreleaser/goreleaser v0.111.0
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/websocket v1.4.0
	github.com/gotnospirit/makeplural v0.0.0-20180622080156-a5f48d94d976 // indirect
	github.com/gotnospirit/messageformat v0.0.0-20180622080451-0eab1176a3fb
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.9.3
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/howeyc/gopass v0.0.0-20170109162249-bf9dde6d0d2c
	github.com/jacobsa/crypto v0.0.0-20190317225127-9f44e2d11115
	github.com/jacobsa/oglematchers v0.0.0-20150720000706-141901ea67cd // indirect
	github.com/jacobsa/oglemock v0.0.0-20150831005832-e94d794d06ff // indirect
	github.com/jacobsa/ogletest v0.0.0-20170503003838-80d50a735a11 // indirect
	github.com/jacobsa/reqtrace v0.0.0-20150505043853-245c9e0234cb // indirect
	github.com/jarcoal/httpmock v1.0.4
	github.com/jaytaylor/html2text v0.0.0-20190408195923-01ec452cbe43
	github.com/jinzhu/gorm v1.9.10
	github.com/jtacoma/uritemplates v1.0.0
	github.com/kr/pretty v0.1.0
	// Do not upgrade Echo beyond v4.1.2 - see https://github.com/TheThingsNetwork/lorawan-stack/issues/977 .
	github.com/labstack/echo/v4 v4.1.2
	github.com/labstack/gommon v0.2.9
	github.com/lib/pq v1.1.1
	github.com/magefile/mage v1.8.1-0.20190702025601-9a6d7fe3be74
	github.com/mattn/go-isatty v0.0.8
	github.com/mattn/goveralls v0.0.2
	github.com/mdempsky/unconvert v0.0.0-20190325185700-2f5dc3378ed3
	github.com/mgechev/revive v0.0.0-20190702162933-cf3705f1b271
	github.com/mitchellh/mapstructure v1.1.2
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/nats-io/gnatsd v1.4.1
	github.com/nats-io/go-nats v1.7.2
	github.com/nats-io/nats-server v1.4.1
	github.com/nats-io/nats-server/v2 v2.0.0 // indirect
	github.com/nats-io/nats.go v1.8.1
	github.com/oklog/ulid v2.0.0+incompatible
	github.com/openshift/osin v1.0.1
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.0.0
	github.com/robertkrimen/otto v0.0.0-00010101000000-000000000000
	github.com/satori/go.uuid v1.2.0
	github.com/sendgrid/rest v2.4.1+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.5.0+incompatible
	github.com/smartystreets/assertions v1.0.0
	github.com/soheilhy/cmux v0.1.4
	github.com/spf13/cast v1.3.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.4.0
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	go.opencensus.io v0.22.0
	go.thethings.network/lorawan-stack-legacy v0.0.0-20190118141410-68812c833a78
	gocloud.dev v0.15.0
	gocloud.dev/pubsub/natspubsub v0.15.0
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4
	golang.org/x/image v0.0.0-20190622003408-7e034cad6442 // indirect
	golang.org/x/net v0.0.0-20190628185345-da137c7871d7
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/tools v0.0.0-20190702201734-44aeb8b7c377
	golang.org/x/xerrors v0.0.0-20190513163551-3ee3066db522
	google.golang.org/api v0.7.0
	google.golang.org/genproto v0.0.0-20190701230453-710ae3a149df
	google.golang.org/grpc v1.22.0
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/mail.v2 v2.3.1 // indirect
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/square/go-jose.v2 v2.3.1
	gopkg.in/yaml.v2 v2.2.2
)
