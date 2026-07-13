/**
 * Whiteboard v2 Tier 1 verification — real Chromium via playwright-core.
 * Features: quota meter + auto-compression, corrupt recovery, multi-tab,
 * global search, session restore.
 */
import { chromium } from 'playwright-core';
import path from 'path';
import fs from 'fs';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const htmlPath = path.resolve(__dirname, 'whiteboard.html');
const fileUrl = 'file://' + htmlPath;
const CHROME =
  process.env.CHROME_PATH ||
  path.join(process.env.HOME || '', '.cache/ms-playwright/chromium-1228/chrome-linux64/chrome');

const results = [];
let failed = 0;

function pass(name, detail = '') {
  results.push({ ok: true, name, detail });
  console.log(`  PASS  ${name}${detail ? ' — ' + detail : ''}`);
}
function fail(name, detail = '') {
  failed++;
  results.push({ ok: false, name, detail });
  console.log(`  FAIL  ${name}${detail ? ' — ' + detail : ''}`);
}
function assert(cond, name, detail = '') {
  if (cond) pass(name, detail);
  else fail(name, detail);
}
function beforeAfter(a, b) {
  return a + ' → ' + b + ' (' + Math.round((1 - b / a) * 100) + '% smaller)';
}

async function waitBoard(page, ms = 200) {
  await page.waitForTimeout(ms);
  await page
    .waitForFunction(() => {
      const root = document.getElementById('board-root');
      return root && !root.classList.contains('switching') && window.__whiteboard;
    }, { timeout: 8000 })
    .catch(() => {});
  await page.waitForTimeout(80);
}

async function freshPage(browser, viewport = { width: 1280, height: 860 }) {
  const context = await browser.newContext({ viewport });
  const page = await context.newPage();
  const consoleErrors = [];
  page.on('console', (msg) => {
    if (msg.type() === 'error') consoleErrors.push(msg.text());
  });
  page.on('pageerror', (err) => consoleErrors.push(String(err)));
  page._consoleErrors = consoleErrors;
  page._context = context;
  await page.goto(fileUrl, { waitUntil: 'domcontentloaded', timeout: 30000 });
  await page.evaluate(() => localStorage.clear());
  await page.reload({ waitUntil: 'domcontentloaded' });
  await waitBoard(page, 350);
  return page;
}

async function pageWithInitStorage(browser, storageSetupFn, viewport = { width: 1280, height: 860 }) {
  const context = await browser.newContext({ viewport });
  await context.addInitScript(storageSetupFn);
  const page = await context.newPage();
  const consoleErrors = [];
  page.on('console', (msg) => {
    if (msg.type() === 'error') consoleErrors.push(msg.text());
  });
  page.on('pageerror', (err) => consoleErrors.push(String(err)));
  page._consoleErrors = consoleErrors;
  page._context = context;
  await page.goto(fileUrl, { waitUntil: 'domcontentloaded', timeout: 30000 });
  await waitBoard(page, 400);
  return page;
}

async function main() {
  console.log('Whiteboard v2 Tier-1 verification');
  console.log('File:', fileUrl);
  console.log('Chrome:', CHROME);
  console.log('');

  if (!fs.existsSync(CHROME)) {
    console.error('Chrome binary not found:', CHROME);
    process.exit(2);
  }

  const browser = await chromium.launch({
    executablePath: CHROME,
    headless: true,
    args: ['--allow-file-access-from-files', '--disable-web-security']
  });

  try {
    // ============================================================
    // A. Shell / chrome for Tier 1
    // ============================================================
    console.log('--- Shell ---');
    let page = await freshPage(browser);

    assert(await page.title() === 'Whiteboard', 'Page title');
    const ver = await page.evaluate(() => window.__whiteboard.VERSION);
    assert(ver === '2.0.0', 'App version 2.0.0', ver);
    assert(await page.locator('#global-search').count() === 1, 'Global search input present');
    assert(await page.locator('#quota-meter').count() === 1, 'Quota meter present');
    assert(await page.locator('#quota-label').count() === 1, 'Quota label present');
    assert(await page.locator('#btn-zoom-in').count() === 1, 'Zoom in control');
    assert(await page.locator('#btn-zoom-out').count() === 1, 'Zoom out control');
    assert(await page.locator('#tab-banner').count() === 1, 'Multi-tab banner present');
    assert(await page.locator('#recovery-banner').count() === 1, 'Recovery banner present');

    // ============================================================
    // B. Quota meter updates after save
    // ============================================================
    console.log('--- Quota meter ---');
    await page.evaluate(() => window.__whiteboard.App.flushSave({ forceOverwrite: true, sync: true }));
    await page.waitForTimeout(200);
    const quotaText = await page.locator('#quota-label').textContent();
    assert(/^\d+%$/.test((quotaText || '').trim()), 'Quota shows percentage', quotaText);
    const usedBytes = await page.evaluate(() => window.__whiteboard.estimateLocalStorageBytes());
    assert(usedBytes > 0, 'localStorage has data after save', String(usedBytes));

    // ============================================================
    // C. Auto-compression of large paint PNG
    // ============================================================
    console.log('--- Auto-compression ---');
    const compressResult = await page.evaluate(async () => {
      const wb = window.__whiteboard;
      const c = document.createElement('canvas');
      c.width = 1200;
      c.height = 900;
      const g = c.getContext('2d');
      const imgData = g.createImageData(c.width, c.height);
      for (let i = 0; i < imgData.data.length; i++) {
        imgData.data[i] = (Math.random() * 256) | 0;
      }
      g.putImageData(imgData, 0, 0);
      const bigUrl = c.toDataURL('image/png');
      const beforeLen = bigUrl.length;

      const board = wb.App.currentBoard();
      board.templateId = 'paint';
      board.state = { color: '#1a1a1a', sizeId: 'm', imageDataUrl: bigUrl };
      const payload = {
        version: 2,
        theme: 'light',
        currentBoardId: board.id,
        boards: wb.App.store.boards,
        session: wb.App.store.session || { zoom: 1, scrolls: {}, lastTool: {} },
        savedAt: new Date().toISOString()
      };
      const compressed = await wb.compressStorePayload(payload);
      const afterUrl = compressed.boards.find((b) => b.id === board.id).state.imageDataUrl;
      return {
        beforeLen,
        afterLen: afterUrl ? afterUrl.length : 0,
        compressed: afterUrl && afterUrl.length < beforeLen,
        edge: compressed.boards.find((b) => b.id === board.id).state._compressedEdge || null
      };
    });
    assert(compressResult.beforeLen > 200000, 'Synthetic paint PNG is large', String(compressResult.beforeLen));
    assert(compressResult.compressed, 'Auto-compression reduced PNG size',
      beforeAfter(compressResult.beforeLen, compressResult.afterLen) +
      (compressResult.edge ? ' edge=' + compressResult.edge : ''));

    const flushCompress = await page.evaluate(async () => {
      const wb = window.__whiteboard;
      const c = document.createElement('canvas');
      c.width = 1000;
      c.height = 800;
      const g = c.getContext('2d');
      const imgData = g.createImageData(c.width, c.height);
      for (let i = 0; i < imgData.data.length; i += 4) {
        imgData.data[i] = (Math.random() * 256) | 0;
        imgData.data[i + 1] = (Math.random() * 256) | 0;
        imgData.data[i + 2] = (Math.random() * 256) | 0;
        imgData.data[i + 3] = 255;
      }
      g.putImageData(imgData, 0, 0);
      const url = c.toDataURL('image/png');
      const board = wb.App.currentBoard();
      board.state = { color: '#1a1a1a', sizeId: 'm', imageDataUrl: url };
      const before = url.length;
      // compressStorePayload is the mechanism flushSave uses past soft threshold
      const payload = {
        version: 2,
        theme: 'light',
        currentBoardId: board.id,
        boards: JSON.parse(JSON.stringify(wb.App.store.boards)),
        session: wb.App.store.session || { zoom: 1, scrolls: {}, lastTool: {} },
        savedAt: new Date().toISOString()
      };
      payload.boards.find((b) => b.id === board.id).state = board.state;
      const compressed = await wb.compressStorePayload(payload);
      const afterState = compressed.boards.find((b) => b.id === board.id).state;
      // Write compressed via normal save path
      wb.App.store.boards = compressed.boards;
      const result = await wb.App.flushSave({ forceOverwrite: true, sync: true });
      return {
        before,
        after: (afterState && afterState.imageDataUrl && afterState.imageDataUrl.length) || 0,
        saved: !!localStorage.getItem(wb.STORAGE_KEY),
        writeOk: !!(result && result.ok)
      };
    });
    assert(flushCompress.saved && flushCompress.writeOk, 'flushSave wrote storage after large paint');
    assert(flushCompress.after < flushCompress.before, 'Compressed image smaller than original',
      beforeAfter(flushCompress.before, flushCompress.after));

    await page._context.close();

    // ============================================================
    // D. Corrupt-state recovery
    // ============================================================
    console.log('--- Corrupt recovery ---');
    // Init script sets corrupt data BEFORE app scripts run (avoids beforeunload overwrite)
    page = await pageWithInitStorage(browser, () => {
      try {
        localStorage.setItem('whiteboard-app-v2', '{not valid json!!!');
      } catch (_) {}
    });
    const recovery1 = await page.evaluate(() => {
      const banner = document.getElementById('recovery-banner');
      const q = localStorage.getItem(window.__whiteboard.STORAGE_KEY_QUARANTINE);
      return {
        bannerVisible: banner && banner.classList.contains('visible'),
        hasQuarantine: !!q,
        boards: window.__whiteboard.App.store.boards.length,
        version: window.__whiteboard.App.store.version,
        recoveryReason: window.__whiteboard.App._recovery && window.__whiteboard.App._recovery.reason
      };
    });
    assert(recovery1.bannerVisible, 'Recovery banner shown on bad JSON', JSON.stringify(recovery1));
    assert(recovery1.hasQuarantine, 'Corrupt data quarantined in localStorage');
    assert(recovery1.boards >= 1 && recovery1.version === 2, 'Fell back to clean default store',
      JSON.stringify({ boards: recovery1.boards, version: recovery1.version }));

    const dlPromise = page.waitForEvent('download', { timeout: 5000 }).catch(() => null);
    await page.locator('#btn-export-quarantine').click({ force: true });
    const dl = await dlPromise;
    assert(!!dl, 'Quarantine download triggered', dl ? dl.suggestedFilename() : 'no download');
    await page._context.close();

    // Schema-invalid but parseable JSON
    page = await pageWithInitStorage(browser, () => {
      try {
        localStorage.setItem('whiteboard-app-v2', JSON.stringify({
          version: 2,
          boards: [{ id: 123, name: null, templateId: 'paint' }]
        }));
      } catch (_) {}
    });
    const recovery2 = await page.evaluate(() => ({
      banner: document.getElementById('recovery-banner').classList.contains('visible'),
      reason: (JSON.parse(localStorage.getItem(window.__whiteboard.STORAGE_KEY_QUARANTINE) || '{}')).reason || '',
      appOk: !!window.__whiteboard.App.currentBoard()
    }));
    assert(recovery2.banner, 'Recovery banner on schema failure');
    assert(recovery2.appOk, 'App still opens after schema failure');
    assert(recovery2.reason.length > 0, 'Quarantine records reason', recovery2.reason);
    await page._context.close();

    // ============================================================
    // E. Multi-tab safety
    // ============================================================
    console.log('--- Multi-tab safety ---');
    page = await freshPage(browser);
    await page.evaluate(() => window.__whiteboard.App.flushSave({ forceOverwrite: true, sync: true }));
    await page.waitForTimeout(150);

    const multi = await page.evaluate(() => {
      const key = window.__whiteboard.STORAGE_KEY;
      const remote = {
        version: 2,
        theme: 'dark',
        currentBoardId: 'board_remote_1',
        boards: [{
          id: 'board_remote_1',
          name: 'Remote tab board',
          templateId: 'wordpad',
          state: { html: '<p>from other tab</p>' },
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString()
        }],
        session: { zoom: 1, scrolls: {}, lastTool: {} },
        savedAt: new Date().toISOString(),
        writerTabId: 'tab_other'
      };
      const newValue = JSON.stringify(remote);
      localStorage.setItem(key, newValue);
      window.dispatchEvent(new StorageEvent('storage', {
        key,
        newValue,
        oldValue: null,
        storageArea: localStorage
      }));
      return {
        banner: document.getElementById('tab-banner').classList.contains('visible'),
        pending: !!window.__whiteboard.App._remotePending
      };
    });
    assert(multi.banner, 'Tab conflict banner visible after remote write');
    assert(multi.pending, 'Remote pending flag set');

    const paused = await page.evaluate(async () => window.__whiteboard.App.flushSave());
    assert(paused && paused.ok === false && paused.reason === 'remote-pending',
      'Save paused while remote pending', JSON.stringify(paused));

    await page.locator('#btn-tab-reload').click();
    await waitBoard(page, 400);
    const synced = await page.evaluate(() => {
      const b = window.__whiteboard.App.currentBoard();
      return {
        name: b && b.name,
        templateId: b && b.templateId,
        theme: window.__whiteboard.App.store.theme,
        banner: document.getElementById('tab-banner').classList.contains('visible')
      };
    });
    assert(synced.name === 'Remote tab board', 'Synced board name from other tab', synced.name);
    assert(synced.templateId === 'wordpad', 'Synced template from other tab');
    assert(synced.theme === 'dark', 'Synced theme from other tab');
    assert(!synced.banner, 'Banner hidden after reload');

    // Overwrite path: remote write then keep-editing
    await page.evaluate(() => {
      const key = window.__whiteboard.STORAGE_KEY;
      const remote = JSON.parse(localStorage.getItem(key));
      remote.boards[0].name = 'Should be overwritten';
      const newValue = JSON.stringify(remote);
      localStorage.setItem(key, newValue);
      window.dispatchEvent(new StorageEvent('storage', {
        key, newValue, oldValue: null, storageArea: localStorage
      }));
    });
    await page.waitForTimeout(100);
    assert(
      await page.evaluate(() => document.getElementById('tab-banner').classList.contains('visible')),
      'Banner reappears on second remote write'
    );
    await page.locator('#btn-tab-dismiss').click();
    await page.waitForTimeout(400);
    const afterDismiss = await page.evaluate(() => {
      const b = window.__whiteboard.App.currentBoard();
      return {
        name: b && b.name,
        banner: document.getElementById('tab-banner').classList.contains('visible')
      };
    });
    assert(!afterDismiss.banner, 'Banner dismissed on keep-editing');
    assert(afterDismiss.name === 'Remote tab board', 'Local state kept after dismiss+overwrite', afterDismiss.name);
    await page._context.close();

    // ============================================================
    // F. Global search
    // ============================================================
    console.log('--- Global search ---');
    page = await freshPage(browser);

    await page.evaluate(async () => {
      const App = window.__whiteboard.App;
      App.currentBoard().name = 'Canvas Alpha';
      App.store.boards.push({
        id: 'board_notes_1',
        name: 'Ideas board',
        templateId: 'notepad',
        state: {
          notes: [
            { id: 'n1_abc', x: 40, y: 40, w: 200, h: 180, text: 'UniqueZebra phrase here', color: 'yellow' },
            { id: 'n2_def', x: 260, y: 40, w: 200, h: 180, text: 'ordinary note', color: 'blue' }
          ]
        },
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      });
      App.store.boards.push({
        id: 'board_wp_1',
        name: 'Journal',
        templateId: 'wordpad',
        state: { html: '<p>Meeting notes about UniqueZebra migration plan</p>' },
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      });
      App.store.boards.push({
        id: 'board_snip_1',
        name: 'Screenshot',
        templateId: 'snip',
        state: {
          tool: 'text',
          strokeColor: '#e03131',
          imageDataUrl: null,
          textLayers: [{ text: 'UniqueZebra callout', x: 10, y: 20, color: '#e03131' }]
        },
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      });
      await App.flushSave({ forceOverwrite: true, sync: true });
      App.refreshBoardSelect();
    });

    const searchHits = await page.evaluate(() =>
      window.__whiteboard.App.searchAll('uniquezebra').map((r) => ({
        kind: r.kind, title: r.title, preview: (r.preview || '').slice(0, 40)
      }))
    );
    const kinds = searchHits.map((h) => h.kind).sort();
    assert(searchHits.length >= 3, 'Search finds multiple matches', JSON.stringify(searchHits));
    assert(kinds.includes('note'), 'Search hits sticky note');
    assert(kinds.includes('wordpad'), 'Search hits wordpad text');
    assert(kinds.includes('annotate-text'), 'Search hits annotate text layers');

    const nameHits = await page.evaluate(() => window.__whiteboard.App.searchAll('canvas alpha'));
    assert(nameHits.some((h) => h.kind === 'board-name'), 'Search hits board name');

    await page.locator('#global-search').fill('UniqueZebra');
    await page.waitForTimeout(200);
    const uiResults = await page.locator('.search-result').count();
    assert(uiResults >= 3, 'Search dropdown shows results', 'count=' + uiResults);

    const jumped = await page.evaluate(() => {
      const App = window.__whiteboard.App;
      const hits = App.searchAll('uniquezebra');
      const note = hits.find((h) => h.kind === 'note');
      if (!note) return { ok: false, reason: 'no note hit' };
      App.jumpToSearchResult(note);
      return { ok: true, boardId: note.boardId, noteId: note.noteId };
    });
    await waitBoard(page, 400);
    await page.waitForTimeout(300);
    const afterJump = await page.evaluate(() => {
      const b = window.__whiteboard.App.currentBoard();
      return {
        boardName: b && b.name,
        templateId: b && b.templateId,
        hasNoteEl: !!document.querySelector('.sticky-note[data-id="n1_abc"]')
      };
    });
    assert(jumped.ok, 'jumpToSearchResult invoked');
    assert(afterJump.templateId === 'notepad' && afterJump.boardName === 'Ideas board',
      'Jumped to notes board', JSON.stringify(afterJump));
    assert(afterJump.hasNoteEl, 'Target sticky note element present');
    await page._context.close();

    // ============================================================
    // G. Session restore (board, scroll, zoom, tool)
    // ============================================================
    console.log('--- Session restore ---');
    page = await freshPage(browser);

    await page.evaluate(async () => {
      const App = window.__whiteboard.App;
      const paint = App.currentBoard();
      paint.name = 'Session Paint';
      // Apply tools through the live instance so snapshot/getState persists them
      App.instance.setState({ color: '#e03131', sizeId: 'l' });
      App.snapshotCurrent();
      App.store.session.zoom = 1.3;
      App.applyZoom(1.3, true);
      App.store.session.lastTool = { paint: { color: '#e03131', sizeId: 'l' } };

      App.store.boards.push({
        id: 'board_scroll_1',
        name: 'Scrolly Notes',
        templateId: 'notepad',
        state: {
          notes: [
            { id: 'n_top', x: 20, y: 20, w: 180, h: 140, text: 'top', color: 'yellow' },
            { id: 'n_far', x: 40, y: 1600, w: 180, h: 140, text: 'far down note', color: 'pink' }
          ]
        },
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      });
      await App.flushSave({ forceOverwrite: true, sync: true });
    });

    await page.evaluate(() =>
      window.__whiteboard.App.openBoard('board_scroll_1', { force: true, restoreSession: true })
    );
    await waitBoard(page, 400);
    await page.evaluate(async () => {
      const App = window.__whiteboard.App;
      const scroller = document.querySelector('.notepad-board');
      if (scroller) {
        scroller.scrollTop = 1200;
        scroller.scrollLeft = 30;
      }
      App.captureScroll();
      App.store.session.zoom = 1.3;
      App.applyZoom(1.3, true);
      App.store.currentBoardId = 'board_scroll_1';
      await App.flushSave({ forceOverwrite: true, sync: true });
    });

    await page.reload({ waitUntil: 'domcontentloaded' });
    await waitBoard(page, 500);
    await page.waitForTimeout(450);

    const restored = await page.evaluate(() => {
      const App = window.__whiteboard.App;
      const scroller = document.querySelector('.notepad-board');
      return {
        boardId: App.currentBoardId,
        boardName: App.currentBoard() && App.currentBoard().name,
        zoom: App.store.session && App.store.session.zoom,
        zoomLabel: document.getElementById('zoom-label').textContent,
        scrollTop: scroller ? scroller.scrollTop : null,
        scrollLeft: scroller ? scroller.scrollLeft : null
      };
    });
    assert(restored.boardId === 'board_scroll_1', 'Restored last board id', restored.boardId);
    assert(restored.boardName === 'Scrolly Notes', 'Restored board name', restored.boardName);
    assert(Math.abs((restored.zoom || 0) - 1.3) < 0.05, 'Restored zoom level', String(restored.zoom));
    assert(restored.zoomLabel === '130%', 'Zoom label matches', restored.zoomLabel);
    assert(restored.scrollTop !== null && restored.scrollTop >= 1000,
      'Restored scroll position', 'scrollTop=' + restored.scrollTop);

    await page.evaluate(() => {
      const App = window.__whiteboard.App;
      App.openBoard(App.store.boards.find((b) => b.templateId === 'paint').id, {
        force: true,
        restoreSession: true
      });
    });
    await waitBoard(page, 400);
    const toolState = await page.evaluate(() => {
      const st = window.__whiteboard.App.instance.getState();
      return { color: st.color, sizeId: st.sizeId };
    });
    assert(toolState.color === '#e03131', 'Restored paint color tool', toolState.color);
    assert(toolState.sizeId === 'l', 'Restored paint brush size', toolState.sizeId);

    await page.locator('#btn-zoom-in').click();
    await page.waitForTimeout(100);
    const zoomAfter = await page.evaluate(() => window.__whiteboard.App.store.session.zoom);
    assert(zoomAfter > 1.3, 'Zoom in control increases zoom', String(zoomAfter));

    // ============================================================
    // H. Validation helpers + console
    // ============================================================
    console.log('--- Validation / console ---');
    const validation = await page.evaluate(() => {
      const v = window.__whiteboard.validateStore;
      return {
        good: v({ version: 2, boards: [{ id: 'a', name: 'n', templateId: 'paint', state: {} }] }).ok,
        badVer: v({ version: 1, boards: [{ id: 'a', name: 'n', templateId: 'paint', state: {} }] }).ok,
        badBoard: v({ version: 2, boards: [{ id: 1, name: 'n', templateId: 'paint' }] }).ok
      };
    });
    assert(validation.good === true, 'validateStore accepts good data');
    assert(validation.badVer === false, 'validateStore rejects bad version');
    assert(validation.badBoard === false, 'validateStore rejects bad board id type');

    const errs = (page._consoleErrors || []).filter((e) => !/ResizeObserver loop/.test(e));
    assert(errs.length === 0, 'No page console errors', errs.slice(0, 3).join(' | '));

    await page._context.close();
  } finally {
    await browser.close();
  }

  const report = {
    when: new Date().toISOString(),
    version: '2.0.0',
    tier: 'v2-tier1',
    file: htmlPath,
    chrome: CHROME,
    passed: results.filter((r) => r.ok).length,
    failed,
    total: results.length,
    results
  };
  const out = path.resolve(__dirname, 'verify-report-v2-tier1.json');
  fs.writeFileSync(out, JSON.stringify(report, null, 2));
  console.log('');
  console.log(failed ? `FAILED ${failed}/${results.length}` : `ALL PASSED ${results.length}/${results.length}`);
  console.log('Report:', out);
  process.exit(failed ? 1 : 0);
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
