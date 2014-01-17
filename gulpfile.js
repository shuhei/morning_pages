var gulp = require('gulp');
var concat = require('gulp-concat');
var browserify = require('gulp-browserify');

var libs = ['jquery', 'underscore', 'backbone', 'react', 'domready'];

gulp.task('js', function () {
  return gulp.src('./front/js/app.js')
             .pipe(browserify({ transform: ['reactify'], debug: !gulp.env.production }))
             .on('prebundle', function (bundler) {
                libs.forEach(function (lib) { bundler.external(lib); });
             })
             .pipe(gulp.dest('./public/js'));
});

gulp.task('lib', function () {
  return gulp.src('./front/js/lib.js')
             .pipe(browserify({ debug: !gulp.env.production }))
             .on('prebundle', function (bundler) {
               libs.forEach(function (lib) { bundler.require(lib); });
             })
             .pipe(gulp.dest('./public/js'));
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
