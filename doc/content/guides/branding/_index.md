---
title: 'Web UI Branding'
description: ''
weight: 50
---

This reference gives details on how to customize the branding of the login pages and the Console.

<!--more-->

## Title, Subtitle and Description

The title, subtitle and description of the login pages and the console can be changed using configuration options. See for details the [Identity Server configuration reference]({{< ref "/reference/configuration/identity-server#oauth-ui-options" >}}) and the [Console configuration reference]({{< ref "/reference/configuration/console" >}}).

## Logos

It is possible to change the logos of the web UI by changing the "branding base URL" to a location that contains the following files:

| **Filename**           | **Size** | **Purpose**                                                                   |
| ---------------------- | -------- | ----------------------------------------------------------------------------- |
| console-favicon.ico    | multiple | The logo for the console that is shown in browser tabs and bookmarks.         |
| console-og-image.png   | 1200x600 | The logo for the console that is shown when sharing links on social media     |
| console-touch-icon.png | 400x400  | The logo for the console that is shown mobile devices                         |
| logo.svg               | vector   | The logo for the console that is shown in the menu bar of the console         |
| oauth-favicon.ico      | multiple | The logo for the login pages that is shown in browser tabs and bookmarks      |
| oauth-og-image.png     | 1200x600 | The logo for the login pages that is shown when sharing links on social media |
| oauth-touch-icon.png   | 400x400  | The logo for the login pages that is shown mobile devices                     |

If the "branding base URL" option is set, "logo.svg" is used to display a secondary logo next to the logo of The Things Stack for LoRaWAN. It is recommended to use a logo with a wide (e.g. 5:1) or square (1:1) aspect ratio. Tall logos (e.g. 1:5; height larger than width) will be displayed very small, due to the limited height of the header bar.

For the exact configuration options that are required to set a custom "branding base URL", see the [Identity Server configuration reference]({{< ref "/reference/configuration/identity-server#oauth-ui-options" >}}) and the [Console configuration reference]({{< ref "/reference/configuration/console" >}}).

#### Hint

If you have your favicon as a PNG, use ImageMagick to convert it to ICO:
\$ convert console-favicon.png -define icon:auto-resize=64,48,32,16 console-favicon.ico
