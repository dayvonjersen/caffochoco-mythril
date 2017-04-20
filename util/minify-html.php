<?php
/**
 * usage: php -f minify-html.php -- [INPUT FILE]
 */
set_error_handler(function($errno, $errstr) {
    // php is love
    // php is life
    $type = 'PHP ERROR';
    switch($errno) {
    case E_ERROR:             $type = 'E_ERROR'; break;
    case E_WARNING:           $type = 'E_WARNING'; break;
    case E_PARSE:             $type = 'E_PARSE'; break;
    case E_NOTICE:            $type = 'E_NOTICE'; break;
    case E_CORE_ERROR:        $type = 'E_CORE_ERROR'; break;
    case E_CORE_WARNING:      $type = 'E_CORE_WARNING'; break;
    case E_COMPILE_ERROR:     $type = 'E_COMPILE_ERROR'; break;
    case E_COMPILE_WARNING:   $type = 'E_COMPILE_WARNING'; break;
    case E_USER_ERROR:        $type = 'E_USER_ERROR'; break;
    case E_USER_WARNING:      $type = 'E_USER_WARNING'; break;
    case E_USER_NOTICE:       $type = 'E_USER_NOTICE'; break;
    case E_STRICT:            $type = 'E_STRICT'; break;
    case E_RECOVERABLE_ERROR: $type = 'E_RECOVERABLE_ERROR'; break;
    case E_DEPRECATED:        $type = 'E_DEPRECATED'; break;
    case E_USER_DEPRECATED:   $type = 'E_USER_DEPRECATED'; break;
    }
    fwrite(STDERR, "\033[31m$type\033[0m: $errstr\n");
});
fwrite(STDERR, "minifying {$argv[1]}\n");
fwrite(STDERR, "\n\033[32m".
     'this is going to spew a bunch of warnings'."\033[0m".' because'."\n".
     'polymer lets you assign the same id attribute to '."\n".
     'multiple elements as long as they\'re contained in'. "\n".
     'separate shadow trees.'."\n");

// do `composer install` first

require_once 'vendor/autoload.php';
use Masterminds\HTML5;

$html5 = new HTML5();
$document = $html5->loadHTML(file_get_contents($argv[1]));

// process <script> tags
//
// dump tag contents into separate files which will be concatenated
// and minified externally (see make-dist.sh)
// 
// NOTE: appending an extra semicolon is necessary for concatenation,
// any amount of semicolons e.g. ;; is valid javascript.
//
// all these nodes are then removed from the document

$nodes = [];
$nodeList = $document->getElementsByTagName('script');
$i = 0;
foreach($nodeList as $node) {
    $tmpfile = sprintf('tmp/script_%03d.js', $i);
    file_put_contents($tmpfile, $node->textContent . ';'); 
    $i++;
    $nodes[] = $node;
}
foreach($nodes as $node) {
    $node->parentNode->removeChild($node);
}
$scriptElement = $document->createElement('script');
$scriptElement->setAttribute('src', '/app.min.js');
$bodyElement = $document->getElementsByTagName('body')->item(0);
$bodyElement->appendChild($scriptElement);

// process <style> tags
//
// polymer 1.0 uses <style is="custom-style"> which tends to have
// invalid CSS like @apply and --variables: { --with-nested-properties: ...
// so I'm just collapsing the whitespace
//
// the call to exec() is majorly slow because when you call node.js tools
// like this it fires up a new instance of node every time
//
// exec() is reading each style tag into a file first because 
// PHP was cutting it off if I passed it as a string
// 
// if the minification fails, the style tag's whitespace is simply collapsed

function collapseWhitespace($css) {
    // removes /* */ comments
    $css = preg_replace('/\/\*([^*]|[\r\n\s]|(\*+?([^*\/]|[\r\n\s])))*\*+?\//ms', '', $css);
    // removes excess whitespace around and between { : ; ,
    $css = preg_replace('/\s*(.*?)\s*([\{:;,])\s*/ms','\1\2',$css);
    // removes leading and trailing whitespace
    $css = trim($css);
    return $css;
}

$nodeList = $document->getElementsByTagName('style');
$i = 0;
foreach($nodeList as $node) {
    if(!$node->hasAttribute('is')) {
        $tmpfile = sprintf('tmp/style_%03d.css', $i);
        file_put_contents($tmpfile, $node->textContent);
        $i++;

        exec('bash -c "csso '.$tmpfile.' --restructure-off" 2>&1', $out, $exit_code);
        $output = trim(implode('', str_replace("\r", '', $out)));
        if($exit_code == 0) {
            $node->nodeValue = $output;
        } else {
            trigger_error($output);
            $node->nodeValue = collapseWhitespace($node->textContent);
        }
        unset($out, $exit_code);
    } else {
        $node->nodeValue = collapseWhitespace($node->textContent);
    }
}

// remove all HTML comments and whitespace textNodes 
// (which are likely indentation and newlines)
$nodes = [];
$xpath = new DOMXPath($document);
$commentNodes = $xpath->query('//comment()');
foreach($commentNodes as $commentNode) {
    $nodes[] = $commentNode;
}
$textNodes = $xpath->query('//text()');
foreach($textNodes as $textNode) {
    if(!trim($textNode->textContent)) {
        $nodes[] = $textNode;
    }
}
foreach($nodes as $node) {
    $node->parentNode->removeChild($node);
}

$output = $html5->saveHTML($document);
echo $output;
