<?php
require_once 'vendor/autoload.php';
use Masterminds\HTML5;

function getImports($html_file, $basedir = '.') {
    static $paths = [];
    $html_file = $basedir .'/'.$html_file;
    if(!file_exists($html_file)) {
        echo 'ERROR: ', $html_file, ' not found',"\n";
        exit;
        return;
    }
    
    if(in_array($html_file, $paths)) {
        echo $html_file, ' already included.',"\n";
        return;
    }
    $paths[] = $html_file;

    $document = new DOMDocument();
    $document->formatOutput = false;
    $document->preserveWhitespace = false;
    $html5 = new HTML5([
        'encode_entities' => false,
        'disable_html_ns' => true,
        'target_document' => $document,
    ]);

    $html5->loadHTML(file_get_contents($html_file));
    $document->normalizeDocument();

    $nodeList = $document->getElementsByTagName('link');
    $i = 0;
    foreach($nodeList as $node) {
        if($node->hasAttribute('rel') && $node->getAttribute('rel') == "import") {
            $importHref = ltrim($node->getAttribute('href'), '/');

            $file = pathinfo($importHref, PATHINFO_BASENAME);
            $dir = realpath($basedir .'/'.pathinfo($importHref, PATHINFO_DIRNAME));
            echo 'import ', $dir, '/', $file, "\n";
            file_put_contents('test_output/'.$file, getImports($file, $dir));
        }
    }
    return $html5->saveHTML($document);
}

getImports('index.html');
