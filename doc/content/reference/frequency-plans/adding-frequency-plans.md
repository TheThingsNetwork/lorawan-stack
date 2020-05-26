---
title: "Adding Frequency Plans"
description: ""
---

{{% tts %}} uses frequency plans from the [lorawan-frequency-plans Github repository.](https://github.com/TheThingsNetwork/lorawan-frequency-plans/)

To add frequency plans, you may open a pull request on the [lorawan-frequency-plans Github repository](https://github.com/TheThingsNetwork/lorawan-frequency-plans/), which makes the plan available in most deployments.

You may also fork the repository and reference your fork, with potentially custom plans, or with standard plans left out.

You may also configure a local directory (or cloud storage bucket) where the frequency plans are stored, in case the private network cannot access GitHub or does not want to rely on GitHub.

See [configuration options]({{< ref "/reference/configuration/the-things-stack#frequency-plans-options" >}}) to configure a different repository or directory for frequency plans.
