var gulp = require('gulp');
var concat = require('gulp-concat');
var gutil = require('gulp-util');
var browserify = require('gulp-browserify');
var multipipe = require('multipipe');

var libs = ['jquery', 'underscore', 'backbone', 'react', 'domready'];
var styles = [
  './bower_components/bootstrap/dist/css/bootstrap.css',
  './bower_components/font-awesome/css/font-awesome.css',
  './front/css/*.css'
];

function handleError(e) {
  console.log(e.toString());
  this.emit('end');
}

function pipes() {
  return multipipe.apply(null, arguments)
                  .on('error', handleError);
}

gulp.task('js', function () {
  return pipes(
    gulp.src('./front/js/app.js'),
    browserify({
      transform: ['reactify'],
      external: libs,
      debug: !gutil.env.production
    }),
    gulp.dest('./public/js')
  );
});

gulp.task('lib', function () {
  return pipes(
    gulp.src('./front/js/lib.js'),
    browserify({
      require: libs,
      debug: !gutil.env.production
    }),
    gulp.dest('./public/js')
  );
});

gulp.task('css', function () {
  return pipes(
    gulp.src(styles),
    concat('style.css'),
    gulp.dest('./public/css')
  );
});

gulp.task('fonts', function () {
  return pipes(
    gulp.src('./bower_components/font-awesome/fonts/*'),
    gulp.dest('./public/fonts')
  );
});

gulp.task('watch', function () {
  gulp.watch('./front/js/**/*', ['js']);
  gulp.watch('./front/css/**/*', ['css']);
});

gulp.task('default', ['js', 'lib', 'css', 'fonts']);
