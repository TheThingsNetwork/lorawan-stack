---
title: "Console Options"
description: ""
---

## Console Mount

The Console can be served under any arbitrary path on your server

- `console.mount`: Path on the server where the Console will be served

## OAuth Options

The Console app uses the [OAuth 2.0 authorization flow](https://en.wikipedia.org/wiki/OAuth#OAuth_2.0_2) to authorize actions in the backend. You can customize the authorization parameters if necessary:

- `console.oauth.authorize-url`: The OAuth Authorize URL
- `console.oauth.client-id`: The OAuth client ID
- `console.oauth.client-secret`: The OAuth client secret
- `console.oauth.token-url`: The OAuth Token Exchange URL
- `console.oauth.logout-url`: The logout URL of the OAuth server used to perform client initiated logouts

## Frontend Setup

You can change various values that will be passed to the JavaScript logic and HTML Head tags of the Web UI:

### Component Configuration

The Console needs to know how the individual {{% tts %}} components are configured, so it can execute API calls correctly. Likewise, disabled components will cause possible respective options and menus in the Console frontend to be disabled.

- `console.ui.as.base-url`: Base URL to the HTTP API
- `console.ui.as.enabled`: Enable Application Server related functionality in the Console
- `console.ui.gs.base-url`: Base URL to the HTTP API
- `console.ui.gs.enabled`: Enable Gateway Server related functionality in the Console
- `console.ui.edtc.base-url`: Base URL to the HTTP API
- `console.ui.edtc.enabled`: Enable End Device Template Converter related functionality in the Console
- `console.ui.is.base-url`: Base URL to the HTTP API
- `console.ui.is.enabled`: Enable Identity Server related functionality in the Console
- `console.ui.js.base-url`: Base URL to the HTTP API
- `console.ui.js.enabled`: Enable Join Server related functionality in the Console
- `console.ui.ns.base-url`: Base URL to the HTTP API
- `console.ui.ns.enabled`: Enable Network Server related functionality in the Console
- `console.ui.qrg.base-url`: Base URL to the HTTP API
- `console.ui.qrg.enabled`: Enable QR Code Generator related functionality in the Console

### File Includes

You can control which CSS and JavaScript files are included into the rendered HTML of the Console. Note that these files must be present in the `/public` folder during runtime. 

- `console.ui.js-file`: The names of the JavaScript file
- `console.ui.css-file`: The names of the CSS files

### Assets and Custom Branding

You can control the url of the assets folder that the Console frontend will use, e.g. if you plan on using a CDN to serve your assets. The branding base url will enable custom branding for logos, favicons and OpenGraph images. See the [custom branding guide]({{< ref "/reference/branding" >}}) for detailed descriptions on this subject.

- `console.ui.assets-base-url`: The base URL to the page assets
- `console.ui.branding-base-url`: The base URL to the branding assets
- `console.ui.icon-prefix`: The prefix to put before the page icons (favicon.ico, touch-icon.png, og-image.png)
- `console.ui.canonical-url`: The page canonical URL
- `console.ui.descriptions`: The page description
- `console.ui.language`: The page language
- `console.ui.site-name`: The site name
- `console.ui.sub-title`: The page sub-title
- `console.ui.support-link`: The URI that the support button will point to
- `console.ui.documentation-base-url`: The base URL for generating documentation links
- `console.ui.theme-color`: The page theme color
- `console.ui.title`: The page title
