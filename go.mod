module go.thethings.network/lorawan-stack

replace github.com/grpc-ecosystem/grpc-gateway => github.com/ThethingsIndustries/grpc-gateway v1.7.0-gogo

replace github.com/robertkrimen/otto => github.com/ThethingsIndustries/otto v0.0.0-20181129100957-6ddbbb60554a

replace github.com/labstack/echo/v4 => github.com/TheThingsIndustries/echo/v4 v4.0.1-0.20190409124425-ee570f243713

replace github.com/golang/lint => golang.org/x/lint v0.0.0-20190313153728-d0100b6bd8b3

replace github.com/testcontainers/testcontainer-go => github.com/testcontainers/testcontainers-go v0.0.2

require (
	github.com/PuerkitoBio/purell v1.1.1
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/TheThingsIndustries/magepkg v0.0.0-20190214092847-6c0299b7c3ed
	github.com/TheThingsIndustries/mystique v0.0.0-20181023142449-f12a32cee6d6
	github.com/TheThingsNetwork/go-cayenne-lib v1.0.0
	github.com/aws/aws-sdk-go v1.19.12
	github.com/blang/semver v3.5.1+incompatible
	github.com/certifi/gocertifi v0.0.0-20190415143156-92f724a62f3e // indirect
	github.com/client9/misspell v0.3.4
	github.com/cpuguy83/go-md2man v1.0.10 // indirect
	github.com/disintegration/imaging v1.6.0
	github.com/eclipse/paho.mqtt.golang v1.1.1
	github.com/fsnotify/fsnotify v1.4.7
	github.com/getsentry/raven-go v0.2.0
	github.com/go-mail/mail v2.3.1+incompatible
	github.com/go-redis/redis v6.15.2+incompatible
	github.com/gobwas/glob v0.2.3
	github.com/gogo/protobuf v1.2.1
	github.com/golang/gddo v0.0.0-20190312205958-5a2505f3dbf0
	github.com/golang/protobuf v1.3.1
	github.com/goreleaser/goreleaser v0.105.0
	github.com/gorilla/securecookie v1.1.1
	github.com/gotnospirit/makeplural v0.0.0-20180622080156-a5f48d94d976 // indirect
	github.com/gotnospirit/messageformat v0.0.0-20180622080451-0eab1176a3fb
	github.com/gregjones/httpcache v0.0.0-20190212212710-3befbb6ad0cc
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.6.2
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/howeyc/gopass v0.0.0-20170109162249-bf9dde6d0d2c
	github.com/jacobsa/crypto v0.0.0-20190317225127-9f44e2d11115
	github.com/jarcoal/httpmock v1.0.3
	github.com/jaytaylor/html2text v0.0.0-20190408195923-01ec452cbe43
	github.com/jinzhu/gorm v1.9.4
	github.com/kr/pretty v0.1.0
	github.com/labstack/echo/v4 v4.0.0-00010101000000-000000000000
	github.com/labstack/gommon v0.2.8
	github.com/lib/pq v1.1.0
	github.com/lyft/protoc-gen-validate v0.0.14
	github.com/magefile/mage v1.8.0
	github.com/mattn/go-isatty v0.0.7
	github.com/mattn/goveralls v0.0.2
	github.com/mdempsky/unconvert v0.0.0-20190325185700-2f5dc3378ed3
	github.com/mgechev/revive v0.0.0-20190416071613-796760d728e1
	github.com/mitchellh/mapstructure v1.1.2
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/oklog/ulid v1.3.1
	github.com/openshift/osin v1.0.1
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.3-0.20190127221311-3c4408c8b829
	github.com/robertkrimen/otto v0.0.0-00010101000000-000000000000
	github.com/satori/go.uuid v1.2.0
	github.com/sendgrid/rest v2.4.1+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.4.1+incompatible
	github.com/smartystreets/assertions v0.0.0-20190401211740-f487f9de1cd3
	github.com/soheilhy/cmux v0.1.4
	github.com/spf13/cast v1.3.0
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.3.2
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	go.opencensus.io v0.20.2
	go.thethings.network/lorawan-stack-legacy v0.0.0-20190118141410-68812c833a78
	gocloud.dev v0.12.0
	golang.org/x/crypto v0.0.0-20190411191339-88737f569e3a
	golang.org/x/net v0.0.0-20190415214537-1da14a5a36f2
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a
	golang.org/x/tools v0.0.0-20190417005754-4ca4b55e2050
	google.golang.org/genproto v0.0.0-20190415143225-d1146b9035b9
	google.golang.org/grpc v1.20.0
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/square/go-jose.v2 v2.3.1
	gopkg.in/yaml.v2 v2.2.2
)
