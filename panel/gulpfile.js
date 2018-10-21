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
var es = require('event-stream');
var rename = require('gulp-rename');

gulp.task('assets', function() {
	return gulp
		.src(['**/*.html', '**/*.css', 'package.json', 'refudePanel', 'refudePanel.desktop', 'refudeDo'])
		.pipe(gulp.dest('../dist/panel'))
})

gulp.task('js', function() {
    var files = ['panel.jsx', 'do/indicator.jsx'];
    var tasks = files.map((entry) =>
        browserify({entries: [entry], extensions: [".jsx", ".js"], node:true, debug: true })
            .transform(babelify, {presets: ["@babel/react", "@babel/env"], plugins:["@babel/plugin-proposal-class-properties"]})
            .bundle()
            .pipe(source(entry))
            .pipe(rename({extname: '.bundle.js'}))
            .pipe(gulp.dest('../dist/panel'))
        );
    return es.merge.apply(null, tasks);
});

gulp.task('default', ['assets', 'js']);

gulp.task('watch',function() {
	gulp.start('default')
	gulp.watch(['./**/*', '../common/*'],['default'])
});

gulp.task('run', ['default', 'watch'])
