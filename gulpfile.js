var gulp = require('gulp');
var concat = require('gulp-concat');
var gutil = require('gulp-util');
var plumber = require('gulp-plumber');
var browserify = require('gulp-browserify');

var libs = ['jquery', 'underscore', 'backbone', 'react', 'domready'];
var styles = [
  './bower_components/bootstrap/dist/css/bootstrap.css',
  './bower_components/font-awesome/css/font-awesome.css',
  './front/css/*.css'
];

gulp.task('js', function () {
  return gulp.src('./front/js/app.js')
             .pipe(plumber())
             .pipe(browserify({ transform: ['reactify'], debug: !gulp.env.production }))
             .on('prebundle', function (bundler) {
                libs.forEach(function (lib) { bundler.external(lib); });
             })
             .pipe(gulp.dest('./public/js'))
             .on('error', gutil.log);
});

gulp.task('lib', function () {
  return gulp.src('./front/js/lib.js')
             .pipe(plumber())
             .pipe(browserify({ debug: !gulp.env.production }))
             .on('prebundle', function (bundler) {
               libs.forEach(function (lib) { bundler.require(lib); });
             })
             .pipe(gulp.dest('./public/js'))
             .on('error', gutil.log);
});

gulp.task('css', function () {
  return gulp.src(styles)
             .pipe(plumber())
             .pipe(concat('style.css'))
             .pipe(gulp.dest('./public/css'))
             .on('error', gutil.log);
});

gulp.task('fonts', function () {
  return gulp.src('./bower_components/font-awesome/fonts/*')
             .pipe(gulp.dest('./public/fonts'));
});

gulp.task('watch', function () {
  gulp.watch('./front/js/**/*', function () {
    gulp.run('js');
  });

  gulp.watch('./front/css/**/*', function () {
    gulp.run('css');
  });
});

gulp.task('default', function () {
  gulp.run('js', 'lib', 'css', 'fonts');
});
