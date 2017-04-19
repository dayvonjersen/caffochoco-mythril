<?php
function call($cmd) {
    exec("$cmd 2>&1", $output, $exit_code);
    if($exit_code !== 0) {
        echo implode("\n", $output), "\n";
        exit($exit_code);
    }
    return $output;
}

require_once 'vendor/autoload.php';
use Masterminds\HTML5;

$document = new DOMDocument();
$document->formatOutput = false;
$document->preserveWhitespace = false;
$html5 = new HTML5([
    'encode_entities' => false,
    'disable_html_ns' => true,
    'target_document' => $document,
]);
$html5->loadHTML(file_get_contents($argv[1]));
$document->normalizeDocument();

$nodes = [];
$nodeList = $document->getElementsByTagName('script');
$i = 0;
foreach($nodeList as $node) {
    $tmpfile = sprintf('tmp/script_%03d.js', $i);
    file_put_contents($tmpfile, $node->textContent . ';');
    $i++;
    $nodes[] = $node;
}

$nodeList = $document->getElementsByTagName('style');
$i = 0;
foreach($nodeList as $node) {
    if(!$node->hasAttribute('is')) {
    $tmpfile = sprintf('tmp/style_%03d.css', $i);
    file_put_contents($tmpfile, $node->textContent);
    $i++;

    exec('bash -c "csso '.$tmpfile.' --restructure-off" 2>&1', $out, $exit_code);
    if($exit_code == 0) {
        $node->nodeValue = $out[0];
    } else {
        echo 'WARN: ', $out[0], "\n";
    }
    unset($out, $exit_code);
    }
}

foreach($nodes as $node) {
    $node->parentNode->removeChild($node);
}

$head = $document->getElementsByTagName('head')->item(0);

$scriptElement = $document->createElement('script');
$scriptElement->setAttribute('src', '/app.min.js');
$head->appendChild($scriptElement);

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
$output = str_replace('&NewLine;', "\n", $output);

file_put_contents('tmp/index.min.html', $output); 
/* echo $output; */
