<?php
/**
 * usage: php -f util/svg2datauri.php something.svg
 *
 * If you want to inline an svg as a dataURI use this
 */
echo 'data:image/svg+xml;base64,', base64_encode(file_get_contents($argv[1])), "\n";
