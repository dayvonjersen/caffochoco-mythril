#!/bin/bash

sed -i "s/'import' href='\//'import' href='/g" *.html
polymer-bundler --strip-comments --inline-scripts --inline-css index.html > tmp/index.min.html
for f in `ls --color=never *.html`; do
    git checkout $f
done
php -f util/minify-html.php -- tmp/index.min.html
cat tmp/*.js | ./node_modules/uglify-js/bin/uglifyjs > dist/app.min.js
mv tmp/index.min.html dist/index.html
