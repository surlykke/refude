// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
var gulp = require('gulp');
var fs = require("fs");
var browserify = require("browserify");
var babelify = require("babelify");
var source = require('vinyl-source-stream');
var gutil = require('gulp-util');

gulp.task('assets', function() {
	return gulp
		.src(['main.html', 'package.json', 'refudeNotifications', 'refudeNotifications.desktop'])
		.pipe(gulp.dest('../dist/notifications'))
})

gulp.task('js', function() {
	browserify({ entries: ["./notifications.jsx"], extensions: [".jsx", ".js"], node: true, debug: true })
        .transform(babelify, {presets: ["@babel/react", "@babel/env"], plugins:["@babel/plugin-proposal-class-properties"]})
		.bundle()
		.on('error',gutil.log)
		.pipe(source('bundle.js'))
    	.pipe(gulp.dest('../dist/notifications'));
});

gulp.task('default', ['assets', 'js']);

gulp.task('watch',function() {
	gulp.start('default')
	gulp.watch(['*', '../common/*'],['default'])
});

gulp.task('run', ['default', 'watch'])