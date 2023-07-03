const controllers = require('../controllers/cgi-bin');

module.exports = (app) => {
    app.get(`/cgi-bin/sysinfo.json`, controllers.GETSysinfo);
};
