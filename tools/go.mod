module go.thethings.network/lorawan-stack/v3/tools

go 1.14

replace go.thethings.network/lorawan-stack/v3 => ../

replace go.thethings.network/lorawan-stack/v3/cmd => ../cmd

// Use our fork of grpc-gateway.
replace github.com/grpc-ecosystem/grpc-gateway => github.com/TheThingsIndustries/grpc-gateway v1.14.4-gogo

// Dependency of Goreleaser that causes problems with module management.
// See https://github.com/Azure/go-autorest/issues/414.
replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v14.1.0+incompatible

// Do not upgrade go-sqlmock beyond v1.3.0.
// See https://github.com/heptiolabs/healthcheck/issues/23.
replace gopkg.in/DATA-DOG/go-sqlmock.v1 => gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0

// Dependency of Hugo that causes problems with module management.
replace github.com/russross/blackfriday => github.com/russross/blackfriday v1.5.2

// Dependency of Hugo that causes problems with module management.
replace github.com/nicksnyder/go-i18n => github.com/nicksnyder/go-i18n v1.10.0

require (
	cloud.google.com/go v0.57.0 // indirect
	cloud.google.com/go/storage v1.7.0 // indirect
	github.com/Azure/azure-sdk-for-go v42.1.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.10.1 // indirect
	github.com/Azure/go-autorest/autorest/azure/auth v0.4.2 // indirect
	github.com/Azure/go-autorest/autorest/to v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.2.0 // indirect
	github.com/aws/aws-sdk-go v1.30.23 // indirect
	github.com/bep/golibsass v0.7.0 // indirect
	github.com/client9/misspell v0.3.4
	github.com/gobuffalo/envy v1.9.0 // indirect
	github.com/gohugoio/hugo v0.70.0
	github.com/golang/protobuf v1.4.1 // indirect
	github.com/goreleaser/goreleaser v0.133.0
	github.com/hashicorp/go-retryablehttp v0.6.6 // indirect
	github.com/kyokomi/emoji v2.2.2+incompatible // indirect
	github.com/magefile/mage v1.9.0
	github.com/markbates/inflect v1.0.4 // indirect
	github.com/mattn/goveralls v0.0.5
	github.com/mdempsky/unconvert v0.0.0-20200228143138-95ecdbfc0b5f
	github.com/mgechev/revive v1.0.2
	github.com/nicksnyder/go-i18n v1.10.1 // indirect
	github.com/rogpeppe/go-internal v1.6.0 // indirect
	github.com/tdewolff/minify/v2 v2.7.4 // indirect
	go.thethings.network/lorawan-stack/v3 v3.0.0-00010101000000-000000000000
	go.thethings.network/lorawan-stack/v3/cmd v0.0.0-00010101000000-000000000000
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	golang.org/x/tools v0.0.0-20200507192325-6441d34c3f03
)
