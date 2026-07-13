/**
 * Whiteboard v2.4.0 packaging verification — working + release copies
 * across Chromium, Firefox, and WebKit (when host deps allow).
 *
 * Designed to catch engine-specific regressions (cf. v1 WebKit unmount race,
 * v2 Firefox Kanban drag). Includes real pointer-drag on Kanban, not state-only hacks.
 */
import { chromium, firefox, webkit } from 'playwright-core';
import path from 'path';
import fs from 'fs';
import { fileURLToPath } from 'url';
import crypto from 'crypto';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const CHROME =
  process.env.CHROME_PATH ||
  path.join(process.env.HOME || '', '.cache/ms-playwright/chromium-1228/chrome-linux64/chrome');

const targets = [
  { name: 'working-copy', htmlPath: path.resolve(__dirname, 'whiteboard.html') },
  { name: 'release-copy', htmlPath: path.resolve(__dirname, 'release', 'whiteboard.html') }
];

function fileUrl(p) {
  return 'file://' + p;
}

function sha256(p) {
  return crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex');
}

function engineConfigs() {
  const list = [];
  list.push({
    id: 'chromium',
    label: 'Chromium',
    launch: () =>
      chromium.launch({
        executablePath: fs.existsSync(CHROME) ? CHROME : undefined,
        headless: true,
        args: ['--allow-file-access-from-files', '--disable-web-security']
      })
  });
  list.push({
    id: 'firefox',
    label: 'Firefox',
    launch: () => firefox.launch({ headless: true })
  });
  list.push({
    id: 'webkit',
    label: 'WebKit',
    launch: () => webkit.launch({ headless: true })
  });
  return list;
}

async function waitBoard(page) {
  await page.waitForTimeout(200);
  await page
    .waitForFunction(() => {
      const r = document.getElementById('board-root');
      return r && !r.classList.contains('switching') && window.__whiteboard;
    }, { timeout: 10000 })
    .catch(() => {});
  await page.waitForTimeout(100);
}

async function runSuite(browser, htmlPath, engineId) {
  const results = [];
  const pass = (name, detail = '') => results.push({ ok: true, name, detail });
  const fail = (name, detail = '') => results.push({ ok: false, name, detail });
  const assert = (cond, name, detail = '') => (cond ? pass : fail)(name, detail);

  const context = await browser.newContext({ viewport: { width: 1280, height: 860 } });
  const page = await context.newPage();
  const consoleErrors = [];
  const requests = [];
  page.on('console', (m) => {
    if (m.type() === 'error') consoleErrors.push(m.text());
  });
  page.on('pageerror', (e) => consoleErrors.push(String(e)));
  page.on('request', (req) => {
    const u = req.url();
    if (!u.startsWith('file:') && !u.startsWith('data:') && !u.startsWith('blob:')) {
      requests.push(u);
    }
  });

  await page.goto(fileUrl(htmlPath), { waitUntil: 'domcontentloaded', timeout: 45000 });
  await page.evaluate(() => localStorage.clear());
  await page.reload({ waitUntil: 'domcontentloaded' });
  await waitBoard(page);

  // --- Identity / packaging ---
  assert((await page.title()) === 'Whiteboard', 'title');
  const ver = await page.evaluate(() => window.__whiteboard && window.__whiteboard.VERSION);
  assert(ver === '2.4.0', 'VERSION constant 2.4.0', String(ver));
  const badge = await page.locator('#version-badge').textContent();
  assert(/v?2\.4\.0/.test((badge || '').trim()), 'version badge', badge);

  const favicon = await page.evaluate(() => {
    const link = document.querySelector('link[rel="icon"]');
    return link ? link.getAttribute('href') || '' : '';
  });
  assert(favicon.startsWith('data:image/svg+xml'), 'inline SVG favicon', favicon.slice(0, 32));

  const templates = await page.evaluate(() =>
    window.__whiteboard.TemplateRegistry.list().map((t) => t.id).sort()
  );
  assert(
    templates.join(',') === 'kanban,notepad,paint,snip,wordpad',
    'five templates incl. kanban',
    templates.join(',')
  );

  // --- Chrome for v2 features ---
  assert((await page.locator('#global-search').count()) === 1, 'global search present');
  assert((await page.locator('#btn-cmd-palette').count()) === 1, 'command palette button');
  assert((await page.locator('#quota-meter').count()) === 1, 'quota meter');
  assert((await page.locator('[data-action="history"]').count()) === 1, 'time machine menu item');
  assert((await page.locator('[data-action="trash"]').count()) === 1, 'trash menu item');

  // About panel content (v2 feature set)
  await page.locator('#btn-about').click();
  await page.waitForTimeout(150);
  const aboutText = await page.evaluate(() => {
    const m = document.querySelector('.modal .about-body');
    return m ? m.innerText : '';
  });
  assert(/Kanban/i.test(aboutText), 'About mentions Kanban');
  assert(/Time machine|snapshot/i.test(aboutText), 'About mentions time machine');
  assert(/Trash/i.test(aboutText), 'About mentions trash');
  assert(/Ctrl\+K|content search/i.test(aboutText), 'About mentions search');
  assert(/Ctrl\+\/|command palette/i.test(aboutText), 'About mentions palette');
  assert(/Smart paste|paste/i.test(aboutText), 'About mentions smart paste');
  await page.locator('#about-close').click().catch(() => {});
  await page.waitForTimeout(100);

  // --- Paint stroke ---
  const canvas = page.locator('.paint-canvas-wrap canvas');
  await canvas.waitFor({ state: 'visible', timeout: 8000 }).catch(() => {});
  const box = await canvas.boundingBox();
  assert(!!box && box.width > 40, 'paint canvas visible', JSON.stringify(box));
  if (box) {
    await page.mouse.move(box.x + 40, box.y + 40);
    await page.mouse.down();
    await page.mouse.move(box.x + 120, box.y + 90, { steps: 10 });
    await page.mouse.up();
    await page.waitForTimeout(150);
    const ink = await page.evaluate(() => {
      const c = document.querySelector('.paint-canvas-wrap canvas');
      if (!c) return 0;
      const d = c.getContext('2d').getImageData(0, 0, Math.min(c.width, 400), Math.min(c.height, 300)).data;
      let n = 0;
      for (let i = 0; i < d.length; i += 4) {
        if (d[i] < 250 || d[i + 1] < 250 || d[i + 2] < 250) n++;
      }
      return n;
    });
    assert(ink > 10, 'paint stroke leaves ink', `pixels=${ink}`);
  }

  // --- Create Kanban + real pointer drag (engine-sensitive) ---
  const kanbanDrag = await page.evaluate(async () => {
    const App = window.__whiteboard.App;
    const wb = window.__whiteboard;
    App.createBoard({
      name: 'Pkg Kanban',
      templateId: 'kanban',
      state: wb.defaultKanbanState(),
      open: true
    });
    await new Promise((r) => setTimeout(r, 300));
    const st = App.currentBoard().state;
    App.instance.addCard(st.columns[0].id, 'Drag me card');
    await new Promise((r) => setTimeout(r, 120));
    App.snapshotCurrent();
    const cardId = App.currentBoard().state.columns[0].cardIds[0];
    const doingId = App.currentBoard().state.columns[1].id;
    const grip = document.querySelector('.kanban-card[data-id="' + CSS.escape(cardId) + '"] .grip');
    const target = document.querySelector('.kanban-col[data-id="' + CSS.escape(doingId) + '"] .kanban-cards');
    if (!grip || !target) return { ok: false, reason: 'missing grip/target' };
    const g = grip.getBoundingClientRect();
    const t = target.getBoundingClientRect();
    const sx = g.left + g.width / 2;
    const sy = g.top + g.height / 2;
    const ex = t.left + t.width / 2;
    const ey = t.top + 30;
    function fire(type, x, y, el) {
      el.dispatchEvent(
        new PointerEvent(type, {
          bubbles: true,
          cancelable: true,
          clientX: x,
          clientY: y,
          pointerId: 1,
          pointerType: 'mouse',
          button: 0,
          buttons: type === 'pointerup' ? 0 : 1
        })
      );
    }
    fire('pointerdown', sx, sy, grip);
    for (let i = 1; i <= 8; i++) {
      fire('pointermove', sx + ((ex - sx) * i) / 8, sy + ((ey - sy) * i) / 8, window);
      await new Promise((r) => setTimeout(r, 12));
    }
    fire('pointerup', ex, ey, window);
    await new Promise((r) => setTimeout(r, 100));
    App.snapshotCurrent();
    const after = App.currentBoard().state;
    return {
      ok: true,
      inDoing: after.columns[1].cardIds.indexOf(cardId) >= 0,
      todoCount: after.columns[0].cardIds.length,
      doingCount: after.columns[1].cardIds.length,
      engine: navigator.userAgent
    };
  });
  assert(kanbanDrag.ok, 'Kanban drag gesture executed', JSON.stringify(kanbanDrag));
  assert(kanbanDrag.inDoing, 'Kanban pointer drag moved card to In Progress', JSON.stringify(kanbanDrag));

  // --- Time machine entry + trash soft-delete via UI ---
  await page.locator('#btn-board-menu').click();
  await page.locator('[data-action="history"]').click();
  await page.waitForTimeout(200);
  const histOpen = await page.locator('#history-list').count();
  assert(histOpen === 1, 'Time machine opens from board menu');
  await page.locator('#history-close').click().catch(() => {});
  await page.waitForTimeout(100);

  // Second board so we can trash one
  await page.evaluate(async () => {
    window.__whiteboard.App.createBoard({
      name: 'Spare Paint',
      templateId: 'paint',
      state: {},
      open: false,
      silent: true
    });
    await new Promise((r) => setTimeout(r, 100));
  });
  await page.locator('#btn-board-menu').click();
  await page.locator('[data-action="delete"]').click();
  await page.waitForTimeout(150);
  // Confirm trash modal
  await page.evaluate(() => {
    const btns = document.querySelectorAll('.modal-backdrop .modal-actions button.active');
    if (btns.length) btns[btns.length - 1].click();
  });
  await waitBoard(page);
  const trashState = await page.evaluate(() => ({
    trash: (window.__whiteboard.App.store.trash || []).length,
    live: window.__whiteboard.App.store.boards.length
  }));
  assert(trashState.trash >= 1, 'Delete moves board to trash', JSON.stringify(trashState));

  // --- Search / palette present and openable ---
  await page.keyboard.press(engineId === 'webkit' ? 'Meta+k' : 'Control+k').catch(() => {});
  // Prefer click focus for reliability across engines
  await page.locator('#global-search').click();
  await page.locator('#global-search').fill('Drag');
  await page.waitForTimeout(150);
  // May or may not have results depending on board state
  assert((await page.locator('#global-search').inputValue()) === 'Drag', 'search input works');

  await page.locator('#btn-cmd-palette').click();
  await page.waitForTimeout(150);
  assert((await page.locator('.cmd-backdrop').count()) === 1, 'command palette opens');
  await page.keyboard.press('Escape');
  await page.waitForTimeout(80);

  // --- Network / console ---
  assert(requests.length === 0, 'no external network requests', requests.slice(0, 3).join(' | '));
  const cleanErrors = consoleErrors.filter((e) => !/ResizeObserver loop/.test(e));
  assert(cleanErrors.length === 0, 'no console errors', cleanErrors.slice(0, 2).join(' | '));

  await context.close();
  return results;
}

async function main() {
  console.log('Whiteboard v2.4.0 packaging verification');
  console.log('Chrome path:', CHROME);
  console.log('');

  // Byte identity of packaged HTML
  const working = path.resolve(__dirname, 'whiteboard.html');
  const release = path.resolve(__dirname, 'release', 'whiteboard.html');
  const identity = {
    workingSha: sha256(working),
    releaseSha: sha256(release),
    identical: sha256(working) === sha256(release),
    workingBytes: fs.statSync(working).size,
    releaseBytes: fs.statSync(release).size
  };
  console.log('HTML identity:', identity.identical ? 'IDENTICAL' : 'DIFFER', identity);

  const report = {
    when: new Date().toISOString(),
    version: '2.4.0',
    tier: 'v2-packaging',
    identity,
    engines: {},
    summary: { passed: 0, failed: 0, skipped: 0 }
  };

  process.env.PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS =
    process.env.PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS || '1';

  for (const eng of engineConfigs()) {
    report.engines[eng.id] = {};
    for (const target of targets) {
      const key = target.name;
      console.log(`\n=== ${eng.label} · ${key} ===`);
      let browser;
      try {
        browser = await eng.launch();
      } catch (e) {
        console.log(`  SKIP launch failed: ${String(e.message).split('\n')[0]}`);
        report.engines[eng.id][key] = {
          status: 'skipped',
          reason: String(e.message).slice(0, 500)
        };
        report.summary.skipped++;
        continue;
      }
      try {
        const results = await runSuite(browser, target.htmlPath, eng.id);
        const failed = results.filter((r) => !r.ok).length;
        const passed = results.filter((r) => r.ok).length;
        results.forEach((r) => {
          console.log(`  ${r.ok ? 'PASS' : 'FAIL'}  ${r.name}${r.detail ? ' — ' + r.detail : ''}`);
        });
        console.log(`  → ${passed}/${results.length} passed`);
        report.engines[eng.id][key] = {
          status: failed ? 'failed' : 'passed',
          passed,
          failed,
          total: results.length,
          results
        };
        report.summary.passed += passed;
        report.summary.failed += failed;
      } catch (e) {
        console.log(`  FAIL suite error: ${e.message}`);
        report.engines[eng.id][key] = { status: 'error', reason: String(e.stack || e) };
        report.summary.failed++;
      } finally {
        await browser.close().catch(() => {});
      }
    }
  }

  const out = path.resolve(__dirname, 'verify-report-v2-package.json');
  fs.writeFileSync(out, JSON.stringify(report, null, 2));
  console.log('\nReport:', out);
  console.log(
    `Summary: ${report.summary.passed} pass assertions, ${report.summary.failed} fail, ${report.summary.skipped} engine skips`
  );

  // Fail CI if identity broken or any non-skipped engine failed
  let hardFail = !identity.identical;
  for (const eng of Object.values(report.engines)) {
    for (const t of Object.values(eng)) {
      if (t.status === 'failed' || t.status === 'error') hardFail = true;
    }
  }
  // Require at least chromium + firefox passed on both copies
  const need = ['chromium', 'firefox'];
  for (const id of need) {
    for (const t of targets) {
      const st = report.engines[id] && report.engines[id][t.name];
      if (!st || st.status !== 'passed') hardFail = true;
    }
  }
  process.exit(hardFail ? 1 : 0);
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
