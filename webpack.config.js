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
            },
            {
                test: /\.css$/,
                use: ['style-loader', 'css-loader']
            }
        ]
    },
    entry: {
        Panel: './panel/Panel.js',
        Do: './do/Do.js',
        Osd: './osd/Osd.js',
    },
    target: 'electron-renderer',
    output: {
        filename: '[name]-bundle.js',
        path: __dirname + '/bundle'
    },
    devtool: 'source-map'
};
