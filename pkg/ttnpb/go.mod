module go.thethings.network/lorawan-stack/v3/pkg/ttnpb

go 1.14

replace go.thethings.network/lorawan-stack/v3 => ../../

// Use our fork of grpc-gateway.
replace github.com/grpc-ecosystem/grpc-gateway => github.com/TheThingsIndustries/grpc-gateway v1.14.4-gogo

// Use our fork of otto.
replace github.com/robertkrimen/otto => github.com/TheThingsIndustries/otto v0.0.0-20181129100957-6ddbbb60554a

// github.com/blang/semver doesn't have a v3 semantic import.
replace github.com/blang/semver => github.com/blang/semver v0.0.0-20190414182527-1a9109f8c4a1

// Dependency of Goreleaser that causes problems with module management.
// See https://github.com/Azure/go-autorest/issues/414.
replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.1+incompatible

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
	github.com/blang/semver v0.0.0-00010101000000-000000000000
	github.com/envoyproxy/protoc-gen-validate v0.3.0-java
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.3.5
	github.com/grpc-ecosystem/grpc-gateway v1.14.3
	github.com/kr/pretty v0.2.0
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/smartystreets/assertions v1.0.1
	go.thethings.network/lorawan-stack/v3 v3.8.0
	google.golang.org/genproto v0.0.0-20200401122417-09ab7b7031d2
	google.golang.org/grpc v1.28.1
)
