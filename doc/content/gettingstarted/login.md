---
title: "Login"
description: ""
weight: 4
---

## <a name="login">Login using the CLI</a>

The CLI needs to be logged on in order to create gateways, applications, devices and API keys. With the stack running in one terminal session, login with the following command:

```bash
$ ttn-lw-cli login
```

This will open the OAuth login page where you can login with your credentials (here those will be the `admin` user and the password you enter when you started the stack). Once you logged in the browser, return to the terminal session to proceed.

> If you run this command on a remote machine, pass `--callback=false` to get a link to login on your local machine.
