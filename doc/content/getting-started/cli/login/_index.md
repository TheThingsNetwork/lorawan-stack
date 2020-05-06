---
title: "Login with the Command-line interface"
description: ""
---

## Login

The CLI needs to be logged on in order to create gateways, applications, devices and API keys. With {{% tts %}} running, login with the following command:

```bash
$ ttn-lw-cli login
```

This will open a browser window with the OAuth login page where you can login with your credentials. This is also where you can create a new account if you do not already have one.

> During the login procedure, the CLI starts a webserver on `localhost` in order to receive the OAuth callback after login. If you are running the CLI on a machine that is not `localhost`, you can pass the `--callback=false` flag. This will allow you to perform part of the OAuth flow on a different machine, and copy-paste a code back into the CLI.