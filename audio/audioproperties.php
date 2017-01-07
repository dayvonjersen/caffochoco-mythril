<?php
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
