var postcss = require('postcss');


module.exports = postcss.plugin('ratio', function (opts) {
    opts = opts || {};

    // Work with options here

    return function (css, result) {
        css.walkDecls('ratio', function (decl) {
            var values = postcss.list.space(decl.value);

            var height = parseInt(values[0], 10);
            var width = parseInt(values[1], 10);

            decl.cloneBefore({
                prop: 'padding-top',
                value: (height / width * 100) + '%'
            });

            decl.remove();
        });
    };
});
