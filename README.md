# <small><small>codename:</small></small> caffochoco-mythril

**Polymer 2 app for a browsable music collection**

Designed specifically for [music.dayvonjersen.com](https://music.dayvonjersen.com/)

---

>Maybe writing a README will be more motivating than writing a post-mortem.
>
>I will be interspersing bits of blog in this README formatted in sections
>like these.
>
>Feel free to ignore them

---

## Installation

Make a `data.yaml` file with the following structure:

```yaml
releases:
  - id: 1
    artist: "Unknown Artist"
    title: "Untitled"
    year: 2017
    genre: "Hip-Hop"
    url: "unknown-artist-untitled"
    tracklists: [1]
    defaultTracklist: 1
    category: "single"
    about: |
        <p>y helo thar</p>
tracklists:
  - id: 1
    title: "tracklist"
    tracks: [1]
tracks:
  - id: 1
    title: "Untitled"
    file: "untitled.mp3"
    length: 456
```

### Development

Run `make setup`, then `make serve` and finally `./serve` to start a development
server on localhost:8080

Refer to `Makefile` and `./serve -h` for more information.

### Production

Run `make dist` to minify the app and `./serve -prod` to serve for production.

See `util/make-dist.sh` for more information.

---

>NOTE: **Be warned** that Polymer 2 + uglifyjs2 + Firefox is a fail-tastic combination
and you will have to edit some core Polymer code somewhere in `bower_components/` in order to get
Firefox to play with the minified app nicely.
>
>A reasonable way to debug this headache is to find the error in the Firefox developer console
and `grep -r` for whatever it's complaining about then compare that code with the (beautified)
`app.min.js`.
>
>Most of the time something `undefined` is being called or accessed so it's usually just a matter of
adding `if(!!problematicVariableOrFunctionNameHere)` around the code in question.
>
>Then call `make dist` again.

---

## Usage

### Adding music

#### Metadata

Refer to the example yaml file above.

**Running `gulp` will watch `data.yaml` for changes and update `data.json`**

*all* `id`s must be unique so you'll have to increment them manually and keep track of them

*release* `artist`, `title`, `year`, and `genre` should be self-explanatory

*release* `about` can contain HTML for formatting

*release* `tracklists`: releases can have multiple tracklists, referred to by their respective `id`

*release* `defaultTracklist` even if a release has only one tracklist, the `id` must be specified as well

*release* `category` can be one of "single", "remix", "ep", "album", or "mashup"

*release* `url`:

>All audio tracks live in subdirectories of `audio/`. So
>
>`mkdir audio/unknown-artist-untitled`
>
>the `url` field is used to refer to a release in several places throughout the app
so it must be unique

*tracklist* `title` should simply be "tracklist" most of the time, but if you have more than one
tracklist for a release you can diffrentiate among them using this field, e.g. "extended mixes"

*tracklist* `tracks`: tracklists can (obviously) have multiple tracks, referred to by their respective `id`.

*track* `title`: should be self-explanatory

*track* `file`: the name of the audio file in the subdirectory of `audio/` as specified by release `url`.

*track* `length` is track duration in seconds. Use `soxi -D` to find this out.

#### Album Art

Add album art to the `image/` directory with a filename corresponding to its release `url`

e.g. `unknown-artist-untitled.jpg`

NOTE: **ONLY** JPEGs with file extension `.jpg` are recognized.

Then generate thumbnails using imagemagick (install imagemagick first...)

`convert --resize 600x600^ -gravity Center -crop 600x600+0+0 +repage image/unknown-artist-untitled.jpg image/unknown-artist-untitled--square.jpg`

#### Writing ID3v2 tags

You'll need [github.com/generaltso/taglib-php](https://github.com/generaltso/taglib-php) first.

`cd audio && php -f ../util/tagger-auto.php`

This will write the metadata specified in `data.yaml` and the album art to the mp3s.

#### Generating waveforms

Use `util/waveformgen.sh` to generate the waveforms that appear in the player.

You'll need to install [github.com/bbc/audiowaveform](https://github.com/bbc/audiowaveform) first (linux/mac only)

Additionally, you'll need to install LAME and soxi.

Then you can `cd audio && ../util/waveformgen.sh */*.mp3`

#### Generating zip files for downloading

Finally, run `./serve precache` to generate zip files

The server can generate zip files on-the-fly but this is unacceptable for production.

---

> Admittedly it's a lot of manual data-entry but that's how you add music releases.

---

### Oh wait that's all there is to it.

You can take a look at the code if you want to figure out how everything works
but for now, I'd just like to explain my motivation for making this, some of my
decisions for the application design, and how it was working with Polymer 2.

*(to be continued...)*
