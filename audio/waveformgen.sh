#!/bin/bash

set -e

dname=${1%%/*}
color=`vibrant ../image/$dname.jpg | awk '{if($1=="Vibrant:"||$1=="LightVibrant:") print toupper(substr($2, 2, length($2)-2))}'`FF
echo $dname: $color
for f in $@; do
    fname=${f%.mp3}
    bname=${fname##*/}
    duration=`soxi -D $f`
    lame --decode --mp3input $f /tmp/$bname.wav && 
        audiowaveform -i /tmp/$bname.wav -o $fname.png --no-axis-labels --background-color 00000000 --waveform-color $color -e $duration -w 1200 -h 200 &
    echo "-- OK --"
done
wait;
