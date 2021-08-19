module go.thethings.network/lorawan-stack/v3

go 1.16

// Use our fork of grpc-gateway.
replace github.com/grpc-ecosystem/grpc-gateway => github.com/TheThingsIndustries/grpc-gateway v1.15.2-gogo

// Use our fork of gogo/protobuf.
replace github.com/gogo/protobuf => github.com/TheThingsIndustries/gogoprotobuf v1.3.1

// Do not upgrade Protobuf beyond v1.3.5
replace github.com/golang/protobuf => github.com/golang/protobuf v1.3.5

// Do not upgrade gRPC beyond v1.33.1
replace google.golang.org/grpc => google.golang.org/grpc v1.33.1

// Do not upgrade genproto beyond v0.0.0-20200513103714-09dca8ec2884
replace google.golang.org/genproto => google.golang.org/genproto v0.0.0-20200513103714-09dca8ec2884

// Do not upgrade Echo beyond v4.1.2.
// See https://github.com/TheThingsNetwork/lorawan-stack/issues/977.
replace github.com/labstack/echo/v4 => github.com/labstack/echo/v4 v4.1.2

// Do not upgrade go-sqlmock beyond v1.3.0.
// See https://github.com/heptiolabs/healthcheck/issues/23.
replace gopkg.in/DATA-DOG/go-sqlmock.v1 => gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0

// Versions higher trigger google/protobuf update past v1.3.5.
replace gocloud.dev => gocloud.dev v0.19.0

// Versions higher trigger google/protobuf update past v1.3.5.
replace github.com/onsi/gomega => github.com/onsi/gomega v1.10.0

// Optional dependencies of throttled/v2 update golang/protobuf past v1.3.5.
replace github.com/throttled/throttled/v2 => github.com/TheThingsIndustries/throttled/v2 v2.7.1-noredis

// Do not upgrade Mapstructure beyond v1.3.0.
// See https://github.com/TheThingsNetwork/lorawan-stack/issues/3736.
replace github.com/mitchellh/mapstructure => github.com/mitchellh/mapstructure v1.3.0

// Do not upgrade Redis beyond v8.4.0.
// See https://github.com/TheThingsNetwork/lorawan-stack/pull/3848.
replace github.com/go-redis/redis/v8 => github.com/go-redis/redis/v8 v8.4.0

require (
	contrib.go.opencensus.io/exporter/prometheus v0.1.0
	github.com/PuerkitoBio/purell v1.1.1
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/TheThingsIndustries/mystique v0.0.0-20200127144137-4aa959111fe7
	github.com/TheThingsNetwork/go-cayenne-lib v1.1.0
	github.com/aws/aws-sdk-go v1.38.31
	github.com/blang/semver v3.5.1+incompatible
	github.com/blevesearch/bleve v1.0.13
	github.com/bluele/gcache v0.0.2
	github.com/chrj/smtpd v0.1.2
	github.com/cznic/b v0.0.0-20181122101859-a26611c4d92d // indirect
	github.com/disintegration/imaging v1.6.2
	github.com/dop251/goja v0.0.0-20210427212725-462d53687b0d
	github.com/dustin/go-humanize v1.0.0
	github.com/eclipse/paho.mqtt.golang v1.3.4
	github.com/envoyproxy/protoc-gen-validate v0.4.0
	github.com/felixge/httpsnoop v1.0.2
	github.com/getsentry/sentry-go v0.11.0
	github.com/go-redis/redis/v8 v8.4.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang/gddo v0.0.0-20210115222349-20d68f94ee1f
	// NOTE: github.com/golang/protobuf is actually pinned to v1.3.5 above.
	github.com/golang/protobuf v1.5.1
	github.com/google/go-cmp v0.5.5
	github.com/gorilla/csrf v1.7.0
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/websocket v1.4.2
	github.com/gotnospirit/makeplural v0.0.0-20180622080156-a5f48d94d976 // indirect
	github.com/gotnospirit/messageformat v0.0.0-20190719172517-c1d0bdacdea2
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.14.5
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/howeyc/gopass v0.0.0-20190910152052-7cb4b85ec19c
	github.com/jacobsa/crypto v0.0.0-20190317225127-9f44e2d11115
	github.com/jacobsa/oglematchers v0.0.0-20150720000706-141901ea67cd // indirect
	github.com/jacobsa/oglemock v0.0.0-20150831005832-e94d794d06ff // indirect
	github.com/jacobsa/ogletest v0.0.0-20170503003838-80d50a735a11 // indirect
	github.com/jacobsa/reqtrace v0.0.0-20150505043853-245c9e0234cb // indirect
	github.com/jarcoal/httpmock v1.0.5
	github.com/jaytaylor/html2text v0.0.0-20200412013138-3577fbdbcff7
	github.com/jinzhu/gorm v1.9.16
	github.com/jmhodges/levigo v1.0.0 // indirect
	github.com/jtacoma/uritemplates v1.0.0
	github.com/kr/pretty v0.2.1
	github.com/labstack/echo/v4 v4.1.16
	github.com/labstack/gommon v0.3.0
	github.com/lib/pq v1.10.1
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mitchellh/mapstructure v1.4.1
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/nats-io/nats-server/v2 v2.2.2
	github.com/nats-io/nats.go v1.11.0
	github.com/oklog/ulid/v2 v2.0.2
	github.com/olekukonko/tablewriter v0.0.4 // indirect
	github.com/openshift/osin v1.0.1
	github.com/pkg/browser v0.0.0-20210115035449-ce105d075bb4
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.10.0
	github.com/satori/go.uuid v1.2.0
	github.com/sendgrid/rest v2.6.3+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.9.0+incompatible
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	github.com/smartystreets/assertions v1.2.0
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/tecbot/gorocksdb v0.0.0-20191217155057-f0fad39f321c // indirect
	github.com/throttled/throttled v2.2.5+incompatible
	github.com/throttled/throttled/v2 v2.7.1
	github.com/vmihailenco/msgpack/v5 v5.3.1
	go.opencensus.io v0.23.0
	go.packetbroker.org/api/iam v1.5.16-tts
	go.packetbroker.org/api/iam/v2 v2.6.16-tts
	go.packetbroker.org/api/mapping/v2 v2.1.14-tts
	go.packetbroker.org/api/routing v1.8.7-tts
	go.packetbroker.org/api/v3 v3.10.7-tts
	go.thethings.network/lorawan-application-payload v0.0.0-20210625082552-27377194bcca
	go.thethings.network/lorawan-stack-legacy/v2 v2.0.2
	go.uber.org/zap v1.13.0
	gocloud.dev v0.20.0
	gocloud.dev/pubsub/natspubsub v0.19.0
	golang.org/x/crypto v0.0.0-20210503195802-e9a32991a82e
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e
	golang.org/x/oauth2 v0.0.0-20210427180440-81ed05c6b58c
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	google.golang.org/api v0.46.0 // indirect
	// NOTE: google.golang.org/genproto is actually pinned to v0.0.0-20200513103714-09dca8ec2884 above.
	google.golang.org/genproto v0.0.0-20210429181445-86c259c2b4ab
	// NOTE: google.golang.org/grpc is actually pinned to v1.33.1 above.
	google.golang.org/grpc v1.37.0
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.0.0-00010101000000-000000000000 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/mail.v2 v2.3.1
	gopkg.in/square/go-jose.v2 v2.5.1
	gopkg.in/yaml.v2 v2.4.0
)
