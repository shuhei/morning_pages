var gulp = require('gulp');
var concat = require('gulp-concat');
var browserify = require('browserify');
var source = require('vinyl-source-stream');

var libs = ['jquery', 'underscore', 'backbone', 'react'];

gulp.task('js', function () {
  var b = browserify('./front/js/app.js');
  libs.forEach(function (l) { b.external(l); });
  b.transform('reactify');
  return b.bundle()
          .pipe(source('app.js'))
          .pipe(gulp.dest('./public/js/app.js'));
});

gulp.task('lib', function () {
  var b = browserify('./front/js/lib.js');
  libs.forEach(function (l) { b.require(l); });
  return b.bundle()
          .pipe(source('lib.js'))
          .pipe(gulp.dest('./public/js/lib.js'));
});

gulp.task('css', function () {
  var styles = [
    './bower_components/bootstrap/dist/css/bootstrap.css',
    './bower_components/font-awesome/css/font-awesome.css',
    './front/css/*.css'
  ];
  return gulp.src(styles)
             .pipe(concat('style.css'))
             .pipe(gulp.dest('./public/css'));
});

gulp.task('fonts', function () {
  return gulp.src('./bower_components/font-awesome/fonts/*')
             .pipe(gulp.dest('./public/fonts'));
});

gulp.task('watch', function () {
  gulp.watch(['./front/js/**/*', './front/css/**/*'], function () {
    gulp.run('js', 'css');
  });
});

gulp.task('default', function () {
  gulp.run('js', 'lib', 'css', 'fonts');
});
