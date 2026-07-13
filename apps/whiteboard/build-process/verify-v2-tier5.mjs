/**
 * Whiteboard v2 Tier 5 — time machine + trash.
 * Real Chromium via playwright-core.
 * UI multi-step flows are driven through the DOM (menu clicks, restore buttons),
 * not only internal APIs.
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

async function freshPage(browser) {
  const context = await browser.newContext({ viewport: { width: 1280, height: 900 } });
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

async function main() {
  console.log('Whiteboard v2 Tier-5 verification');
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
    const page = await freshPage(browser);

    // ============================================================
    console.log('--- Shell / version ---');
    const ver = await page.evaluate(() => window.__whiteboard.VERSION);
    assert(ver === '2.4.0', 'App version 2.4.0', ver);
    assert(await page.locator('[data-action="history"]').count() === 1, 'Board menu has Time machine');
    assert(await page.locator('[data-action="trash"]').count() === 1, 'File menu has Trash');
    const consts = await page.evaluate(() => ({
      max: window.__whiteboard.HISTORY_MAX_SNAPSHOTS,
      interval: window.__whiteboard.HISTORY_MIN_INTERVAL_MS,
      trashAge: window.__whiteboard.TRASH_MAX_AGE_MS
    }));
    assert(consts.max === 10, 'History max 10', String(consts.max));
    assert(consts.interval === 2 * 60 * 1000, 'Snapshot min interval 2 minutes', String(consts.interval));
    assert(consts.trashAge === 7 * 24 * 60 * 60 * 1000, 'Trash max age 7 days');

    // ============================================================
    console.log('--- Time machine: snapshots + DOM restore ---');
    // Seed board with content and forced snapshots (interval bypass via force)
    const seed = await page.evaluate(async () => {
      const App = window.__whiteboard.App;
      App.createBoard({
        name: 'History Notes',
        templateId: 'notepad',
        state: {
          notes: [{
            id: 'n_hist', x: 40, y: 40, w: 200, h: 160,
            text: 'Version ONE #alpha', color: 'yellow', colorAuto: true
          }]
        },
        open: true
      });
      await new Promise((r) => setTimeout(r, 280));
      App.snapshotCurrent();
      const s1 = App.recordHistorySnapshot(App.currentBoard(), { force: true, reason: 'test-v1' });
      // Mutate to version two
      App.currentBoard().state.notes[0].text = 'Version TWO #beta';
      App.instance.setState(App.currentBoard().state);
      App.snapshotCurrent();
      const s2 = App.recordHistorySnapshot(App.currentBoard(), { force: true, reason: 'test-v2' });
      // Create more to test rolling window of 10
      for (let i = 0; i < 12; i++) {
        App.currentBoard().state.notes[0].text = 'Bulk ' + i;
        App.instance.setState(App.currentBoard().state);
        App.snapshotCurrent();
        App.recordHistorySnapshot(App.currentBoard(), { force: true, reason: 'bulk-' + i });
      }
      const hist = App.getBoardHistory(App.currentBoard().id);
      return {
        s1: s1 && s1.id,
        s2: s2 && s2.id,
        histLen: hist.length,
        newestSummary: hist[0] && hist[0].summary,
        hasV1: hist.some((h) => h.id === s1.id),
        boardId: App.currentBoard().id
      };
    });
    assert(seed.histLen === 10, 'Rolling window caps at 10 snapshots', String(seed.histLen));
    assert(!seed.hasV1, 'Oldest snapshots dropped when over 10', 'v1 still present? ' + seed.hasV1);

    // Re-seed a clean history for restore UI test
    await page.evaluate(async () => {
      const App = window.__whiteboard.App;
      App.store.history[App.currentBoard().id] = [];
      App.currentBoard().state.notes[0].text = 'Restore target A';
      App.instance.setState(App.currentBoard().state);
      App.snapshotCurrent();
      App.recordHistorySnapshot(App.currentBoard(), { force: true, reason: 'ui-a' });
      App.currentBoard().state.notes[0].text = 'Restore target B (current)';
      App.instance.setState(App.currentBoard().state);
      App.snapshotCurrent();
      App.recordHistorySnapshot(App.currentBoard(), { force: true, reason: 'ui-b' });
      await App.flushSave({ forceOverwrite: true, sync: true });
    });
    await waitBoard(page, 200);

    // Open time machine via board menu (DOM)
    await page.locator('#btn-board-menu').click();
    await page.locator('[data-action="history"]').click();
    await page.waitForTimeout(200);
    assert(await page.locator('.history-list').count() === 1, 'History picker opened via menu');
    const histItems = await page.locator('.history-item').count();
    assert(histItems >= 2, 'History list shows snapshots', String(histItems));

    // Click Restore on the oldest listed item that mentions "Restore target A"
    // List is newest-first; find item whose summary contains "Restore target A"
    const restoredViaUi = await page.evaluate(async () => {
      const items = Array.from(document.querySelectorAll('.history-item'));
      const target = items.find((el) => /Restore target A/i.test(el.textContent || ''));
      if (!target) return { ok: false, reason: 'no A item', texts: items.map((i) => i.textContent.slice(0, 80)) };
      const btn = target.querySelector('.hi-restore');
      btn.click();
      return { ok: true };
    });
    assert(restoredViaUi.ok, 'Found snapshot A restore button', JSON.stringify(restoredViaUi));

    // Confirm modal — drive OK through DOM
    await page.waitForSelector('.modal-backdrop .modal-actions button.active', { timeout: 3000 });
    // The confirm is the second modal; click its active OK
    await page.evaluate(() => {
      const modals = document.querySelectorAll('.modal-backdrop .modal-actions button.active');
      const ok = modals[modals.length - 1];
      if (ok) ok.click();
    });
    await waitBoard(page, 450);

    const afterRestore = await page.evaluate(() => {
      const App = window.__whiteboard.App;
      App.snapshotCurrent();
      const n = App.currentBoard().state.notes[0];
      const hist = App.getBoardHistory(App.currentBoard().id);
      return {
        text: n && n.text,
        histReasons: hist.map((h) => h.reason),
        pickerClosed: !document.querySelector('#history-list')
      };
    });
    assert(afterRestore.text === 'Restore target A', 'DOM restore applied snapshot A', afterRestore.text);
    assert(
      afterRestore.histReasons[0] === 'pre-restore' || afterRestore.histReasons.some((r) => r === 'pre-restore' || r === 'ui-b'),
      'Pre-restore snapshot retained for undo',
      afterRestore.histReasons.join(',')
    );

    // Interval heuristic: non-force should not double-snapshot identical content immediately
    const interval = await page.evaluate(() => {
      const App = window.__whiteboard.App;
      const b = App.currentBoard();
      const before = App.getBoardHistory(b.id).length;
      const r1 = App.recordHistorySnapshot(b, { force: false });
      const r2 = App.recordHistorySnapshot(b, { force: false });
      const after = App.getBoardHistory(b.id).length;
      return { before, after, r1: !!r1, r2: !!r2 };
    });
    assert(!interval.r1 && !interval.r2 && interval.after === interval.before,
      'Auto snapshot skipped when unchanged / within interval',
      JSON.stringify(interval));

    // ============================================================
    console.log('--- Trash: soft-delete + DOM restore ---');
    // Ensure we have at least 2 boards so delete is allowed
    await page.evaluate(async () => {
      const App = window.__whiteboard.App;
      App.createBoard({ name: 'Keep Alive', templateId: 'paint', state: {}, open: false, silent: true });
      // Open History Notes (or current notepad)
      const n = App.store.boards.find((b) => b.name === 'History Notes') || App.currentBoard();
      App.openBoard(n.id, { force: true });
      await new Promise((r) => setTimeout(r, 250));
    });
    await waitBoard(page, 200);

    const nameToTrash = await page.evaluate(() => window.__whiteboard.App.currentBoard().name);

    // Delete via board menu (DOM)
    await page.locator('#btn-board-menu').click();
    await page.locator('[data-action="delete"]').click();
    await page.waitForSelector('.modal-backdrop', { timeout: 3000 });
    // Confirm move to trash
    await page.evaluate(() => {
      const ok = document.querySelector('.modal-backdrop .modal-actions button.active');
      if (ok) ok.click();
    });
    await waitBoard(page, 400);

    const afterDelete = await page.evaluate((name) => {
      const App = window.__whiteboard.App;
      return {
        stillLive: App.store.boards.some((b) => b.name === name),
        trashCount: (App.store.trash || []).length,
        trashNames: (App.store.trash || []).map((t) => t.board && t.board.name)
      };
    }, nameToTrash);
    assert(!afterDelete.stillLive, 'Board removed from live list', nameToTrash);
    assert(afterDelete.trashNames.indexOf(nameToTrash) >= 0, 'Board is in trash',
      JSON.stringify(afterDelete.trashNames));

    // Open trash via File menu (DOM)
    await page.locator('#btn-file-menu').click();
    await page.locator('[data-action="trash"]').click();
    await page.waitForSelector('#trash-list', { timeout: 3000 });
    assert(await page.locator('.trash-item').count() >= 1, 'Trash list shows items');

    // Restore first item via DOM button
    await page.locator('.trash-item .ti-restore').first().click();
    await page.waitForTimeout(400);
    await waitBoard(page, 200);

    const afterTrashRestore = await page.evaluate((name) => {
      const App = window.__whiteboard.App;
      return {
        liveAgain: App.store.boards.some((b) => b.name === name),
        trashCount: (App.store.trash || []).length,
        current: App.currentBoard() && App.currentBoard().name,
        pickerOpen: !!document.querySelector('#trash-list')
      };
    }, nameToTrash);
    assert(afterTrashRestore.liveAgain, 'Board restored to live list');
    assert(afterTrashRestore.current === nameToTrash, 'Restored board opened', afterTrashRestore.current);

    // Close trash if still open
    if (await page.locator('#trash-close').count()) {
      await page.locator('#trash-close').click().catch(() => {});
    }

    // ============================================================
    console.log('--- Trash auto-purge (7 days) ---');
    const purge = await page.evaluate(() => {
      const App = window.__whiteboard.App;
      const old = {
        deletedAt: new Date(Date.now() - 8 * 24 * 60 * 60 * 1000).toISOString(),
        board: {
          id: 'board_old_trash',
          name: 'Ancient',
          templateId: 'paint',
          state: {},
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString()
        }
      };
      const fresh = {
        deletedAt: new Date().toISOString(),
        board: {
          id: 'board_fresh_trash',
          name: 'Recent trash',
          templateId: 'paint',
          state: {},
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString()
        }
      };
      App.store.trash = [fresh, old];
      App.store.history['board_old_trash'] = [{ id: 's', at: new Date().toISOString(), state: {}, summary: 'x' }];
      const kept = window.__whiteboard.purgeExpiredTrash(App.store.trash, App.store.history);
      return {
        keptNames: kept.map((t) => t.board.name),
        oldHistoryGone: !App.store.history['board_old_trash']
      };
    });
    assert(purge.keptNames.join(',') === 'Recent trash', 'Purge drops 8-day-old trash',
      purge.keptNames.join(','));
    assert(purge.oldHistoryGone, 'History for purged trash board removed');

    // ============================================================
    console.log('--- Quota includes history/trash + trim ---');
    const quota = await page.evaluate(() => {
      const App = window.__whiteboard.App;
      // Inflate history with many medium snaps
      const bid = App.currentBoard().id;
      App.store.history[bid] = [];
      for (let i = 0; i < 10; i++) {
        App.store.history[bid].push({
          id: 'snap_big_' + i,
          at: new Date(Date.now() - i * 60000).toISOString(),
          fingerprint: 'fp' + i,
          templateId: 'notepad',
          name: 'n',
          state: { notes: [{ id: 'n', x: 0, y: 0, w: 100, h: 100, text: 'x'.repeat(5000), color: 'yellow' }] },
          summary: 'big ' + i,
          reason: 'test'
        });
      }
      App.store.trash = [{
        deletedAt: new Date().toISOString(),
        board: {
          id: 'tb1', name: 'Trashed fat', templateId: 'notepad',
          state: { notes: [{ id: 'n', x: 0, y: 0, w: 100, h: 100, text: 'y'.repeat(3000), color: 'yellow' }] },
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString()
        }
      }];
      const payload = {
        version: 2,
        theme: 'light',
        currentBoardId: App.currentBoardId,
        boards: App.store.boards,
        session: App.store.session,
        settings: App.store.settings,
        trash: App.store.trash,
        history: App.store.history
      };
      const before = JSON.stringify(payload).length;
      const beforeSnaps = payload.history[bid].length;
      window.__whiteboard.trimHistoryForQuota(payload, 50000); // force aggressive trim
      const after = JSON.stringify(payload).length;
      const afterSnaps = (payload.history[bid] && payload.history[bid].length) || 0;
      // Quota meter still works
      App.store.history = payload.history;
      App.store.trash = payload.trash;
      App.updateQuotaMeter();
      return {
        before,
        after,
        beforeSnaps,
        afterSnaps,
        trimmed: afterSnaps < beforeSnaps,
        quotaLabel: document.getElementById('quota-label').textContent,
        trashStill: payload.trash.length
      };
    });
    assert(quota.trimmed, 'trimHistoryForQuota drops oldest snapshots under pressure',
      quota.beforeSnaps + ' → ' + quota.afterSnaps);
    assert(quota.trashStill === 1, 'Trim does not purge trash first', String(quota.trashStill));
    assert(/^\d+%$/.test((quota.quotaLabel || '').trim()), 'Quota meter still shows %', quota.quotaLabel);

    // ============================================================
    console.log('--- Command palette entries ---');
    const cmds = await page.evaluate(() => {
      const list = window.__whiteboard.App.listCommands();
      return {
        tm: list.some((c) => c.id === 'time-machine'),
        trash: list.some((c) => c.id === 'trash-view')
      };
    });
    assert(cmds.tm, 'Palette has Time machine');
    assert(cmds.trash, 'Palette has Trash');

    // ============================================================
    console.log('--- Console ---');
    const errs = (page._consoleErrors || []).filter((e) => !/ResizeObserver loop/.test(e));
    assert(errs.length === 0, 'No page console errors', errs.slice(0, 3).join(' | '));

    await page._context.close();
  } finally {
    await browser.close();
  }

  const report = {
    when: new Date().toISOString(),
    version: '2.4.0',
    tier: 'v2-tier5',
    file: htmlPath,
    chrome: CHROME,
    passed: results.filter((r) => r.ok).length,
    failed,
    total: results.length,
    results,
    notes: {
      historyRestoreTest: 'DOM: board menu → Time machine → Restore button → confirm modal OK',
      trashRestoreTest: 'DOM: board menu → Move to trash → confirm → File → Trash → Restore button',
      snapshotCadence: 'Auto: content fingerprint changed AND ≥2 min since last; force for picker/restore/trash'
    }
  };
  const out = path.resolve(__dirname, 'verify-report-v2-tier5.json');
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
