const fs = require('fs');
const path = require('path');

module.exports = (app) => {
    fs.readdirSync(__dirname)
        .filter((file) => {
            return (file.indexOf('.') !== 0) && (file !== 'index.js');
        })
        .forEach((file) => {
            require(path.join(__dirname, file))(app);
        });
};
