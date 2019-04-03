module go.thethings.network/lorawan-stack

replace github.com/grpc-ecosystem/grpc-gateway => github.com/ThethingsIndustries/grpc-gateway v1.7.0-gogo

replace github.com/robertkrimen/otto => github.com/ThethingsIndustries/otto v0.0.0-20181129100957-6ddbbb60554a

replace github.com/golang/lint => golang.org/x/lint v0.0.0-20190313153728-d0100b6bd8b3

replace github.com/testcontainers/testcontainer-go => github.com/testcontainers/testcontainers-go v0.0.2

require (
	cloud.google.com/go v0.37.0 // indirect
	github.com/Azure/azure-storage-blob-go v0.0.0-20190123011202-457680cc0804 // indirect
	github.com/PuerkitoBio/purell v1.1.1
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/TheThingsIndustries/magepkg v0.0.0-20190214092847-6c0299b7c3ed
	github.com/TheThingsIndustries/mystique v0.0.0-20181023142449-f12a32cee6d6
	github.com/TheThingsNetwork/go-cayenne-lib v1.0.0
	github.com/aws/aws-sdk-go v1.18.3
	github.com/blang/semver v3.5.1+incompatible
	github.com/certifi/gocertifi v0.0.0-20190105021004-abcd57078448 // indirect
	github.com/client9/misspell v0.3.4
	github.com/cpuguy83/go-md2man v1.0.10 // indirect
	github.com/denisenkom/go-mssqldb v0.0.0-20190315220205-a8ed825ac853 // indirect
	github.com/disintegration/imaging v1.6.0
	github.com/eclipse/paho.mqtt.golang v1.1.1
	github.com/erikstmartin/go-testdb v0.0.0-20160219214506-8d10e4a1bae5 // indirect
	github.com/fsnotify/fsnotify v1.4.7
	github.com/getsentry/raven-go v0.2.0
	github.com/go-mail/mail v2.3.1+incompatible
	github.com/go-redis/redis v6.15.2+incompatible
	github.com/gobwas/glob v0.2.3
	github.com/gofrs/uuid v3.2.0+incompatible // indirect
	github.com/gogo/protobuf v1.2.1
	github.com/golang/gddo v0.0.0-20190312205958-5a2505f3dbf0
	github.com/golang/protobuf v1.3.1
	github.com/google/uuid v1.1.1 // indirect
	github.com/googleapis/gax-go/v2 v2.0.4 // indirect
	github.com/goreleaser/goreleaser v0.103.1
	github.com/gorilla/securecookie v1.1.1
	github.com/gotnospirit/makeplural v0.0.0-20180622080156-a5f48d94d976 // indirect
	github.com/gotnospirit/messageformat v0.0.0-20180622080451-0eab1176a3fb
	github.com/gregjones/httpcache v0.0.0-20190212212710-3befbb6ad0cc
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.8.5
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/howeyc/gopass v0.0.0-20170109162249-bf9dde6d0d2c
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/jacobsa/crypto v0.0.0-20180924003735-d95898ceee07
	github.com/jacobsa/oglematchers v0.0.0-20150720000706-141901ea67cd // indirect
	github.com/jacobsa/oglemock v0.0.0-20150831005832-e94d794d06ff // indirect
	github.com/jacobsa/ogletest v0.0.0-20170503003838-80d50a735a11 // indirect
	github.com/jacobsa/reqtrace v0.0.0-20150505043853-245c9e0234cb // indirect
	github.com/jarcoal/httpmock v1.0.0
	github.com/jaytaylor/html2text v0.0.0-20190311042500-a93a6c6ea053
	github.com/jinzhu/gorm v1.9.2
	github.com/jinzhu/inflection v0.0.0-20180308033659-04140366298a // indirect
	github.com/jinzhu/now v1.0.0 // indirect
	github.com/kamilsk/retry/v4 v4.0.2 // indirect
	github.com/kr/pretty v0.1.0
	github.com/labstack/echo/v4 v4.0.0
	github.com/labstack/gommon v0.2.8
	github.com/lib/pq v1.0.0
	github.com/lyft/protoc-gen-validate v0.0.13
	github.com/magefile/mage v1.8.1-0.20190314142316-8dce728c572d
	github.com/mattn/go-colorable v0.1.1 // indirect
	github.com/mattn/go-isatty v0.0.7
	github.com/mattn/go-runewidth v0.0.4 // indirect
	github.com/mattn/go-sqlite3 v1.10.0 // indirect
	github.com/mattn/go-zglob v0.0.1 // indirect
	github.com/mattn/goveralls v0.0.2
	github.com/mdempsky/unconvert v0.0.0-20190117010209-2db5a8ead8e7
	github.com/mgechev/dots v0.0.0-20181228164730-18fa4c4b71cc // indirect
	github.com/mgechev/revive v0.0.0-20190301194522-6a62ee9f0248
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.1.2
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/oklog/ulid v1.3.1
	github.com/olekukonko/tablewriter v0.0.1 // indirect
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/openshift/osin v1.0.1
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.3-0.20190127221311-3c4408c8b829
	github.com/prometheus/client_model v0.0.0-20190129233127-fd36f4220a90 // indirect
	github.com/prometheus/procfs v0.0.0-20190315082738-e56f2e22fc76 // indirect
	github.com/robertkrimen/otto v0.0.0-20180617131154-15f95af6e78d
	github.com/satori/go.uuid v1.2.0
	github.com/sendgrid/rest v2.4.1+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.4.1+incompatible
	github.com/smartystreets/assertions v0.0.0-20190215210624-980c5ac6f3ac
	github.com/soheilhy/cmux v0.1.4
	github.com/spf13/afero v1.2.1 // indirect
	github.com/spf13/cast v1.3.0
	github.com/spf13/cobra v0.0.3
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.3.2
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/valyala/fasttemplate v1.0.1 // indirect
	go.opencensus.io v0.19.1
	go.thethings.network/lorawan-stack-legacy v0.0.0-20190118141410-68812c833a78
	gocloud.dev v0.11.0
	golang.org/x/crypto v0.0.0-20190313024323-a1f597ede03a
	golang.org/x/image v0.0.0-20190227222117-0694c2d4d067 // indirect
	golang.org/x/net v0.0.0-20190313220215-9f648a60d977
	golang.org/x/oauth2 v0.0.0-20190226205417-e64efc72b421
	golang.org/x/sys v0.0.0-20190316082340-a2f829d7f35f // indirect
	golang.org/x/tools v0.0.0-20190315214010-f0bfdbff1f9c
	golang.org/x/xerrors v0.0.0-20190315151331-d61658bd2e18 // indirect
	google.golang.org/api v0.2.0 // indirect
	google.golang.org/genproto v0.0.0-20190307195333-5fe7a883aa19
	google.golang.org/grpc v1.19.0
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/mail.v2 v2.3.1 // indirect
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/square/go-jose.v2 v2.3.0
	gopkg.in/yaml.v2 v2.2.2
)
