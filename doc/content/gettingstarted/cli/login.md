---
title: "Logging in"
description: ""
weight: 1
---

The CLI needs to be logged on in order to create gateways, applications, devices and API keys. With the stack running in one terminal session, login with the following command:

```bash
$ ttn-lw-cli login
```

This will open the OAuth login page where you can login with your credentials. Once you logged in in the browser, return to the terminal session to proceed.

If you run this command on a remote machine, pass `--callback=false` to get a link to login on your local machine.
