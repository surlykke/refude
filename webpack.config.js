const WebpackShellPlugin = require('webpack-shell-plugin');

module.exports = {
    module: {
        rules: [
            {
                test: /\.js$/,
                exclude: /node_modules/,
                use: {
                    loader: "babel-loader",
                    options: {
                        "presets": [
                            //"@babel/preset-env",
                            "@babel/preset-react"
                        ],
                        "plugins": [
                            "@babel/plugin-proposal-class-properties"
                        ]
                    }
                }
            }
        ]
    },
    entry: {
        panel: './panel/index.js',
        appchooser: './appchooser/index.js'
    },
    target: 'node',
    output: {
        path: __dirname + '/dist',
        filename: '[name]/index_bundle.js'
    }
};