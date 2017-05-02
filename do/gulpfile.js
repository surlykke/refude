var gulp = require('gulp');
var fs = require("fs");
var browserify = require("browserify");
var babelify = require("babelify");
var source = require('vinyl-source-stream');
var gutil = require('gulp-util');

gulp.task('assets', function() {
	return gulp
		.src(['main.html', '../common/refude.css', 'package.json', 'refudeDo'])
		.pipe(gulp.dest('../dist/do'))
})

gulp.task('js', function() {
	browserify({ entries: ["./container.jsx"], extensions: [".jsx", ".js"], debug: true })
		.transform(babelify, {presets: ["react", "es2015", "stage-0"]})
		.bundle()
		.on('error',gutil.log)
		.pipe(source('bundle.js'))
    	.pipe(gulp.dest('../dist/do'));
});

gulp.task('default', ['assets', 'js']);

gulp.task('watch',function() {
	gulp.watch(['./**/*'],['default'])
});

gulp.task('run', ['default', 'watch'])
