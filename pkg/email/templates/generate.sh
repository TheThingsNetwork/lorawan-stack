#!/usr/bin/env bash

for template in $(find . -name '*.mjml' -not -name '_*.mjml')
do
  yarn run mjml $PWD/$template -o $PWD/${template%.mjml}.html.tmpl
done
