const http = require('http');
const fs = require('fs');
const path = require('path');

const PORT = 3000;
const MIME_TYPES = {
  '.html': 'text/html; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.js': 'application/javascript; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.svg': 'image/svg+xml'
};

const server = http.createServer((req, res) => {
  let reqPath = req.url.split('?')[0];
  if (reqPath === '/' || reqPath === '/preview') {
    reqPath = '/preview/index.html';
  }

  // Resolve file path relative to root directory
  let filePath = path.join(__dirname, '..', reqPath);

  // If pointing to static inside preview
  if (!fs.existsSync(filePath)) {
    filePath = path.join(__dirname, reqPath);
  }

  fs.stat(filePath, (err, stats) => {
    if (err || !stats.isFile()) {
      res.writeHead(404, { 'Content-Type': 'text/plain; charset=utf-8' });
      res.end('404 Halaman Tidak Ditemukan');
      return;
    }

    const ext = path.extname(filePath).toLowerCase();
    const contentType = MIME_TYPES[ext] || 'application/octet-stream';

    res.writeHead(200, { 'Content-Type': contentType });
    fs.createReadStream(filePath).pipe(res);
  });
});

server.listen(PORT, '127.0.0.1', () => {
  console.log(`====================================================`);
  console.log(`Pos Alat Gudang (Frontend Preview) berjalan!`);
  console.log(`Buka browser di: http://127.0.0.1:${PORT}`);
  console.log(`====================================================`);
});

