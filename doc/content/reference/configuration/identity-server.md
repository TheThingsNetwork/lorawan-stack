---
title: 'Identity Server Options'
description: ''
weight: 2
---

## Database Options

The Identity Server needs to be connected to a PostgreSQL-compatible database. Details for the form of the URI can be found in the [PostgreSQL documentation](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING).

- `is.database-uri`: Database connection URI

## Email Options

The Identity Server can be configured with different providers for sending emails. Currently the `sendgrid` and `smtp` providers are implemented.

- `is.email.provider`: Email provider to use

When `sendgrid` is used as provider, an API key is required. For testing, use the sandbox to prevent emails from actually being sent.

- `is.email.sendgrid.api-key`: The SendGrid API key to use
- `is.email.sendgrid.sandbox`: Use SendGrid sandbox mode for testing

When `smtp` is used as provider, provide the address of the SMTP server (`host:port`), as well as the username and password for the SMTP server.

- `is.email.smtp.address`: SMTP server address
- `is.email.smtp.username`: Username to authenticate with
- `is.email.smtp.password`: Password to authenticate with
- `is.email.smtp.connections`: Maximum number of connections to the SMTP server

The email address and name of the sender should be configured regardless of the provider that is used.

- `is.email.sender-address`: The address of the sender
- `is.email.sender-name`: The name of the sender

Most emails contain the name of the network and links to the Identity Server or Console.

- `is.email.network.name`: The name of the network
- `is.email.network.identity-server-url`: The URL of the Identity Server
- `is.email.network.console-url`: The URL of the Console

Although The Things Stack comes with a number of builtin email templates, it is possible to override those with custom templates. You can specify the source where to load templates from, and options for that source. For more information on email templates, see the [email templates reference]({{< relref "../email-templates" >}}).

- `is.email.templates.source`: Source of the email template files (directory, url, blob)
- `is.email.templates.directory`: Directory on the filesystem where email templates are located
- `is.email.templates.url`: URL where email templates are located
- `is.email.templates.blob.bucket`: Bucket where email templates are located
- `is.email.templates.blob.path`: Path within the bucket.

If your custom templates rely on other files, such as headers or footers, those files need to be included.

- `is.email.templates.includes`: The email templates that will be preloaded on startup

## OAuth UI Options

The OAuth user interface needs to be configured with at least the canonical URL and the base URL of the Identity Server's HTTP API. The canonical URL needs to be the full URL of the UI, and looks like `https://thethings.example.com/oauth`. The base URL of the Identity Server's HTTP API looks like `https://thethings.example.com/api/v3`.

- `is.oauth.ui.canonical-url`: The page canonical URL
- `is.oauth.ui.is.base-url`: Base URL to the HTTP API

If you do not want to serve the OAuth user interface on `/oauth`, you may customize the mount path.

- `is.oauth.mount`: Path on the server where the OAuth server will be served

If page assets for the OAuth UI are served from a CDN or on a different path on the server, the base URL needs to be customized as well. If you want to [customize the branding]({{< relref "../../branding" >}}) of the OAuth UI, you can set the base URL for where your branding assets are located.

- `is.oauth.ui.assets-base-url`: The base URL to the page assets
- `is.oauth.ui.branding-base-url`: The base URL to the branding assets

The appearance of The Things Stack can optionally be customized.

- `is.oauth.ui.site-name`: The site name
- `is.oauth.ui.title`: The page title
- `is.oauth.ui.sub-title`: The page sub-title
- `is.oauth.ui.descriptions`: The page description
- `is.oauth.ui.language`: The page language
- `is.oauth.ui.theme-color`: The page theme color

Further customization of the CSS files, JS files and icons is also possible:

- `is.oauth.ui.css-file`: The names of the CSS files
- `is.oauth.ui.js-file`: The names of the JS files
- `is.oauth.ui.icon-prefix`: The prefix to put before the page icons (favicon.ico, touch-icon.png, og-image.png)

## Profile Picture Storage Options

The profile pictures that users upload for their accounts are stored in a blob bucket. The global [blob configuration]({{< relref "the-things-stack.md#blob-options" >}}) is used for this. In addition to those options, specify the name of the bucket and the public URL to the bucket.

- `is.profile-picture.bucket`: Bucket used for storing profile pictures
- `is.profile-picture.bucket-url`: Base URL for public bucket access

It is also possible to use [Gravatar](https://gravatar.com) for profile pictures.

- `is.profile-picture.use-gravatar`: Use Gravatar fallback for users without profile picture

## End Device Picture Storage Options

Similar to profile pictures, end devices can have pictures associated with them.

- `is.end-device-picture.bucket`: Bucket used for storing end device pictures
- `is.end-device-picture.bucket-url`: Base URL for public bucket access

## User Registration Options

The user registration process can be customized by requiring approval by admin users, requiring email validation or by requiring new users to be invited by existing users.

- `is.user-registration.admin-approval.required`: Require admin approval for new users
- `is.user-registration.contact-info-validation.required`: Require contact info validation for new users
- `is.user-registration.invitation.required`: Require invitations for new users
- `is.user-registration.invitation.token-ttl`: TTL of user invitation tokens

There are several options to customize the requirements for user passwords.

- `is.user-registration.password-requirements.max-length`: Maximum password length
- `is.user-registration.password-requirements.min-digits`: Minimum number of digits
- `is.user-registration.password-requirements.min-length`: Minimum password length
- `is.user-registration.password-requirements.min-special`: Minimum number of special characters
- `is.user-registration.password-requirements.min-uppercase`: Minimum number of uppercase letters
