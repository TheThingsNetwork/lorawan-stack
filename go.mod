module go.thethings.network/lorawan-stack

replace github.com/grpc-ecosystem/grpc-gateway => github.com/ThethingsIndustries/grpc-gateway v0.0.0-20190109121333-7004821b0205

replace github.com/robertkrimen/otto => github.com/ThethingsIndustries/otto v0.0.0-20181129100957-6ddbbb60554a

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/PuerkitoBio/purell v1.1.0
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/RangelReale/osin v1.0.1
	github.com/TheThingsIndustries/magepkg v0.0.0-20190121105130-84da34311dab
	github.com/TheThingsIndustries/mystique v0.0.0-20181023142449-f12a32cee6d6
	github.com/TheThingsNetwork/go-cayenne-lib v1.0.0
	github.com/aws/aws-sdk-go v1.16.21
	github.com/blang/semver v3.5.1+incompatible
	github.com/certifi/gocertifi v0.0.0-20190105021004-abcd57078448 // indirect
	github.com/denisenkom/go-mssqldb v0.0.0-20190111225525-2fea367d496d // indirect
	github.com/disintegration/imaging v1.5.0
	github.com/eclipse/paho.mqtt.golang v1.1.1
	github.com/erikstmartin/go-testdb v0.0.0-20160219214506-8d10e4a1bae5 // indirect
	github.com/fsnotify/fsnotify v1.4.7
	github.com/getsentry/raven-go v0.2.0
	github.com/go-mail/mail v2.3.1+incompatible
	github.com/go-redis/redis v6.15.1+incompatible
	github.com/gobwas/glob v0.2.3
	github.com/gofrs/uuid v3.2.0+incompatible // indirect
	github.com/gogo/protobuf v1.2.0
	github.com/golang/gddo v0.0.0-20181116215533-9bd4a3295021
	github.com/golang/protobuf v1.2.0
	github.com/gorilla/securecookie v1.1.1
	github.com/gotnospirit/makeplural v0.0.0-20180622080156-a5f48d94d976 // indirect
	github.com/gotnospirit/messageformat v0.0.0-20180622080451-0eab1176a3fb
	github.com/gregjones/httpcache v0.0.0-20181110185634-c63ab54fda8f
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.6.4
	github.com/howeyc/gopass v0.0.0-20170109162249-bf9dde6d0d2c
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jacobsa/crypto v0.0.0-20180924003735-d95898ceee07
	github.com/jacobsa/oglematchers v0.0.0-20150720000706-141901ea67cd // indirect
	github.com/jacobsa/oglemock v0.0.0-20150831005832-e94d794d06ff // indirect
	github.com/jacobsa/ogletest v0.0.0-20170503003838-80d50a735a11 // indirect
	github.com/jacobsa/reqtrace v0.0.0-20150505043853-245c9e0234cb // indirect
	github.com/jaytaylor/html2text v0.0.0-20180606194806-57d518f124b0
	github.com/jinzhu/gorm v1.9.2
	github.com/jinzhu/inflection v0.0.0-20180308033659-04140366298a // indirect
	github.com/jinzhu/now v0.0.0-20181116074157-8ec929ed50c3 // indirect
	github.com/kr/pretty v0.1.0
	github.com/labstack/echo v3.3.5+incompatible
	github.com/labstack/gommon v0.2.8
	github.com/lib/pq v1.0.0
	github.com/magefile/mage v1.8.0
	github.com/mattn/go-colorable v0.0.9 // indirect
	github.com/mattn/go-isatty v0.0.4
	github.com/mattn/go-runewidth v0.0.4 // indirect
	github.com/mattn/go-sqlite3 v1.10.0 // indirect
	github.com/mitchellh/mapstructure v1.1.2
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/mwitkow/go-proto-validators v0.0.0-20180403085117-0950a7990007
	github.com/oklog/ulid v1.3.1
	github.com/olekukonko/tablewriter v0.0.1 // indirect
	github.com/onsi/ginkgo v1.7.0 // indirect
	github.com/onsi/gomega v1.4.3 // indirect
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.2
	github.com/robertkrimen/otto v0.0.0-20180617131154-15f95af6e78d
	github.com/satori/go.uuid v1.2.0
	github.com/sendgrid/rest v2.4.1+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.4.1+incompatible
	github.com/smartystreets/assertions v0.0.0-20190116191733-b6c0e53d7304
	github.com/soheilhy/cmux v0.1.4
	github.com/spf13/cast v1.3.0
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.3.1
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v0.0.0-20170224212429-dcecefd839c4 // indirect
	go.opencensus.io v0.18.0
	go.thethings.network/lorawan-stack-legacy v0.0.0-20190118141410-68812c833a78
	gocloud.dev v0.9.0
	golang.org/x/crypto v0.0.0-20190103213133-ff983b9c42bc
	golang.org/x/image v0.0.0-20190118043309-183bebdce1b2 // indirect
	golang.org/x/net v0.0.0-20190110200230-915654e7eabc
	golang.org/x/oauth2 v0.0.0-20190115181402-5dab4167f31c
	google.golang.org/genproto v0.0.0-20190111180523-db91494dd46c
	google.golang.org/grpc v1.18.0
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/jarcoal/httpmock.v1 v1.0.0-20181117152235-275e9df93516
	gopkg.in/mail.v2 v2.3.1 // indirect
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/yaml.v2 v2.2.2
)
