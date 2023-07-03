const app = require('./app');

const { name } = require('../package.json');

app.listen(8081, () => console.log(`${name} listening on port 8081!`));
