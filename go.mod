module go.thethings.network/lorawan-stack

replace github.com/grpc-ecosystem/grpc-gateway => github.com/ThethingsIndustries/grpc-gateway v1.7.0-gogo

replace github.com/robertkrimen/otto => github.com/ThethingsIndustries/otto v0.0.0-20181129100957-6ddbbb60554a

replace github.com/alecthomas/gometalinter => github.com/alecthomas/gometalinter v3.0.0+incompatible

replace gopkg.in/alecthomas/kingpin.v3-unstable => gopkg.in/alecthomas/kingpin.v3-unstable v3.0.0-20171010053543-63abe20a23e2

require (
	github.com/PuerkitoBio/purell v1.1.1
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/RangelReale/osin v1.0.1
	github.com/TheThingsIndustries/magepkg v0.0.0-20190214092847-6c0299b7c3ed
	github.com/TheThingsIndustries/mystique v0.0.0-20181023142449-f12a32cee6d6
	github.com/TheThingsNetwork/go-cayenne-lib v1.0.0
	github.com/alecthomas/gometalinter v3.0.0+incompatible
	github.com/aws/aws-sdk-go v1.17.5
	github.com/blang/semver v3.5.1+incompatible
	github.com/certifi/gocertifi v0.0.0-20190105021004-abcd57078448 // indirect
	github.com/client9/misspell v0.3.4
	github.com/disintegration/imaging v1.6.0
	github.com/eclipse/paho.mqtt.golang v1.1.1
	github.com/erikstmartin/go-testdb v0.0.0-20160219214506-8d10e4a1bae5 // indirect
	github.com/fsnotify/fsnotify v1.4.7
	github.com/getsentry/raven-go v0.2.0
	github.com/go-mail/mail v2.3.1+incompatible
	github.com/go-redis/redis v6.15.1+incompatible
	github.com/gobwas/glob v0.2.3
	github.com/gofrs/uuid v3.2.0+incompatible // indirect
	github.com/gogo/protobuf v1.2.1
	github.com/golang/gddo v0.0.0-20181116215533-9bd4a3295021
	github.com/golang/protobuf v1.2.0
	github.com/google/shlex v0.0.0-20181106134648-c34317bd91bf // indirect
	github.com/goreleaser/goreleaser v0.101.0
	github.com/gorilla/securecookie v1.1.1
	github.com/gotnospirit/makeplural v0.0.0-20180622080156-a5f48d94d976 // indirect
	github.com/gotnospirit/messageformat v0.0.0-20180622080451-0eab1176a3fb
	github.com/gregjones/httpcache v0.0.0-20190212212710-3befbb6ad0cc
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.6.2
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/howeyc/gopass v0.0.0-20170109162249-bf9dde6d0d2c
	github.com/jacobsa/crypto v0.0.0-20180924003735-d95898ceee07
	github.com/jacobsa/oglematchers v0.0.0-20150720000706-141901ea67cd // indirect
	github.com/jacobsa/oglemock v0.0.0-20150831005832-e94d794d06ff // indirect
	github.com/jacobsa/ogletest v0.0.0-20170503003838-80d50a735a11 // indirect
	github.com/jacobsa/reqtrace v0.0.0-20150505043853-245c9e0234cb // indirect
	github.com/jaytaylor/html2text v0.0.0-20180606194806-57d518f124b0
	github.com/jinzhu/gorm v1.9.2
	github.com/jinzhu/inflection v0.0.0-20180308033659-04140366298a // indirect
	github.com/jinzhu/now v1.0.0 // indirect
	github.com/kr/pretty v0.1.0
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/gommon v0.2.8
	github.com/lib/pq v1.0.0
	github.com/magefile/mage v1.8.0
	github.com/mattn/go-isatty v0.0.6
	github.com/mattn/go-sqlite3 v1.10.0 // indirect
	github.com/mattn/goveralls v0.0.2
	github.com/mdempsky/unconvert v0.0.0-20190117010209-2db5a8ead8e7
	github.com/mitchellh/mapstructure v1.1.2
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/mwitkow/go-proto-validators v0.0.0-20190212092829-1f388280e944
	github.com/nicksnyder/go-i18n v1.10.0 // indirect
	github.com/oklog/ulid v1.3.1
	github.com/olekukonko/tablewriter v0.0.1 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.2
	github.com/robertkrimen/otto v0.0.0-20181129100957-6ddbbb60554a
	github.com/satori/go.uuid v1.2.0
	github.com/sendgrid/rest v2.4.1+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.4.1+incompatible
	github.com/smartystreets/assertions v0.0.0-20190215210624-980c5ac6f3ac
	github.com/soheilhy/cmux v0.1.4
	github.com/spf13/cast v1.3.0
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.3.1
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v0.0.0-20170224212429-dcecefd839c4 // indirect
	go.opencensus.io v0.19.0
	go.thethings.network/lorawan-stack-legacy v0.0.0-20190118141410-68812c833a78
	gocloud.dev v0.10.0
	golang.org/x/crypto v0.0.0-20190225124518-7f87c0fbb88b
	golang.org/x/net v0.0.0-20190225153610-fe579d43d832
	golang.org/x/oauth2 v0.0.0-20190220154721-9b3c75971fc9
	golang.org/x/tools v0.0.0-20190225234524-2dc4ef2775b8
	google.golang.org/genproto v0.0.0-20190219182410-082222b4a5c5
	google.golang.org/grpc v1.18.0
	gopkg.in/alecthomas/kingpin.v3-unstable v3.0.0-20171010053543-63abe20a23e2 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/jarcoal/httpmock.v1 v1.0.0-20190204112747-618f46f3f0c8
	gopkg.in/mail.v2 v2.3.1 // indirect
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/square/go-jose.v2 v2.2.2
	gopkg.in/yaml.v2 v2.2.2
)
