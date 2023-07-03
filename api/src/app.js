const express = require('express');

const app = express();

require('./middleware')(app);

require('./routes')(app);

module.exports = app;
