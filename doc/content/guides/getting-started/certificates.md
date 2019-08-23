---
title: "Certificates"
description: ""
weight: 2
---

## Certificates

The Things Stack will be configured with Transport Layer Security (TLS) and HTTPS. This requires a TLS certificate and a corresponding key. In this guide we'll request a free, trusted certificate from [Let's Encrypt](https://letsencrypt.org/getting-started/), but if you already have a certificate (`cert.pem`) and key (`key.pem`), you can also use those.

For automatic certificates, we're going to need an `acme` directory where The Things Stack can store the certificate data:

```bash
$ mkdir ./acme
$ chown 886:886 ./acme
```

> `886` is the uid and the gid of the user that runs The Things Stack in the Docker container. If you don't set these permissions, you'll get an error saying something like `open /var/lib/acme/acme_account+key<...>: permission denied`.

If you want to use the certificate (`cert.pem`) and key (`key.pem`) that you already have, you also need to set these permissions.

```bash
$ chown 886:886 ./cert.pem ./key.pem
```

> If you don't set these permissions, you'll get an error saying something like `/run/secrets/key.pem: permission denied`.

For development deployments on `localhost` you can follow the steps in the `DEVELOPMENT.md` file in the Github repository of The Things Stack, but keep in mind that self-signed development certificates are not trusted by browsers and operating systems, resulting in warnings and errors such as `certificate signed by unknown authority` or `ERR_CERT_AUTHORITY_INVALID`. In most browsers you can add exceptions for your development certificate. You can configure the CLI to trust the certificate as well.
