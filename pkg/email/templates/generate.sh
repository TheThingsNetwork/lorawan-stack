#!/usr/bin/env bash

# MJML v5+ disables mj-include by default; our base template composes trusted,
# repo-local partials (_attributes/_header/_footer), so includes are enabled here.
for template in $(find . -name '*.mjml' -not -name '_*.mjml')
do
  yarn run mjml --config.allowIncludes true $PWD/$template -o $PWD/${template%.mjml}.html.tmpl
  # mjml v5 omits the trailing newline; append one to keep a POSIX-clean file.
  printf '\n' >> $PWD/${template%.mjml}.html.tmpl
done
