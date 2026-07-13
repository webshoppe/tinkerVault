/**
 * Tier 3 cross-engine verification for whiteboard.html (and release copy).
 * Tries Chromium (Chrome), Firefox, and WebKit when binaries are available.
 */
import { chromium, firefox, webkit } from 'playwright-core';
import path from 'path';
import fs from 'fs';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const CHROME = process.env.CHROME_PATH || 'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe';
const FIREFOX = process.env.FIREFOX_PATH || 'C:\\Program Files\\Mozilla Firefox\\firefox.exe';

const targets = [
  {
    name: 'working-copy',
    htmlPath: path.resolve(__dirname, 'whiteboard.html')
  },
  {
    name: 'release-copy',
    htmlPath: path.resolve(__dirname, 'release', 'whiteboard.html')
  }
];

function fileUrl(p) {
  return 'file:///' + p.replace(/\\/g, '/');
}

function engineConfigs() {
  const list = [];
  // Chromium via system Chrome
  if (fs.existsSync(CHROME) || process.platform !== 'win32') {
    list.push({
      id: 'chromium',
      label: 'Chromium/Chrome',
      launch: () =>
        chromium.launch({
          executablePath: fs.existsSync(CHROME) ? CHROME : undefined,
          headless: true,
          args: ['--allow-file-access-from-files', '--disable-web-security']
        })
    });
  }
  // Firefox — prefer Playwright-managed browser (avoids system version skew)
  list.push({
    id: 'firefox',
    label: 'Firefox',
    launch: () => firefox.launch({ headless: true })
  });
  // WebKit — Playwright-managed
  list.push({
    id: 'webkit',
    label: 'WebKit',
    launch: () => webkit.launch({ headless: true })
  });
  return list;
}

async function runSuite(browser, htmlPath, engineId) {
  const results = [];
  const pass = (name, detail = '') => results.push({ ok: true, name, detail });
  const fail = (name, detail = '') => results.push({ ok: false, name, detail });
  const assert = (cond, name, detail = '') => (cond ? pass : fail)(name, detail);

  const context = await browser.newContext({ viewport: { width: 1200, height: 800 } });
  const page = await context.newPage();
  const consoleErrors = [];
  page.on('console', (m) => {
    if (m.type() === 'error') consoleErrors.push(m.text());
  });
  page.on('pageerror', (e) => consoleErrors.push(String(e)));

  const url = fileUrl(htmlPath);
  await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 30000 });
  await page.evaluate(() => localStorage.clear());
  await page.reload({ waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(400);
  await page.waitForFunction(() => {
    const r = document.getElementById('board-root');
    return r && !r.classList.contains('switching') && window.__whiteboard;
  }, { timeout: 8000 }).catch(() => {});

  assert(await page.title() === 'Whiteboard', 'title');
  const ver = await page.evaluate(() => window.__whiteboard && window.__whiteboard.VERSION);
  assert(ver === '1.0.0', 'VERSION constant', String(ver));
  const badge = await page.locator('#version-badge').textContent();
  assert(/v?1\.0\.0/.test((badge || '').trim()), 'version badge', badge);

  const favicon = await page.evaluate(() => {
    const link = document.querySelector('link[rel="icon"]');
    return link ? link.getAttribute('href') || '' : '';
  });
  assert(favicon.startsWith('data:image/svg+xml'), 'inline SVG favicon', favicon.slice(0, 40));

  const templates = await page.evaluate(() =>
    window.__whiteboard.TemplateRegistry.list().map((t) => t.id).sort()
  );
  assert(templates.join(',') === 'notepad,paint,snip,wordpad', 'four templates', templates.join(','));

  // Paint smoke
  const canvas = page.locator('.paint-canvas-wrap canvas');
  await canvas.waitFor({ state: 'visible', timeout: 5000 }).catch(() => {});
  const box = await canvas.boundingBox();
  assert(!!box && box.width > 50, 'paint canvas', JSON.stringify(box));
  if (box) {
    await page.mouse.move(box.x + 30, box.y + 30);
    await page.mouse.down();
    await page.mouse.move(box.x + 100, box.y + 80, { steps: 8 });
    await page.mouse.up();
    await page.waitForTimeout(200);
    const ink = await page.evaluate(() => {
      const c = document.querySelector('.paint-canvas-wrap canvas');
      if (!c) return 0;
      const d = c.getContext('2d').getImageData(0, 0, c.width, c.height).data;
      let n = 0;
      for (let i = 0; i < d.length; i += 4) if (d[i] < 250 || d[i + 1] < 250 || d[i + 2] < 250) n++;
      return n;
    });
    assert(ink > 20, 'paint stroke', `pixels=${ink}`);
  }

  // New board (Annotate)
  await page.locator('#btn-new-board').click();
  await page.locator('.modal').waitFor({ state: 'visible', timeout: 5000 });
  await page.locator('.modal input').fill('T3 Annotate');
  await page.locator('.template-picker button').filter({ hasText: 'Annotate' }).click();
  await page.locator('.modal-actions button.active').click();
  await page.waitForTimeout(250);
  await page.waitForFunction(() => {
    const r = document.getElementById('board-root');
    return r && !r.classList.contains('switching');
  }).catch(() => {});
  assert(await page.locator('.snip-stage').count() === 1, 'annotate mounted');

  const capture = await page.evaluate(() => {
    const api = !!(navigator.mediaDevices && navigator.mediaDevices.getDisplayMedia);
    const btn = Array.from(document.querySelectorAll('#template-tools button'))
      .find((b) => /Capture screen/i.test(b.textContent || ''));
    return {
      api,
      isSecureContext: window.isSecureContext,
      protocol: location.protocol,
      buttonPresent: !!btn,
      buttonDisabled: btn ? !!btn.disabled : null
    };
  });
  pass('getDisplayMedia probe', JSON.stringify(capture));

  // About
  await page.locator('#btn-about').click();
  await page.locator('.modal').waitFor({ state: 'visible' });
  const aboutText = await page.locator('.modal').innerText();
  assert(/1\.0\.0/.test(aboutText), 'about shows version');
  await page.locator('#about-close').click();

  // Multi-board count
  const boards = await page.locator('#board-select option').count();
  assert(boards >= 2, 'multiple boards', `count=${boards}`);

  // Network
  const external = [];
  page.on('request', (req) => {
    const u = req.url();
    if (!u.startsWith('file:') && !u.startsWith('data:') && !u.startsWith('blob:')) external.push(u);
  });
  await page.reload({ waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(400);
  assert(external.length === 0, 'no external requests after reload', external.join(', ') || 'none');

  const serious = consoleErrors.filter((e) => !/favicon/i.test(e));
  // Firefox sometimes logs non-fatal noise on file://
  const filtered = serious.filter((e) => !/Download the React/i.test(e));
  assert(filtered.length === 0, 'no console errors', filtered.join(' | ') || 'none');

  await context.close();
  const failed = results.filter((r) => !r.ok).length;
  return { engineId, results, failed, capture };
}

async function main() {
  const engines = engineConfigs();
  const report = {
    tier: 3,
    at: new Date().toISOString(),
    engines: [],
    targets: []
  };

  console.log('Whiteboard Tier-3 cross-engine verification\n');

  for (const target of targets) {
    if (!fs.existsSync(target.htmlPath)) {
      console.log(`SKIP target ${target.name}: missing ${target.htmlPath}`);
      continue;
    }
    console.log(`=== Target: ${target.name} ===`);
    console.log(target.htmlPath);

    for (const eng of engines) {
      process.stdout.write(`  ${eng.label} ... `);
      let browser;
      try {
        browser = await eng.launch();
      } catch (e) {
        const msg = e && e.message ? e.message.split('\n')[0] : String(e);
        console.log(`SKIP (${msg.slice(0, 120)})`);
        report.engines.push({
          target: target.name,
          engine: eng.id,
          label: eng.label,
          status: 'skipped',
          reason: msg
        });
        continue;
      }
      try {
        const out = await runSuite(browser, target.htmlPath, eng.id);
        const status = out.failed === 0 ? 'pass' : 'fail';
        console.log(`${status.toUpperCase()} (${out.results.length - out.failed}/${out.results.length})`);
        if (out.failed) {
          out.results.filter((r) => !r.ok).forEach((r) => {
            console.log(`      FAIL ${r.name}: ${r.detail}`);
          });
        }
        report.engines.push({
          target: target.name,
          engine: eng.id,
          label: eng.label,
          status,
          failed: out.failed,
          total: out.results.length,
          capture: out.capture,
          results: out.results
        });
      } catch (e) {
        console.log('ERROR', e.message || e);
        report.engines.push({
          target: target.name,
          engine: eng.id,
          label: eng.label,
          status: 'error',
          reason: String(e.message || e)
        });
      } finally {
        try { await browser.close(); } catch (_) {}
      }
    }
    console.log('');
  }

  // Compare working vs release file hashes (content equality)
  const work = path.resolve(__dirname, 'whiteboard.html');
  const rel = path.resolve(__dirname, 'release', 'whiteboard.html');
  if (fs.existsSync(work) && fs.existsSync(rel)) {
    const same = fs.readFileSync(work).equals(fs.readFileSync(rel));
    report.workingMatchesRelease = same;
    console.log('working copy === release copy:', same ? 'YES' : 'NO');
  }

  const outPath = path.join(__dirname, 'verify-report-tier3.json');
  fs.writeFileSync(outPath, JSON.stringify(report, null, 2));
  console.log('\nReport:', outPath);

  const hardFails = report.engines.filter((e) => e.status === 'fail' || e.status === 'error');
  // Skip is OK for unavailable engines; fail on actual fails for chromium at least
  const chromeFail = report.engines.some(
    (e) => e.engine === 'chromium' && (e.status === 'fail' || e.status === 'error')
  );
  if (chromeFail || hardFails.some((e) => e.engine === 'chromium')) {
    process.exit(1);
  }
  // If firefox/webkit ran and failed, also fail the job
  if (hardFails.length) process.exit(1);
}

main().catch((e) => {
  console.error(e);
  process.exit(2);
});
