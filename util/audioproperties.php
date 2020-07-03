<?php
//
// usage: cd audio && php -f ../util/audioproperties.php */*.mp3
//
// PREREQUISITE: github.com/dayvonjersen/taglib-php
//
// This script reads audioProperties from MP3 tag
// and prints them for usage in data.yaml
//
// It's kinda shit and should use soxi instead
//
array_shift($argv);
foreach($argv as $arg) {
    $t = new TagLibMPEG($arg);
    $props = $t->getAudioProperties();
    echo 
        $arg, ' ', 
        $props['bitrate'], 'kbps MP',$props['layer'], 
        ' ', 
        $props['sampleRate'],'kHz ', $props['channelMode'],

        "\n",'      file: ',basename($arg),
        "\n",'      length: ',$props['length'], 
        "\n\n";
}
