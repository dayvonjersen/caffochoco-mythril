"use strict";
const gulp = require('gulp');
const sass = require('gulp-sass');

gulp.task('sass', () => {
    return gulp
        .src('*.sass')
        .pipe(sass().on('error', sass.logError))
        .pipe(gulp.dest('.'));
});

gulp.task('watch', () => {
    gulp.watch('*.sass', ['sass']);
});

gulp.task('default', ['watch']);
