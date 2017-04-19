<?php
echo 'data:image/svg+xml;base64,', base64_encode(file_get_contents($argv[1])), "\n";
