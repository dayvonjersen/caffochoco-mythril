<?php
/**
 * usage: cd audio && php -f ../util/tagger-auto.php
 *
 * --> USE THIS ONE <--
 *
 * PREREQUISITE: github.com/dayvonjersen/taglib-php
 *
 * Uses information in data.json to find and tag all releases
 * including adding album art to the tag
 *
 * Alerts for missing files
 */
$frames = [
    'TPE1' => 'Artist',
    'TIT2' => 'Title',
    'TALB' => 'Album',
    'TCON' => 'Genre',
    'COMM' => 'Comment',
    'TDRC' => 'Year',
];
$art = [];

$d  = json_decode(file_get_contents('../data.json'));
$tracklists = [];
foreach($d->tracklists as $tl) {
    $tracklists[$tl->id] = $tl;
}
$tracks = [];
foreach($d->tracks as $t) {
    $tracks[$t->id] = $t;
}
foreach($d->releases as $r) {
    if(!file_exists(''.$r->url)) {
        echo 'missing directory ', $r->url, "\n";
    } else {
        foreach($r->tracklists as $tl) {
            $tracklist = $tracklists[$tl];
            foreach($tracklist->tracks as $t) {
                $track = $tracks[$t];
                if(!file_exists(''.$r->url.'/'.$track->file)) {
                    echo 'missing file '.$r->url.'/'.$track->file, "\n";
                } else {
                    $artist = $title = $album = '';
                    if(strpos($track->title, ' - ') !== false) {
                        list($artist, $title) = explode(' - ', $track->title);
                        $title = str_replace('"', '', $title);
                    } else {
                        $artist = $r->artist;
                        $title = $track->title;
                    }
                    if($r->category == 'album' || $r->category == 'ep' || $r->category == 'remix') {
                        $album = $r->title;
                    }
                    if($tracklist->title !== 'tracklist') {
                        $album = $r->title . ' ' . $tracklist->title;
                    }
                    $tags = [
                        'TPE1' => $artist,
                        'TIT2' => $title,
                        'TALB' => $album,
                        'TCON' => $r->genre,
                        'COMM' => trim(strip_tags($r->about)),
                        'TDRC' => sprintf("%d", $r->year),
                    ];
                    
                    echo $r->url, '/', $track->file, ':', "\n";
                    foreach($tags as $frameID => $tag) {
                        echo "\t", $frames[$frameID], ': ', $tag, "\n";
                    }
                    $taglib = new TagLibMPEG($r->url .'/'.$track->file);
                    $id3v2 = [];
                    foreach($taglib->getID3v2() as $frame) {
                        $id3v2[$frame['frameID']] = $frame['data'];
                    }
                    $same = true;
                    foreach($tags as $field => $value) {
                        if(!isset($id3v2[$field]) || $id3v2[$field] !== $value) {
                            $same = false;
                            break;
                        }
                    }
                    if($same) {
                        echo "\nUP-TO-DATE\n[SKIP]\n";
                    } else {
                        echo "\nWRITING TAG...\n";
                        $taglib->stripTags();
                        $taglib->setID3v2($tags);
                        if(file_exists('../image/'.$r->url.'.jpg')) {
                            echo "WRITING ALBUM ART...\n";
                            if(!isset($art[$r->url])) {
                                $art[$r->url] = base64_encode(file_get_contents('../image/'.$r->url.'.jpg'));
                            }
                            $data = $art[$r->url];
                            $taglib->setID3v2([
                                'APIC' => [
                                    'data' => $data,
                                    'mime' => 'image/jpeg',
                                    'type' => TagLib::APIC_FRONTCOVER,
                                ]
                            ]);
                        } else {
                            echo "NO ALBUM ART\n";
                        }
                        echo "[ OK ]\n";
                    }
                }
            }
        }
    }
}
