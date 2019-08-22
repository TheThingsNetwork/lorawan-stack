---
title: "Certificates"
description: ""
weight: 2
---

## Certificates

By default, The Things Stack requires a `cert.pem` and `key.pem`, in order to to serve content over TLS.

Typically you'll get these from a trusted Certificate Authority. Use the "full chain" for `cert.pem` and the "private key" for `key.pem`. The Things Stack also has support for automated certificate management (ACME). This allows you to easily get trusted TLS certificates for your server from [Let's Encrypt](https://letsencrypt.org/getting-started/). If you want this, you'll need to create an `acme` directory that The Things Stack can write in:

```bash
$ mkdir ./acme
$ chown 886:886 ./acme
```

> If you don't do this, you'll get an error saying something like `open /var/lib/acme/acme_account+key<...>: permission denied`.

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
