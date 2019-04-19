---
title: "Install"
description: ""
weight: 1
draft: false
--- 
 
## macOS
 
 ```bash
 $ brew install TheThingsNetwork/lorawan-stack/ttn-lw-stack
 ```
 
## Linux
 
 ```bash
 $ sudo snap install ttn-lw-stack
 $ sudo snap alias ttn-lw-stack.ttn-lw-cli ttn-lw-cli
 ```
 
## Binaries
 
 If your operating system or package manager is not mentioned, please [download binaries](https://github.com/TheThingsNetwork/lorawan-stack/releases) for your operating system and processor architecture.

> In the unforeseen event none of our release are compatible with your system you can build the stack from source, see [DEVELOPMENT.md]({{< reffile "DEVELOPMENT.md" >}}) on the git repository.

## Certificates
 
 By default, the stack requires a `cert.pem` and `key.pem`, in order to to serve content over TLS.
 
 Typically you'll get these from a trusted Certificate Authority. We recommend [Let's Encrypt](https://letsencrypt.org/getting-started/) for free and trusted TLS certificates for your server. Use the "full chain" for `cert.pem` and the "private key" for `key.pem`.
 
 For local (development) deployments, you can generate self-signed certificates. If you have your [Go environment](../DEVELOPMENT.md#development-environment) set up, you can run the following command to generate a key and certificate for `localhost`:
 
 ```bash
 $ go run $(go env GOROOT)/src/crypto/tls/generate_cert.go -ca -host localhost
 ```
 
 In order for the user in our Docker container to read these files, you have to change the ownership of the certificate and key:
 
 ```bash
 $ chown 886:886 ./cert.pem ./key.pem
 ```
 
 > If you don't do this, you'll get an error saying something like `/run/secrets/key.pem: permission denied`.
 
 Keep in mind that self-signed certificates are not trusted by browsers and operating systems, resulting in warnings and errors such as `certificate signed by unknown authority` or `ERR_CERT_AUTHORITY_INVALID`. In most browsers you can add an exception for your self-signed certificate. You can configure the CLI to trust the certificate as well.
