var gulp = require('gulp');
var concat = require('gulp-concat');
var browserify = require('gulp-browserify');

gulp.task('js', function () {
  return gulp.src('./front/js/app.js')
             .pipe(browserify({
               transform: ['reactify', 'debowerify'],
               debug: !gulp.env.production
             }))
             .pipe(concat('script.js'))
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
  gulp.watch(['./front/**/*'], function () {
    gulp.run('default');
  });
});

gulp.task('default', function () {
  gulp.run('js', 'css', 'fonts');
});
