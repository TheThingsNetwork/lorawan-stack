---
title: "Install"
description: ""
weight: 1
draft: false
--- 

# <a name="docker">Docker</a>

Docker is available for most platforms, you can find most platform on its [documentation website](https://docs.docker.com/install/).
If your platform isn't listed, refer to your OS documentation. Note that Linux distributions are likely to provide docker through their package manager.

## <a name="docker-compose">Docker-compose</a>

Along docker you will need to have docker-compose. It will allow us to quickly run and configure our docker images.
Instruction can also be found on the docker [documentation website](https://docs.docker.com/compose/install/). Same apply here for Linux distributions.

# <a name="certificated">Certificates</a>

By default, the Stack requires a `cert.pem` and `key.pem`, in order to to serve content over TLS.

+ To generate self-signed certificates for `localhost`, use the following command. This requires a [Go environment setup](https://golang.org/doc/install).

```bash
go run $(go env GOROOT)/src/crypto/tls/generate_cert.go -ca -host localhost && chmod 0444 ./key.pem
```

Keep in mind that self-signed certificates are not trusted by browsers and operating systems, resulting oftentimes in warnings and sometimes in errors. Consider [Let's Encrypt](https://letsencrypt.org/getting-started/) for free and trusted TLS certificates for your server.
