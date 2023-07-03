const controllers = require('../controllers/meta');

module.exports = (app) => {
    app.get(`/api/version`, controllers.GETApiVersion);
    app.get(`/api/ping`, controllers.GETApiPing);
};
