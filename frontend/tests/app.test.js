const { test } = require('uvu');
const assert = require('uvu/assert');
const fs = require('fs');
const cheerio = require('cheerio');

test('index.html has expected root elements and scripts', () => {
  const htmlPath = __dirname + '/../index.html';
  const html = fs.readFileSync(htmlPath, 'utf8');
  const $ = cheerio.load(html);
  assert.ok($('#root').length === 1, 'root div exists');
  assert.ok($('#confetti-canvas').length === 1, 'confetti canvas exists');
  // app.js should be referenced
  const scripts = $('script').map((i, el) => $(el).attr('src')).get();
  assert.ok(scripts.some(s => s && s.includes('app.js')), 'app.js script tag present');
});

test('app.js contains expected component names', () => {
  const appJs = fs.readFileSync(__dirname + '/../app.js', 'utf8');
  // Look for key component/function names to catch regressions
  assert.ok(appJs.includes('ConfettiCanvas'), 'ConfettiCanvas present in app.js');
  assert.ok(appJs.includes('CreateSession'), 'CreateSession present in app.js');
  assert.ok(appJs.includes('SessionPlayer'), 'SessionPlayer present in app.js');
});

test.run();
