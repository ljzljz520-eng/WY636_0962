const fs = require('fs');
const path = require('path');
const source = fs.readFileSync(path.join(__dirname, 'src', 'index.html'), 'utf8');
fs.mkdirSync(path.join(__dirname, 'dist'), { recursive: true });
fs.writeFileSync(path.join(__dirname, 'dist', 'index.html'), source);
process.stdout.write('built animal cage console\n');
