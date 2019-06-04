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

Hugo is available on most platform via packet managers. See HUGO install instruction
for more details.

https://github.com/gohugoio/hugo#choose-how-to-install

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
