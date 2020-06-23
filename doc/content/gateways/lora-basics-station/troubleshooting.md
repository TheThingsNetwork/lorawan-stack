---
title: "Troubleshooting LBS"
description: ""
---

This section contains help for common problems you may encounter when using {{% lbs %}} to connect to {{% tts %}}.

<!--more-->

## What is my server address?

This is the address you use to access {{% tts %}}. If you followed the [Getting Started guide]({{< ref "/getting-started" >}}) this is the same as what you use instead of `thethings.example.com`.

CUPS uses the URI: `https://<server-address>:443`

LNS uses the URI: `wss://<server-address>:8887`

## How do I find the CA Trust?

Some device manufacturers include [common CAs](https://www.ccadb.org/) in device firmware, and these devices should connect automatically to {{% tts %}}.

If your device does not contain common CAs, and you are using Let's Encrypt to secure your domain, you may download the Let's Encrypt DST X3 Trust file [here](https://letsencrypt.org/certs/lets-encrypt-x3-cross-signed.pem.txt). Save the contents of the file as `cert.pem` and upload it as the Server Certificate on your gateway when connecting to {{% lbs %}}.

If you are using self signed certificates, you are your own Trust. You may generate a Root Certificate from your private key using openssl:

```bash
$ openssl req -x509 -new -nodes -key <rootCA.key> -sha256 -days 1024  -out <rootCA.pem>
```

## How do I use an API Key?

This varies from device to device. If your device allows you to upload a `.key` file, copy your gateway API Key in to a `gateway-api.key` file (the filename is not important) as an HTTP header in the following format:

```
Authorization: <gateway-api-key>
```

See the [{{% lbs %}} Authorization documentation](https://doc.sm.tc/station/authmodes.html) or your manufacturers guidelines for additional information.

## Is an API Key required?

CUPS requires an API Key with the following rights:
- View gateway information
- Edit basic gateway settings

LNS does not require an API Key. If you wish to use Token Authentication, create an API Key with the following rights:
- Link as Gateway to a Gateway Server for traffic exchange, i.e. write uplink and read downlink

