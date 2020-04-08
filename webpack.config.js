module.exports = {
    module: {
        rules: [
            {
                test: /\.js$/,
                exclude: /node_modules/,
                use: {
                    loader: 'babel-loader',
                    options: {
                        'presets': ['@babel/preset-env', '@babel/preset-react'],
                        'plugins': ['@babel/plugin-proposal-class-properties']
                    }
                }
            }
        ]
    },
    entry: {
        renderPanel: './panel/render.js',
        renderDo: './do/render.js',
        renderIndicator: './indicator/render.js'
    },
    target: 'electron-renderer',
    output: {
        filename: '[name]-bundle.js',
        path: __dirname + '/bundle'
    },
    devtool: 'source-map'
};
