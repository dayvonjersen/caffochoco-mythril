<?php
$id3v2 = parse_ini_file('id3v2.ini');
$frames = ['TPE1','TIT2','TALB','TCON','COMM','TDRC'];

array_shift($argv);
foreach($argv as $arg) {
    echo $arg,"\n";
    $t = new TagLibMPEG($arg);
    if($t->hasID3v1()) {
        echo 'ID3v1:',"\n";
        foreach($t->getID3v1() as $frameID => $data) {
            echo "\t$frameID: $data\n";
        }
    }
    echo 'ID3v2:',"\n";
    $read = $write = [];
    foreach($t->getID3v2() as $frame) {
        echo "\t", $id3v2[$frame['frameID']], ': ', $frame['data'], "\n";
        $read[$frame['frameID']] = $frame['data'];
    }
    if(empty($read)) echo '(empty)';
    echo "\n";
    foreach($frames as $frameID) {
        echo $id3v2[$frameID], isset($read[$frameID]) ? ' ('.$read[$frameID].')' : '', ': ';
        $stdin = fopen('php://stdin', 'r');
        $input = trim(fgets($stdin));
        fclose($stdin);
        if($input == '' && isset($read[$frameID])) $input = $read[$frameID];
        $write[$frameID] = $input;
    }
    $t->stripTags();
    $t->setID3v2($write);
    echo "\n\n-- OK --\n";
}
