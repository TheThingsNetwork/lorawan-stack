module go.thethings.network/lorawan-stack/v3

go 1.18

// Use our fork of grpc-gateway.
replace github.com/grpc-ecosystem/grpc-gateway => github.com/TheThingsIndustries/grpc-gateway v1.15.2-gogo

// But the original grpc-gateway v2.
replace github.com/grpc-ecosystem/grpc-gateway/v2 => github.com/grpc-ecosystem/grpc-gateway/v2 v2.10.3

// Use our fork of gogo/protobuf.
replace github.com/gogo/protobuf => github.com/TheThingsIndustries/gogoprotobuf v1.3.1

// Use our fork of throttled/throttled/v2.
replace github.com/throttled/throttled/v2 => github.com/TheThingsIndustries/throttled/v2 v2.7.1-noredis

// Pin dependencies that would break because of our old golang/protobuf.
replace (
	cloud.google.com/go => cloud.google.com/go v0.81.0
	cloud.google.com/go/pubsub => cloud.google.com/go/pubsub v1.3.1
	cloud.google.com/go/storage => cloud.google.com/go/storage v1.16.0
	github.com/Azure/azure-storage-blob-go => github.com/Azure/azure-storage-blob-go v0.10.0
	github.com/golang/protobuf => github.com/golang/protobuf v1.3.5
	github.com/googleapis/gax-go/v2 => github.com/googleapis/gax-go/v2 v2.0.5
	github.com/onsi/gomega => github.com/onsi/gomega v1.10.0
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.11.0
	github.com/spf13/cobra => github.com/spf13/cobra v1.2.1
	github.com/spf13/viper => github.com/spf13/viper v1.8.1
	gocloud.dev => gocloud.dev v0.19.0
	gocloud.dev/pubsub/natspubsub => gocloud.dev/pubsub/natspubsub v0.19.0
	google.golang.org/api => google.golang.org/api v0.53.0
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20200513103714-09dca8ec2884
	google.golang.org/grpc => google.golang.org/grpc v1.33.1
)

// Do not upgrade go-sqlmock beyond v1.3.0.
// See https://github.com/heptiolabs/healthcheck/issues/23.
replace gopkg.in/DATA-DOG/go-sqlmock.v1 => gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0

// See https://github.com/mattn/go-ieproxy/issues/31
replace github.com/mattn/go-ieproxy => github.com/mattn/go-ieproxy v0.0.1

// See https://github.com/mitchellh/mapstructure/pull/278
replace github.com/mitchellh/mapstructure v1.4.3 => github.com/TheThingsIndustries/mapstructure v0.0.0-20220329135826-c42f9f170b2a

require (
	contrib.go.opencensus.io/exporter/prometheus v0.4.0
	// NOTE: github.com/Azure/azure-storage-blob-go is actually a different version (see above).
	github.com/Azure/azure-storage-blob-go v0.10.0
	github.com/Azure/go-autorest/autorest v0.11.24
	github.com/Azure/go-autorest/autorest/adal v0.9.18
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/TheThingsIndustries/mystique v0.0.0-20211230093812-d4088bd06959
	github.com/TheThingsIndustries/protoc-gen-go-flags v1.0.0
	github.com/TheThingsIndustries/protoc-gen-go-json v1.4.0
	github.com/TheThingsNetwork/go-cayenne-lib v1.1.0
	github.com/aws/aws-sdk-go v1.42.53
	github.com/blang/semver v3.5.1+incompatible
	github.com/blevesearch/bleve v1.0.14
	github.com/bluele/gcache v0.0.2
	github.com/disintegration/imaging v1.6.2
	github.com/dop251/goja v0.0.0-20220214123719-b09a6bfa842f
	github.com/dustin/go-humanize v1.0.0
	github.com/eclipse/paho.mqtt.golang v1.3.5
	github.com/emersion/go-smtp v0.15.0
	github.com/envoyproxy/protoc-gen-validate v0.6.3
	github.com/felixge/httpsnoop v1.0.2
	github.com/getsentry/sentry-go v0.12.0
	github.com/go-redis/redis/v8 v8.11.4
	github.com/gogo/protobuf v1.3.2
	github.com/golang/gddo v0.0.0-20210115222349-20d68f94ee1f
	// NOTE: github.com/golang/protobuf is actually a different version (see above).
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.8
	github.com/gorilla/csrf v1.7.1
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/schema v1.2.0
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/websocket v1.4.2
	github.com/gotnospirit/messageformat v0.0.0-20190719172517-c1d0bdacdea2
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.0.0-00010101000000-000000000000
	github.com/heptiolabs/healthcheck v0.0.0-20211123025425-613501dd5deb
	github.com/howeyc/gopass v0.0.0-20210920133722-c8aef6fb66ef
	github.com/iancoleman/strcase v0.2.0
	github.com/jackc/pgconn v1.12.1
	github.com/jackc/pgerrcode v0.0.0-20220416144525-469b46aa5efa
	github.com/jackc/pgx/v4 v4.16.1
	github.com/jacobsa/crypto v0.0.0-20190317225127-9f44e2d11115
	github.com/jarcoal/httpmock v1.1.0
	github.com/jaytaylor/html2text v0.0.0-20211105163654-bc68cce691ba
	github.com/jinzhu/gorm v1.9.16
	github.com/jtacoma/uritemplates v1.0.0
	github.com/kr/pretty v0.3.0
	github.com/lib/pq v1.10.4
	github.com/mitchellh/mapstructure v1.4.3
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/nats-io/nats-server/v2 v2.7.4
	github.com/nats-io/nats.go v1.13.1-0.20220308171302-2f2f6968e98d
	github.com/oklog/ulid/v2 v2.0.2
	github.com/openshift/osin v1.0.1
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/satori/go.uuid v1.2.0
	github.com/sendgrid/sendgrid-go v3.11.0+incompatible
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	github.com/smartystreets/assertions v1.2.1
	github.com/spf13/cast v1.4.1
	// NOTE: github.com/spf13/cobra is actually a different version (see above).
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	// NOTE: github.com/spf13/viper is actually a different version (see above).
	github.com/spf13/viper v1.10.1
	github.com/throttled/throttled v2.2.5+incompatible
	github.com/throttled/throttled/v2 v2.7.1
	github.com/uptrace/bun v1.1.6
	github.com/uptrace/bun/dialect/pgdialect v1.1.6
	github.com/uptrace/bun/driver/pgdriver v1.1.6
	github.com/vmihailenco/msgpack/v5 v5.3.5
	go.opencensus.io v0.23.0
	go.opentelemetry.io/otel v1.7.0
	go.opentelemetry.io/otel/trace v1.7.0
	go.packetbroker.org/api/iam v1.5.27-tts
	go.packetbroker.org/api/iam/v2 v2.7.8-tts
	go.packetbroker.org/api/mapping/v2 v2.1.27-tts
	go.packetbroker.org/api/routing v1.8.18-tts
	go.packetbroker.org/api/v3 v3.12.4-tts
	go.thethings.network/lorawan-application-payload v0.0.0-20220125153912-1198ff1e403e
	go.thethings.network/lorawan-stack-legacy/v2 v2.0.2
	go.uber.org/zap v1.21.0
	// NOTE: gocloud.dev is actually a different version (see above).
	gocloud.dev v0.20.0
	// NOTE: gocloud.dev/pubsub/natspubsub is actually a different version (see above).
	gocloud.dev/pubsub/natspubsub v0.19.0
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d
	golang.org/x/exp v0.0.0-20220706164943-b4a6d9510983
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd
	golang.org/x/oauth2 v0.0.0-20220411215720-9780585627b5
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	// NOTE: google.golang.org/genproto is actually a different version (see above).
	google.golang.org/genproto v0.0.0-20220519153652-3a47de7e79bd
	// NOTE: google.golang.org/grpc is actually a different version (see above).
	google.golang.org/grpc v1.46.2
	gopkg.in/mail.v2 v2.3.1
	gopkg.in/square/go-jose.v2 v2.6.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	cloud.google.com/go v0.90.0 // indirect
	cloud.google.com/go/pubsub v1.3.1 // indirect
	cloud.google.com/go/storage v1.16.0 // indirect
	github.com/Azure/azure-pipeline-go v0.2.3 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.10 // indirect
	github.com/Azure/go-autorest/autorest/azure/cli v0.4.4 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.1.1 // indirect
	github.com/RoaringBitmap/roaring v0.9.4 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bits-and-blooms/bitset v1.2.1 // indirect
	github.com/blevesearch/go-porterstemmer v1.0.3 // indirect
	github.com/blevesearch/mmap-go v1.0.3 // indirect
	github.com/blevesearch/segment v0.9.0 // indirect
	github.com/blevesearch/snowballstem v0.9.0 // indirect
	github.com/blevesearch/zap/v11 v11.0.14 // indirect
	github.com/blevesearch/zap/v12 v12.0.14 // indirect
	github.com/blevesearch/zap/v13 v13.0.6 // indirect
	github.com/blevesearch/zap/v14 v14.0.5 // indirect
	github.com/blevesearch/zap/v15 v15.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/couchbase/vellum v1.0.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dlclark/regexp2 v1.4.1-0.20201116162257-a2a8dda75c91 // indirect
	github.com/emersion/go-sasl v0.0.0-20211008083017-0b9dcfb154ac // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/go-kit/log v0.2.0 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-sourcemap/sourcemap v2.1.3+incompatible // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/golang-jwt/jwt/v4 v4.3.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.3.0
	github.com/google/wire v0.5.0 // indirect
	github.com/googleapis/gax-go v2.0.2+incompatible // indirect
	github.com/googleapis/gax-go/v2 v2.1.1 // indirect
	github.com/gotnospirit/makeplural v0.0.0-20180622080156-a5f48d94d976 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/huandu/xstrings v1.3.1 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.11.0 // indirect
	github.com/jacobsa/oglematchers v0.0.0-20150720000706-141901ea67cd // indirect
	github.com/jacobsa/oglemock v0.0.0-20150831005832-e94d794d06ff // indirect
	github.com/jacobsa/ogletest v0.0.0-20170503003838-80d50a735a11 // indirect
	github.com/jacobsa/reqtrace v0.0.0-20150505043853-245c9e0234cb // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/jstemmer/go-junit-report v0.9.1 // indirect
	github.com/klauspost/compress v1.14.4 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mattn/go-ieproxy v0.0.3 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/mattn/go-sqlite3 v1.14.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/nats-io/jwt/v2 v2.2.1-0.20220113022732-58e87895b296 // indirect
	github.com/nats-io/nkeys v0.3.0 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pborman/uuid v1.2.1 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/prometheus/statsd_exporter v0.22.4 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.8.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sendgrid/rest v2.6.8+incompatible // indirect
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/spf13/afero v1.8.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/steveyen/gtreap v0.1.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/tmthrgd/go-hex v0.0.0-20190904060850-447a3041c3bc // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/willf/bitset v1.1.11 // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	golang.org/x/image v0.0.0-20211028202545-6944b10bf410 // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220106191415-9b9b3d81d5e3 // indirect
	golang.org/x/sys v0.0.0-20220708085239-5a0f0661e09d // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20220210224613-90d013bbcef8 // indirect
	golang.org/x/tools v0.1.10 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/api v0.61.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	// NOTE: gopkg.in/DATA-DOG/go-sqlmock.v1 is actually a different version (see above).
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.0.0-00010101000000-000000000000 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/ini.v1 v1.66.4 // indirect
	mellium.im/sasl v0.2.1 // indirect
)
