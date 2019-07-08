---
title: "Documentation guidelines"
description: "Extensive guide on how to write documentation for The Things Network LoRaWAN stack"
weight: 3
category: guidelines
---

The documentation is build using [HUGO](https://gohugo.io/).
All the content is written in Markdown.

## HUGO

### Install

Since HUGO is written essentially in `go` you don't have to install anything, our build
system will task itself to `run` the right commands

### Commands

There are two important command when working on the documentation
* `mage docs:deps` pull the dependencies necessary to generate the doc.
* `mage docs:server` to start HUGO live server and visualize the doc.

### Front Matter

The Front Matter refer to the top part of each content page.
It contains metadata about the pages such as `title` and `tags`.

We impose the following minimum for as a Front Matter;
```yaml
title: "Documentation guidelines"
description: "Extensive guide on how to write documentation for The Things Network LoRaWAN stack"
weight: 3
```

`description` can be empty,in this case just put an empty string "".
You are welcome to add any `category` or `tags` you find relevant.


## Renaming

Hugo support aliasing and permalinks. If you find yourself in a situation where
you need to change the URL of a page makes sure to add aliases to the 
Front Matter.
```yaml
aliases:
    - /posts/my-original-url/
    - /2010/01/01/even-earlier-url.html
```
[Source](https://gohugo.io/content-management/urls/#example-aliases)

## Configuration

For developement you can make change to the `config.toml`.
If you want to make permanent changes you need to edit `config.tmpl` and run `mage docs:config`
then commit your changes.

## Resources

Leave images and others sources you are using in the same folder as your written file.
