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
    if($t->hasID3v2()) {
        foreach($t->getID3v2() as $frame) {
            if($frame['frameID'] == 'APIC') {
                echo "\tAlbum Art: (yes)\n";
            } else {
               echo "\t", $id3v2[$frame['frameID']], ': ', $frame['data'], "\n";
            }
            $read[$frame['frameID']] = $frame['data'];
        }
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
    if(file_exists('../image/' . dirname($arg) . '.jpg')) {
        echo "Found Album Art.\n";
        $write['APIC'] = [
            'data' => base64_encode(file_get_contents('../image/' . dirname($arg) . '.jpg')),
            'mime' => 'image/jpeg',
            'type' => TagLib::APIC_FRONTCOVER,
        ];
    }
    $t->stripTags();
    $t->setID3v2($write);
    echo "\n\n-- OK --\n";
}
