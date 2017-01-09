// observer pattern
function Subject() {
    var observers = new Set();

    return {
        add: function(item) {
            observers.add(item);
        },
        remove: function(item) {
            observers.delete(item);
        },
        removeAll: function() {
            observers.clear();
        },
        notifyObservers: function(data) {
            observers.forEach((o) => o.notify(data));
        }
    }
};

class PlayerModel {
  constructor(audioElement) {
    // const
    this.DEFAULT_VOLUME = 0.75;

    // observer pattern
    this.subject = new Subject();

    // instance variables
    this.state = {
      nowPlayingIndex: 0,
      playlist: [],
      previousVolume: 0,

      isPlaying: false,
      currentTime: 0,

      tracklistId: 0,
    };

    //
    this.audioElement = audioElement;
    this.audioElement.addEventListener('ended', () => {
      this.next();
      this.togglePlayback();
    });
    this.audioElement.addEventListener('timeupdate', () => {
      this.state.currentTime = this.audioElement.currentTime;
      this.update();
    });
  }

  // observer pattern
  update() {
    this.state.isPlaying = !this.audioElement.paused;
    this.state.nowPlaying = this.nowPlaying;
    this.subject.notifyObservers(this.state);
  }

  // playlist
  get nowPlaying() {
    if(this.state.playlist.length) {
      return this.state.playlist[this.state.nowPlayingIndex];
    }
    return '';
  }

  previous() {
    if(!this.state.playlist.length) return;

    --this.state.nowPlayingIndex;

    if(this.state.nowPlayingIndex < 0) {
      this.state.nowPlayingIndex = this.state.playlist.length - 1
    }

    this.load();
    this.update();
  }

  next() {
    if(!this.state.playlist.length) return;

    ++this.state.nowPlayingIndex;

    if(this.state.nowPlayingIndex >= this.state.playlist.length) {
      this.state.nowPlayingIndex = 0;
    }

    this.load();
    this.update();
  }

  add(srcUrl) {
    let playnow = this.state.playlist.length === 0;
    this.state.playlist.push(srcUrl);
    if(playnow) {
      this.load();
    }
    this.update();
  }

  clear() {
    this.state.playlist = this.state.nowPlayingIndex > 0 ? [this.nowPlaying] : [];
    this.state.nowPlayingIndex = 0;
    this.load();
    this.update();
  }

  load() {
    this.audioElement.src = this.nowPlaying;
    if(this.isPlaying) this.audioElement.play();
  }

  // playback
  togglePlayback() {
    if(this.audioElement.paused) {
      this.audioElement.play();
    } else {
      this.audioElement.pause();
    }

    this.update();
  }

  stopPlayback() {
    if(!this.audioElement.paused) {
      this.audioElement.pause();
    }

    if(this.audioElement.readyState != HTMLMediaElement.HAVE_NOTHING) {
      this.audioElement.currentTime = 0;
    }

    this.update();
  }

  // volume
  toggleMute(force=null) {
    let mute;
    let volume = this.audioElement.volume;
    if(force !== null) {
      mute = !!force;
      if((mute && volume === 0) || (!mute && volume > 0)) {
        return;
      }
    } else {
      mute = !(volume === 0);
    }
    if(!this.state.previousVolume) {
      this.state.previousVolume = this.DEFAULT_VOLUME;
    }

    if(mute) {
      this.state.previousVolume = volume;
      this.audioElement.volume = 0;
    } else {
      this.audioElement.volume = this.state.previousVolume;
    }

    this.update();
  }

  setVolume(volume) {
    if(volume < 0)   volume = 0;
    if(volume > 100) volume = 100;

    var v = volume/100;
    this.audioElement.volume = v;

    this.previousVolume = this.audioElement.volume;

    this.update();
  }

  // seek
  seekTo(position) {
      position = parseInt(position);
      if(position < 0) {
        position = 0;
      }
      if(position > this.audioElement.duration) {
        position = this.audioElement.duration;
      }
      this.audioElement.currentTime = position;
      this.update();
  }


}

class PlayerView {
  notify(playerState) {
    // update something
  }
}

class PlayerController {
  constructor() {}
}
