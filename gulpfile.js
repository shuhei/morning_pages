var gulp = require('gulp');
var concat = require('gulp-concat');
var react = require('gulp-react');
var es = require('event-stream');

gulp.task('js', function () {
  var libs = [
    './bower_components/jquery/jquery.js',
    './bower_components/bootstrap/dist/js/bootstrap.js',
    './bower_components/underscore/underscore.js',
    './bower_components/backbone/backbone.js',
    './bower_components/react/react-with-addons.js'
  ];
  return es.concat(
    gulp.src(libs),
    gulp.src('./front/jsx/*.jsx')
        .pipe(react())
  ).pipe(concat('script.js'))
   .pipe(gulp.dest('./public/js'));
});

gulp.task('css', function () {
  var styles = [
    './bower_components/bootstrap/dist/css/bootstrap.css',
    './bower_components/font-awesome/css/font-awesome.css',
    './front/css/*.css'
  ];
  gulp.src(styles)
      .pipe(concat('style.css'))
      .pipe(gulp.dest('./public/css'));
});

gulp.task('fonts', function () {
  gulp.src('./bower_components/font-awesome/fonts/*')
      .pipe(gulp.dest('./public/fonts'));
});

gulp.task('default', function () {
  gulp.run('js', 'css', 'fonts');
});
