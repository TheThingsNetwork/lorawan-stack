module go.thethings.network/lorawan-stack/v3

go 1.17

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
	github.com/TheThingsIndustries/mystique v0.0.0-20200127144137-4aa959111fe7
	github.com/TheThingsIndustries/protoc-gen-go-json v1.1.3
	github.com/TheThingsNetwork/go-cayenne-lib v1.1.0
	github.com/aws/aws-sdk-go v1.38.31
	github.com/blang/semver v3.5.1+incompatible
	github.com/blevesearch/bleve v1.0.13
	github.com/bluele/gcache v0.0.2
	github.com/cznic/b v0.0.0-20181122101859-a26611c4d92d // indirect
	github.com/disintegration/imaging v1.6.2
	github.com/dop251/goja v0.0.0-20210427212725-462d53687b0d
	github.com/dustin/go-humanize v1.0.0
	github.com/eclipse/paho.mqtt.golang v1.3.4
	github.com/emersion/go-smtp v0.15.0
	github.com/envoyproxy/protoc-gen-validate v0.4.0
	github.com/felixge/httpsnoop v1.0.2
	github.com/getsentry/sentry-go v0.11.0
	github.com/go-redis/redis/v8 v8.4.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang/gddo v0.0.0-20210115222349-20d68f94ee1f
	// NOTE: github.com/golang/protobuf is actually pinned to v1.3.5 above.
	github.com/golang/protobuf v1.5.1
	github.com/google/go-cmp v0.5.6
	github.com/gorilla/csrf v1.7.0
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/schema v1.2.0
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
	github.com/lib/pq v1.10.1
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
	go.thethings.network/lorawan-application-payload v0.0.0-20211109090704-a9a0a6022856
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

require (
	cloud.google.com/go v0.81.0 // indirect
	cloud.google.com/go/pubsub v1.3.1 // indirect
	cloud.google.com/go/storage v1.10.0 // indirect
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/RoaringBitmap/roaring v0.4.23 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blevesearch/go-porterstemmer v1.0.3 // indirect
	github.com/blevesearch/mmap-go v1.0.2 // indirect
	github.com/blevesearch/segment v0.9.0 // indirect
	github.com/blevesearch/snowballstem v0.9.0 // indirect
	github.com/blevesearch/zap/v11 v11.0.13 // indirect
	github.com/blevesearch/zap/v12 v12.0.13 // indirect
	github.com/blevesearch/zap/v13 v13.0.5 // indirect
	github.com/blevesearch/zap/v14 v14.0.4 // indirect
	github.com/blevesearch/zap/v15 v15.0.2 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/couchbase/vellum v1.0.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dlclark/regexp2 v1.4.1-0.20201116162257-a2a8dda75c91 // indirect
	github.com/emersion/go-sasl v0.0.0-20200509203442-7bfe0ed36a21 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/glycerine/go-unsnap-stream v0.0.0-20181221182339-f9677308dec2 // indirect
	github.com/go-sourcemap/sourcemap v2.1.3+incompatible // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/uuid v1.1.2 // indirect
	github.com/google/wire v0.3.0 // indirect
	github.com/googleapis/gax-go v2.0.2+incompatible // indirect
	github.com/googleapis/gax-go/v2 v2.0.5 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.11 // indirect
	github.com/jstemmer/go-junit-report v0.9.1 // indirect
	github.com/klauspost/compress v1.11.12 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/magiconair/properties v1.8.1 // indirect
	github.com/mattn/go-runewidth v0.0.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/minio/highwayhash v1.0.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/nats-io/jwt/v2 v2.0.1 // indirect
	github.com/nats-io/nkeys v0.3.0 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/opentracing/opentracing-go v1.1.0 // indirect
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/pelletier/go-toml v1.2.0 // indirect
	github.com/philhofer/fwd v1.0.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.18.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/russross/blackfriday/v2 v2.0.1 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/spf13/afero v1.1.2 // indirect
	github.com/spf13/jwalterweatherman v1.0.0 // indirect
	github.com/steveyen/gtreap v0.1.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/tinylib/msgp v1.1.0 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/willf/bitset v1.1.10 // indirect
	go.etcd.io/bbolt v1.3.5 // indirect
	go.opentelemetry.io/otel v0.14.0 // indirect
	go.uber.org/atomic v1.5.0 // indirect
	go.uber.org/multierr v1.3.0 // indirect
	go.uber.org/tools v0.0.0-20190618225709-2cfd321de3ee // indirect
	golang.org/x/image v0.0.0-20191009234506-e7c1f5e7dbb8 // indirect
	golang.org/x/lint v0.0.0-20201208152925-83fdc39ff7b5 // indirect
	golang.org/x/mod v0.4.1 // indirect
	golang.org/x/sys v0.0.0-20210616094352-59db8d763f22 // indirect
	golang.org/x/term v0.0.0-20201126162022-7de9c90e9dd1 // indirect
	golang.org/x/text v0.3.6 // indirect
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	golang.org/x/tools v0.1.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/ini.v1 v1.51.1 // indirect
	honnef.co/go/tools v0.0.1-2020.1.4 // indirect
)
