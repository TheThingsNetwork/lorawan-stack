module go.thethings.network/lorawan-stack/tools

go 1.16

replace go.thethings.network/lorawan-stack/v3 => ../

// Dependency of lorawan-stack.
replace github.com/grpc-ecosystem/grpc-gateway => github.com/TheThingsIndustries/grpc-gateway v1.15.2-gogo

// Dependency of lorawan-stack.
replace github.com/gogo/protobuf => github.com/TheThingsIndustries/gogoprotobuf v1.3.1

// Dependency of lorawan-stack.
replace github.com/golang/protobuf => github.com/golang/protobuf v1.3.5

// Dependency of lorawan-stack.
replace google.golang.org/grpc => google.golang.org/grpc v1.33.1

// Dependency of lorawan-stack.
replace google.golang.org/genproto => google.golang.org/genproto v0.0.0-20200513103714-09dca8ec2884

// Dependency of lorawan-stack.
replace gopkg.in/DATA-DOG/go-sqlmock.v1 => gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0

// TODO: Remove once https://github.com/magefile/mage/pull/307 is merged.
replace github.com/magefile/mage => github.com/TheThingsIndustries/mage v1.10.0

// Dependency of lorawan-stack.
replace gocloud.dev => gocloud.dev v0.19.0

require (
	github.com/TheThingsIndustries/magepkg v0.0.0-20190214092847-6c0299b7c3ed
	github.com/blang/semver v3.5.1+incompatible
	github.com/client9/misspell v0.3.4
	github.com/cloudflare/cfssl v1.4.1
	github.com/magefile/mage v1.10.0
	github.com/mattn/goveralls v0.0.5
	github.com/mgechev/revive v1.0.2
	go.thethings.network/lorawan-stack/v3 v3.0.0-00010101000000-000000000000
	gopkg.in/yaml.v2 v2.4.0
)
