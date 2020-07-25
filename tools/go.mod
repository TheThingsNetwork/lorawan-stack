module go.thethings.network/lorawan-stack/tools

go 1.14

replace go.thethings.network/lorawan-stack/v3 => ../

// Dependency of lorawan-stack.
replace github.com/grpc-ecosystem/grpc-gateway => github.com/TheThingsIndustries/grpc-gateway v1.14.5-gogo

// Dependency of lorawan-stack.
replace gopkg.in/DATA-DOG/go-sqlmock.v1 => gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0

// Dependency of Goreleaser that causes problems with module management.
// See https://github.com/Azure/go-autorest/issues/414.
replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.1+incompatible

// TODO: Remove once https://github.com/magefile/mage/pull/307 is merged.
replace github.com/magefile/mage v1.9.1 => github.com/TheThingsIndustries/mage v1.9.1-0.20200520191129-8bccc5d0bd6f

// Dependency of lorawan-stack.
replace gocloud.dev => gocloud.dev v0.19.0

require (
	github.com/Azure/go-autorest/autorest v0.10.1 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.8.3 // indirect
	github.com/Azure/go-autorest/autorest/azure/auth v0.4.2 // indirect
	github.com/Azure/go-autorest/autorest/to v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.2.0 // indirect
	github.com/TheThingsIndustries/magepkg v0.0.0-20190214092847-6c0299b7c3ed
	github.com/blang/semver v3.5.1+incompatible
	github.com/client9/misspell v0.3.4
	github.com/cloudflare/cfssl v1.4.1
	github.com/gobuffalo/envy v1.9.0 // indirect
	github.com/gohugoio/hugo v0.71.0
	github.com/goreleaser/goreleaser v0.140.1
	github.com/magefile/mage v1.9.1
	github.com/mattn/goveralls v0.0.5
	github.com/mdempsky/unconvert v0.0.0-20200228143138-95ecdbfc0b5f
	github.com/mgechev/revive v1.0.2
	go.thethings.network/lorawan-stack/v3 v3.0.0-00010101000000-000000000000
	golang.org/x/tools v0.0.0-20200710042808-f1c4188a97a1
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/yaml.v2 v2.3.0
)
