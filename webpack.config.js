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
        Panel: './panel/Panel.js',
        Do: './do/Do.js',
        Indicator: './indicator/Indicator.js'
    },
    target: 'electron-renderer',
    output: {
        filename: '[name]-bundle.js',
        path: __dirname + '/bundle'
    },
    devtool: 'source-map'
};
