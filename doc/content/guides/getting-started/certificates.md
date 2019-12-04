---
title: "Certificates"
description: ""
weight: 2
---

## Certificates

{{% tts %}} will be configured with Transport Layer Security (TLS) and HTTPS. This requires a TLS certificate and a corresponding key. In this guide we'll request a free, trusted certificate from [Let's Encrypt](https://letsencrypt.org/getting-started/), but if you already have a certificate (`cert.pem`) and key (`key.pem`), you can also use those.

### Automatic Certificate Management (ACME)

For automatic certificates, we're going to need an `acme` directory where {{% tts %}} can store the certificate data:

```bash
$ mkdir ./acme
$ chown 886:886 ./acme
```

> `886` is the uid and the gid of the user that runs {{% tts %}} in the Docker container. If you don't set these permissions, you'll get an error saying something like `open /var/lib/acme/acme_account+key<...>: permission denied`.

### Custom Certificates

If you want to use the certificate (`cert.pem`) and key (`key.pem`) that you already have, you also need to set these permissions.

```bash
$ chown 886:886 ./cert.pem ./key.pem
```

> If you don't set these permissions, you'll get an error saying something like `/run/secrets/key.pem: permission denied`.

### Self-Signed Development Certificates

It is possible to make {{% tts %}} use self-signed development certificates with similar configuration as you would have for custom certificates. Creating and trusting self-signed certificates is not covered by this guide.
