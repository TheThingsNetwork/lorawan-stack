# The Things Stack Email Templates

This directory contains the email templates used by The Things Stack.

## Email Templates

An email template "source" consists of three files:

- `template_name.go` defines the data structure for rendering the template and registers the template. The template for the subject of the email is also defined here.
- `template_name.html.tmpl` defines the HTML template body.
- `template_name.txt.tmpl` defines the plaintext template body.

## Base Templates

Each of these email templates gets rendered together with the base template that's defined in `template_name.go`. For HTML templates, it's currently always `base.html.tmpl` and for plaintext templates we currently don't use a base template.

We use [MJML](https://mjml.io/) to generate the HTML base template `base.html.tmpl` from `base.mjml`. Do not edit `base.html.tmpl` directly. If you want to edit the HTML base template, edit the `.mjml` files and run `go generate ./pkg/email/templates` to update `base.html.tmpl`.

The HTML base template contains three blocks: **title**, **preview** and **body**. The **title** is typically not visible in emails, but would be visible if we start supporting "view this email online" functionality. The **preview** is a short text that many email clients render in the list view, next to the subject.

## Testing

We unit-test our email templates by diff-ing generated emails with the committed "golden" files in the `testdata` folder. Running tests is done as usual, with `go test ./pkg/email/templates`. To update the "golden" files, run the tests with the `-write-golden` flag: `go test ./pkg/email/templates -write-golden`.

## Template Conventions

We start emails to users with `Dear {{ .ReceiverName }},`. Emails that are not sent to users (such as invitations) start with `Hello,`

| **Content**       | **HTML**                                    | **plain text** and **preview** block of **HTML** |
| ----------------- | ------------------------------------------- | ------------------------------------------------ |
| Network Name      | `<b>{{ .Network.Name }}</b>`                | `{{ .Network.Name }}`                            |
| IDs, tokens, etc. | `<code>{{ .Receiver.Ids.IDString }}</code>` | `"{{ .Receiver.Ids.IDString }}"`                 |

## Template Variables

See the `{TemplateName}Data` struct, which typically embeds `email.TemplateData` or `email.NotificationTemplateData`.

### Writing New Email Templates

- Write `template_name.go`, `template_name.html.tmpl` and `template_name.txt.tmpl`
- Add to `templates_test.go`
- Update the "golden" files
