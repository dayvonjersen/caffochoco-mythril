"use strict";
const gulp   = require('gulp');
const sass   = require('gulp-sass');
const exec   = require('child_process').exec;

gulp.task('sass', () => {
    return gulp
        .src('*.sass')
        .pipe(sass().on('error', sass.logError))
        .pipe(gulp.dest('./css/'));
});

gulp.task('data', (cb) => {
    exec(
      "ruby -ryaml -rjson -e 'puts JSON.generate(YAML.load(ARGF))' < data.yaml > data.json",
      (err, stdout, stderr) => {
        cb(err);
        console.log(stdout, stderr);
      }
    )
      .on('exit', (code) => console.log(!code ? 'OK' : 'Fail'));
});

gulp.task('watch', () => {
    gulp.watch('*.sass', ['sass']);
    gulp.watch('data.yaml', ['data']);
});

gulp.task('default', ['watch']);
