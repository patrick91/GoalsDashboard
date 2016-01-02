var path = require('path');
var webpack = require('webpack');
var HtmlWebpackPlugin = require('html-webpack-plugin');
var ExtractTextPlugin = require("extract-text-webpack-plugin");

module.exports = function(options) {
    var entry, jsLoaders, plugins;

    // If production is true
    if (options.prod) {
        // TODO
    } else {
        // Entry
        entry = [
            "webpack-dev-server/client?http://localhost:5000", // Needed for hot reloading
            "webpack/hot/only-dev-server", // See above
            path.resolve(__dirname, 'js/app.js') // Start with js/app.js...
        ];
        // Only plugin is the hot module replacement plugin
        plugins = [
            new webpack.HotModuleReplacementPlugin(), // Make hot loading work
            new HtmlWebpackPlugin({
                template: 'index.html', // Move the index.html file
                inject: true // inject all files that are generated by webpack, e.g. bundle.js, main.css with the correct HTML tags
            })
        ]
    }

    return {
        entry: entry,
        output: { // Compile into js/build.js
            path: path.resolve(__dirname, 'build'),
            filename: "js/bundle.js"
        },
        module: {
            loaders: [{
                test: /\.js$/, // Transform all .js files required somewhere within an entry point...
                loader: 'babel', // ...with the specified loaders...
                exclude: path.join(__dirname, '/node_modules/') // ...except for the node_modules folder.
            }, {
                test: /\.jpe?g$|\.gif$|\.png$/i,
                loader: "url-loader?limit=10000"
            }
        ]},
        plugins: plugins,
        target: "web", // Make web variables accessible to webpack, e.g. window
        stats: false, // Don't show stats in the console
        progress: true
    };
}
