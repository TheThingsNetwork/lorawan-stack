module go.thethings.network/lorawan-stack/v3

go 1.21

// See https://github.com/mitchellh/mapstructure/pull/278
replace github.com/mitchellh/mapstructure => github.com/TheThingsIndustries/mapstructure v0.0.0-20230413130846-941bcd1deec3

require (
	github.com/KimMachineGun/automemlimit v0.6.1
	github.com/Masterminds/sprig/v3 v3.2.3
	github.com/TheThingsIndustries/mystique v0.0.0-20221125120501-80ab21781b6d
	github.com/TheThingsIndustries/protoc-gen-go-flags v1.2.0
	github.com/TheThingsIndustries/protoc-gen-go-json v1.6.0
	github.com/TheThingsNetwork/go-cayenne-lib v1.2.0
	github.com/aws/aws-sdk-go v1.54.3
	github.com/blang/semver v3.5.1+incompatible
	github.com/blevesearch/bleve v1.0.14
	github.com/bluele/gcache v0.0.2
	github.com/disintegration/imaging v1.6.2
	github.com/dop251/goja v0.0.0-20231027120936-b396bb4c349d
	github.com/dustin/go-humanize v1.0.1
	github.com/eclipse/paho.mqtt.golang v1.4.4-0.20230810210351-7b759f1d9ddf
	github.com/emersion/go-smtp v0.21.2
	github.com/envoyproxy/protoc-gen-validate v1.0.4
	github.com/felixge/httpsnoop v1.0.4
	github.com/getsentry/sentry-go v0.28.1
	github.com/golang/gddo v0.0.0-20210115222349-20d68f94ee1f
	github.com/google/go-cmp v0.6.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/csrf v1.7.2
	github.com/gorilla/handlers v1.5.2
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/schema v1.4.0
	github.com/gorilla/securecookie v1.1.2
	github.com/gorilla/websocket v1.5.3
	github.com/gotnospirit/messageformat v0.0.0-20221001023931-dfe49f1eb092
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0
	github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus v1.0.1
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0
	github.com/hellofresh/health-go/v5 v5.5.3
	github.com/howeyc/gopass v0.0.0-20210920133722-c8aef6fb66ef
	github.com/iancoleman/strcase v0.3.0
	github.com/jackc/pgerrcode v0.0.0-20220416144525-469b46aa5efa
	github.com/jackc/pgx/v5 v5.6.0
	github.com/jacobsa/crypto v0.0.0-20190317225127-9f44e2d11115
	github.com/jarcoal/httpmock v1.3.1
	github.com/jaytaylor/html2text v0.0.0-20230321000545-74c2419ad056
	github.com/json-iterator/go v1.1.12
	github.com/jtacoma/uritemplates v1.0.0
	github.com/klauspost/compress v1.17.9
	github.com/kr/pretty v0.3.1
	github.com/lib/pq v1.10.9
	github.com/mileusna/useragent v1.3.4
	github.com/mitchellh/mapstructure v1.5.0
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/nats-io/nats-server/v2 v2.10.16
	github.com/nats-io/nats.go v1.36.0
	github.com/oklog/ulid/v2 v2.1.0
	github.com/openshift/osin v1.0.2-0.20220317075346-0f4d38c6e53f
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.19.1
	github.com/redis/go-redis/v9 v9.5.3
	github.com/sendgrid/sendgrid-go v3.14.0+incompatible
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	github.com/smarty/assertions v1.16.0
	github.com/spf13/cast v1.6.0
	github.com/spf13/cobra v1.8.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.19.0
	github.com/throttled/throttled/v2 v2.12.0
	github.com/uptrace/bun v1.2.1
	github.com/uptrace/bun/dialect/pgdialect v1.2.1
	github.com/uptrace/bun/driver/pgdriver v1.2.1
	github.com/vmihailenco/msgpack/v5 v5.4.1
	go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux v0.52.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.52.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.52.0
	go.opentelemetry.io/otel v1.27.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.27.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.27.0
	go.opentelemetry.io/otel/metric v1.27.0
	go.opentelemetry.io/otel/sdk v1.27.0
	go.opentelemetry.io/otel/trace v1.27.0
	go.packetbroker.org/api/iam v1.8.2
	go.packetbroker.org/api/iam/v2 v2.9.1
	go.packetbroker.org/api/mapping/v2 v2.3.2
	go.packetbroker.org/api/routing v1.9.2
	go.packetbroker.org/api/v3 v3.17.1
	go.thethings.network/lorawan-application-payload v0.0.0-20220125153912-1198ff1e403e
	go.thethings.network/lorawan-stack-legacy/v2 v2.1.0
	go.uber.org/automaxprocs v1.5.3
	go.uber.org/zap v1.27.0
	gocloud.dev v0.37.0
	gocloud.dev/pubsub/natspubsub v0.37.0
	golang.org/x/crypto v0.24.0
	golang.org/x/exp v0.0.0-20240119083558-1b970713d09a
	golang.org/x/net v0.26.0
	golang.org/x/oauth2 v0.21.0
	golang.org/x/sync v0.7.0
	google.golang.org/genproto v0.0.0-20240311173647-c811ad7063a7
	google.golang.org/genproto/googleapis/api v0.0.0-20240520151616-dc85e6b867a5
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240520151616-dc85e6b867a5
	google.golang.org/grpc v1.64.0
	google.golang.org/protobuf v1.34.2
	gopkg.in/mail.v2 v2.3.1
	gopkg.in/square/go-jose.v2 v2.6.0
	gopkg.in/yaml.v2 v2.4.0
	nhooyr.io/websocket v1.8.11
)

require (
	cloud.google.com/go v0.112.1 // indirect
	cloud.google.com/go/compute/metadata v0.3.0 // indirect
	cloud.google.com/go/iam v1.1.6 // indirect
	cloud.google.com/go/pubsub v1.37.0 // indirect
	cloud.google.com/go/storage v1.39.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.11.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.6.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.8.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.3.1 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.2.2 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.2.0 // indirect
	github.com/RoaringBitmap/roaring v0.4.23 // indirect
	github.com/aws/aws-sdk-go-v2 v1.25.3 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.1 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.27.7 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.7 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.15.3 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.16.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.3.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.17.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.51.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sns v1.29.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sqs v1.31.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.20.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.23.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.28.4 // indirect
	github.com/aws/smithy-go v1.20.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blevesearch/go-porterstemmer v1.0.3 // indirect
	github.com/blevesearch/mmap-go v1.0.2 // indirect
	github.com/blevesearch/segment v0.9.0 // indirect
	github.com/blevesearch/snowballstem v0.9.0 // indirect
	github.com/blevesearch/zap/v11 v11.0.14 // indirect
	github.com/blevesearch/zap/v12 v12.0.14 // indirect
	github.com/blevesearch/zap/v13 v13.0.6 // indirect
	github.com/blevesearch/zap/v14 v14.0.5 // indirect
	github.com/blevesearch/zap/v15 v15.0.3 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cilium/ebpf v0.9.1 // indirect
	github.com/containerd/cgroups/v3 v3.0.1 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/couchbase/vellum v1.0.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dlclark/regexp2 v1.7.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/emersion/go-sasl v0.0.0-20200509203442-7bfe0ed36a21 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/glycerine/go-unsnap-stream v0.0.0-20181221182339-f9677308dec2 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-sourcemap/sourcemap v2.1.3+incompatible // indirect
	github.com/godbus/dbus/v5 v5.0.4 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/pprof v0.0.0-20230207041349-798e818bf904 // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/google/wire v0.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.12.3 // indirect
	github.com/gotnospirit/makeplural v0.0.0-20180622080156-a5f48d94d976 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.1.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/huandu/xstrings v1.3.3 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20231201235250-de7065d80cb9 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jacobsa/oglematchers v0.0.0-20150720000706-141901ea67cd // indirect
	github.com/jacobsa/oglemock v0.0.0-20150831005832-e94d794d06ff // indirect
	github.com/jacobsa/ogletest v0.0.0-20170503003838-80d50a735a11 // indirect
	github.com/jacobsa/reqtrace v0.0.0-20150505043853-245c9e0234cb // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/nats-io/jwt/v2 v2.5.7 // indirect
	github.com/nats-io/nkeys v0.4.7 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/opencontainers/runtime-spec v1.0.2 // indirect
	github.com/pbnjay/memory v0.0.0-20210728143218-7b4eea64cf58 // indirect
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/philhofer/fwd v1.0.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sendgrid/rest v2.6.9+incompatible // indirect
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/steveyen/gtreap v0.1.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tinylib/msgp v1.1.0 // indirect
	github.com/tmthrgd/go-hex v0.0.0-20190904060850-447a3041c3bc // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/willf/bitset v1.1.10 // indirect
	go.etcd.io/bbolt v1.3.8 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.27.0 // indirect
	go.opentelemetry.io/proto/otlp v1.2.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/image v0.18.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/term v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	golang.org/x/xerrors v0.0.0-20231012003039-104605ab7028 // indirect
	google.golang.org/api v0.171.0 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	mellium.im/sasl v0.3.1 // indirect
)
