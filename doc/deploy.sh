#!/usr/bin/env bash

tag=`git describe --abbrev=0 --tags`
cd public
doc_url=https://github.com/TheThingsNetwork/lorawan-stack/
echo "<!DOCTYPE html>
<html lang=\"en\">
<head>
  <title>Your Page Title</title>
  <meta http-equiv=\"refresh\" content=\"0; url=`echo $doc_url$tag`\" />
</head>
<body>
<p>Redirecting you to the latest documentation version.</p>

<p>If you are not automatically redirected click ont this link `echo $doc_url$tag` .</p>
</body>
</html>" > index.html

git add --all
git commit -m "Publishing changes of `echo $tag` to gh-pages"
git push origin gh-pages
