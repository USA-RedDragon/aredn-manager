const { version } = require('../../package.json');

module.exports.GETApiVersion = (req, res) => {
    res.json({ version: version });
};

module.exports.GETApiPing = (req, res) => {
    res.json({ uptime: process.uptime() });
};
