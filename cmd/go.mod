module go.thethings.network/lorawan-stack/v3/cmd

go 1.14

replace go.thethings.network/lorawan-stack/v3 => ../

// Use our fork of grpc-gateway.
replace github.com/grpc-ecosystem/grpc-gateway => github.com/TheThingsIndustries/grpc-gateway v1.14.4-gogo

// Use our fork of otto.
replace github.com/robertkrimen/otto => github.com/TheThingsIndustries/otto v0.0.0-20181129100957-6ddbbb60554a

// github.com/blang/semver doesn't have a v3 semantic import.
replace github.com/blang/semver => github.com/blang/semver v0.0.0-20190414182527-1a9109f8c4a1

// Dependency of Goreleaser that causes problems with module management.
// See https://github.com/Azure/go-autorest/issues/414.
replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v14.1.0+incompatible

// Do not upgrade Echo beyond v4.1.2.
// See https://github.com/TheThingsNetwork/lorawan-stack/issues/977.
replace github.com/labstack/echo/v4 => github.com/labstack/echo/v4 v4.1.2

// Do not upgrade go-sqlmock beyond v1.3.0.
// See https://github.com/heptiolabs/healthcheck/issues/23.
replace gopkg.in/DATA-DOG/go-sqlmock.v1 => gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0

// Dependency of Hugo that causes problems with module management.
replace github.com/russross/blackfriday => github.com/russross/blackfriday v1.5.2

// Dependency of Hugo that causes problems with module management.
replace github.com/nicksnyder/go-i18n => github.com/nicksnyder/go-i18n v1.10.0

require (
	github.com/disintegration/imaging v1.6.2
	github.com/getsentry/sentry-go v0.5.1
	github.com/gogo/protobuf v1.3.1
	github.com/howeyc/gopass v0.0.0-20190910152052-7cb4b85ec19c
	github.com/jinzhu/gorm v1.9.12
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	go.thethings.network/lorawan-stack/v3 v3.0.0-00010101000000-000000000000
	gocloud.dev v0.19.0
	golang.org/x/crypto v0.0.0-20200429183012-4b2356b1ed79
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	google.golang.org/grpc v1.29.1
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/yaml.v2 v2.2.8
)
