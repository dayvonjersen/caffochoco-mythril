#!/bin/bash
set -e
# polymer-bundler, formerly vulcanize, will pull in all 
# the imports, external scripts, and stylesheets for the
# entire app and dump it into one giant HTML, but will 
# not minify all assets in an optimum way so this script 
# will try to correct that

# the -p option for polymer-bundler is broken so we 
# specify relative paths with sed
# and then revert those changes

sed -i "s/'import' href='\//'import' href='/g" *.html
polymer-bundler --strip-comments --inline-scripts --inline-css index.html > tmp/index.min.html
git checkout *.html


# minify-html.php
# - strips all comments and excess whitespace from HTML
# - pulls out all javascript from inlined <script> tags
#   so it can be concatenated and minified with uglifyjs
# - minifies inlined <style> tags in-place 
#   IFF the stylesheet does not contain polymer-specific css
#   (can't pull all of it out and minify together because of 
#    shadowdom css scoping and things like :host)
#
# NOTE:
# polymer as part of its databinding system has a $= operator 
# which breaks how minify-html.php parses the html since $ 
# becomes part of the attribute name
#
# sed is being used here to correct this problem by replacing
# $= with a uniquely identifyable token which is also valid html
# so we can let the php script do its business and then revert it
# unfortunately this approach leaks into the extracted javascript

sed -i "s/\$=/xxxxxxdollarsignequals=/g" tmp/index.min.html
php -f util/minify-html.php -- tmp/index.min.html > dist/index.html
sed -i "s/xxxxxxdollarsignequals=/$=/g" dist/index.html

cat tmp/*.js > tmp/app.min.js
sed -i "s/xxxxxxdollarsignequals=/$=/g" tmp/app.min.js
./node_modules/uglify-js/bin/uglifyjs tmp/app.min.js > dist/app.min.js
