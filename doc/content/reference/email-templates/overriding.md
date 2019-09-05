---
title: "Overriding Templates"
description: ""
weight: 6
---

> Note: We recommend getting familiar with the [html/template](https://golang.org/pkg/html/template/) template format first.

## Template components

An email template override has three components:

1. The subject file, which contains the subject of the email and is named `<identifier>.subject.txt`. 
2. The text contents file, which contains the contents of the email in text format and is named `<identifier>.txt`.
3. The HTML contents file, which contains the contents of the email in HTML format and is named `<identifier>.html`.


## Creating the overrides

In order to override a template, one must provide all three files as part of the email templates repository and then provide them as part of the configuration.

Let's consider that we want to override the email that a user receives once they register, the email validation email, which has the identifier `validate`. We need to create the following files:


- `validate.subject.txt`
```text
Please confirm your email address for {{.Network.Name}}
```
- `validate.txt`
```text
Please confirm your email address for {{.Network.Name}}.
Your email address will be used as contact for {{.Entity.Type}} "{{.Entity.ID}}". 

Reference: {{.ID}}
Confirmation Token: {{.Token}}
```
- `validate.html`
```html
Please confirm your email address for {{.Network.Name}}. <br> 
Your email address will be used as contact for {{.Entity.Type}} "{{.Entity.ID}}". <br> <br> 

Reference: {{.ID}} <br> 
Confirmation Token: {{.Token}}
```

## Providing the overrides to the stack

Once you have written your overrides, you can provide them to the stack either through an remote URL, or through the local file system.

### Fetching from a remote URL

In order to allow The Things Stack to access remote files, you must only provide the URL to the root folder that contains the files. Consider that you need to provide the The Things Stack access to a file called `validate.txt`, which you have uploaded on your host at `http://www.example.org/emails/validate.txt`. Then the URL that you provide to the email templates configuration is `http://www.example.org/emails/`.

### Fetching from a local directory

In order to allow the The Things Stack to access files which are hosted in your own file system, you must provide the path to the root folder that contains the files. Consider that you need to provide the The Things Stack access to a file called `validate.txt`, which is available on your file system in the `/srv/emails/validate.txt`. Then the path that you need to provide to the email templates configuration is `/srv/emails/`.

### Example YAML configuration 

```yaml
is:
  email:
    templates:
      # Specify only one source for the files.
      directory: "/path/to/the/template/files"
      url: "http://www.example.com/emails"
```
