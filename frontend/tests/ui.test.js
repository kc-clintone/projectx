const { test } = require('uvu');
const assert = require('uvu/assert');
const fs = require('fs');
const cheerio = require('cheerio');

// Simple static rendering test: load built index.html (or source) and assert login text exists
test('index contains welcome text', () => {
  const htmlPath = __dirname + '/../index.html';
  const html = fs.readFileSync(htmlPath, 'utf8');
  const $ = cheerio.load(html);
  const txt = $('body').text();
  assert.ok(txt.includes('Welcome to Study Coach'));
});

test.run();
