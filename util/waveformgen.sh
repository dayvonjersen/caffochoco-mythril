#!/bin/bash
#
# usage: cd audio && ../util/waveformgen.sh */*.mp3
#
# PREREQUISITE: github.com/bbc/audiowaveform
# PREREQUISITE: lame, soxi
#
# creates the waveform images for audio files
#
set -e

for f in $@; do
    fname=${f%.mp3}
    bname=${fname##*/}
    duration=`soxi -D "$f"`
    lame --decode --mp3input "$f" /tmp/$bname.wav && 
        audiowaveform -i "/tmp/$bname.wav" -o "$fname.png" --no-axis-labels --background-color 00000000 --waveform-color ffffffff -e $duration -w 1200 -h 200 &
    echo "-- OK --"
done
wait;
