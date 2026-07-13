/**
 * Whiteboard v2 Tier 4 — Kanban template + sticky promote.
 * Real Chromium via playwright-core.
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
  const context = await browser.newContext({ viewport: { width: 1280, height: 860 } });
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
  console.log('Whiteboard v2 Tier-4 verification');
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
    console.log('--- Registry / version ---');
    const ver = await page.evaluate(() => window.__whiteboard.VERSION);
    assert(ver === '2.3.0', 'App version 2.3.0', ver);
    const reg = await page.evaluate(() =>
      window.__whiteboard.TemplateRegistry.list().map((t) => t.id).sort()
    );
    assert(
      reg.join(',') === 'kanban,notepad,paint,snip,wordpad',
      'All 5 templates registered',
      reg.join(',')
    );

    // ============================================================
    console.log('--- Kanban board create / defaults ---');
    const created = await page.evaluate(async () => {
      const App = window.__whiteboard.App;
      const wb = window.__whiteboard;
      App.createBoard({
        name: 'Sprint Board',
        templateId: 'kanban',
        state: wb.defaultKanbanState(),
        open: true
      });
      await new Promise((r) => setTimeout(r, 250));
      App.snapshotCurrent();
      const b = App.currentBoard();
      const st = b.state;
      return {
        templateId: b.templateId,
        name: b.name,
        colNames: st.columns.map((c) => c.name),
        colCounts: st.columns.map((c) => c.cardIds.length),
        emptyCards: Object.keys(st.cards || {}).length,
        uiCols: document.querySelectorAll('.kanban-col').length,
        wipBadges: Array.from(document.querySelectorAll('.kanban-col-count')).map((e) => e.textContent)
      };
    });
    assert(created.templateId === 'kanban', 'Opened Kanban board');
    assert(
      created.colNames.join('|') === 'To Do|In Progress|Done',
      'Default 3 columns',
      created.colNames.join('|')
    );
    assert(created.uiCols === 3, 'UI shows 3 columns', String(created.uiCols));
    assert(created.wipBadges.join(',') === '0,0,0', 'WIP counts start at 0', created.wipBadges.join(','));

    // ============================================================
    console.log('--- Cards, auto-color, WIP, real mouse drag ---');
    // NOTE: Drag is pointer-based (not HTML5 DnD). Tests must drive real
    // mouse down → move → up on the card grip, not call moveCard() directly.
    const cardsMeta = await page.evaluate(async () => {
      const App = window.__whiteboard.App;
      const st = App.currentBoard().state;
      const todo = st.columns[0].id;
      App.instance.addCard(todo, 'TODO build feature #sprint');
      App.instance.addCard(todo, 'DONE polish docs');
      await new Promise((r) => setTimeout(r, 120));
      App.snapshotCurrent();
      const state = App.currentBoard().state;
      const cardIds = state.columns[0].cardIds.slice();
      return {
        c0color: state.cards[cardIds[0]].color,
        c1color: state.cards[cardIds[1]].color,
        c0id: cardIds[0],
        c1id: cardIds[1],
        todoId: state.columns[0].id,
        doingId: state.columns[1].id
      };
    });
    assert(cardsMeta.c0color === 'red', 'TODO card auto-color red', cardsMeta.c0color);
    assert(cardsMeta.c1color === 'green', 'DONE card auto-color green', cardsMeta.c1color);

    // Real mouse drag: grip of first card → center of In Progress column
    const dragResult = await page.evaluate(async (meta) => {
      const grip = document.querySelector(
        '.kanban-card[data-id="' + CSS.escape(meta.c0id) + '"] .grip'
      );
      const targetCol = document.querySelector(
        '.kanban-col[data-id="' + CSS.escape(meta.doingId) + '"] .kanban-cards'
      );
      if (!grip || !targetCol) return { ok: false, reason: 'missing grip or target' };
      const g = grip.getBoundingClientRect();
      const t = targetCol.getBoundingClientRect();
      const startX = g.left + g.width / 2;
      const startY = g.top + g.height / 2;
      const endX = t.left + t.width / 2;
      const endY = t.top + Math.min(40, t.height / 2);

      function fire(type, x, y, target) {
        const ev = new PointerEvent(type, {
          bubbles: true,
          cancelable: true,
          clientX: x,
          clientY: y,
          pointerId: 1,
          pointerType: 'mouse',
          button: 0,
          buttons: type === 'pointerup' ? 0 : 1
        });
        target.dispatchEvent(ev);
      }

      fire('pointerdown', startX, startY, grip);
      // Intermediate moves so hit-test updates
      const steps = 8;
      for (let i = 1; i <= steps; i++) {
        const x = startX + ((endX - startX) * i) / steps;
        const y = startY + ((endY - startY) * i) / steps;
        fire('pointermove', x, y, window);
        await new Promise((r) => setTimeout(r, 16));
      }
      fire('pointerup', endX, endY, window);
      await new Promise((r) => setTimeout(r, 80));

      window.__whiteboard.App.snapshotCurrent();
      const state = window.__whiteboard.App.currentBoard().state;
      const badges = Array.from(document.querySelectorAll('.kanban-col-count')).map((e) => e.textContent);
      return {
        ok: true,
        counts: state.columns.map((c) => c.cardIds.length),
        badges,
        doingHasC0: state.columns[1].cardIds.indexOf(meta.c0id) >= 0,
        todoHasC0: state.columns[0].cardIds.indexOf(meta.c0id) >= 0,
        ghostLeft: !document.querySelector('.kanban-card-ghost')
      };
    }, cardsMeta);

    assert(dragResult.ok, 'Pointer drag gesture ran', JSON.stringify(dragResult));
    assert(dragResult.doingHasC0 && !dragResult.todoHasC0,
      'Real pointer drag moved card To Do → In Progress',
      JSON.stringify(dragResult));
    assert(dragResult.counts[0] === 1 && dragResult.counts[1] === 1,
      'Column card counts after drag', JSON.stringify(dragResult.counts));
    assert(dragResult.badges.join(',') === '1,1,0', 'WIP badges update after drag',
      dragResult.badges.join(','));
    assert(dragResult.ghostLeft, 'Drag ghost cleaned up');

    // Reorder within a column via real pointer drag (card below the other)
    const reorder = await page.evaluate(async (meta) => {
      // Ensure both cards in In Progress for reorder test: move remaining todo card too
      // Actually after first drag: todo has c1, doing has c0. Move c1 to doing as well via pointer.
      const grip = document.querySelector(
        '.kanban-card[data-id="' + CSS.escape(meta.c1id) + '"] .grip'
      );
      const other = document.querySelector(
        '.kanban-card[data-id="' + CSS.escape(meta.c0id) + '"]'
      );
      if (!grip || !other) return { ok: false, reason: 'missing elements' };
      const g = grip.getBoundingClientRect();
      const o = other.getBoundingClientRect();
      const startX = g.left + g.width / 2;
      const startY = g.top + g.height / 2;
      // Drop above other card (before) so order becomes [c1, c0] in doing — first move c1 into doing
      const endX = o.left + o.width / 2;
      const endY = o.top + 4; // top edge → placeBefore

      function fire(type, x, y, target) {
        target.dispatchEvent(new PointerEvent(type, {
          bubbles: true, cancelable: true,
          clientX: x, clientY: y, pointerId: 1, pointerType: 'mouse',
          button: 0, buttons: type === 'pointerup' ? 0 : 1
        }));
      }
      fire('pointerdown', startX, startY, grip);
      for (let i = 1; i <= 6; i++) {
        fire('pointermove',
          startX + ((endX - startX) * i) / 6,
          startY + ((endY - startY) * i) / 6,
          window);
        await new Promise((r) => setTimeout(r, 16));
      }
      fire('pointerup', endX, endY, window);
      await new Promise((r) => setTimeout(r, 80));
      window.__whiteboard.App.snapshotCurrent();
      const st = window.__whiteboard.App.currentBoard().state;
      return {
        ok: true,
        doingOrder: st.columns[1].cardIds.slice(),
        todoCount: st.columns[0].cardIds.length
      };
    }, cardsMeta);
    assert(reorder.ok, 'In-column/cross reorder gesture ran', JSON.stringify(reorder));
    assert(reorder.todoCount === 0 && reorder.doingOrder.length === 2,
      'Second card dragged into In Progress',
      JSON.stringify(reorder));

    // Persist: flushSave + reload, card stays in In Progress
    await page.evaluate(() => window.__whiteboard.App.flushSave({ forceOverwrite: true, sync: true }));
    await page.reload({ waitUntil: 'domcontentloaded' });
    await waitBoard(page, 400);
    const persisted = await page.evaluate(() => {
      const App = window.__whiteboard.App;
      // open Sprint Board if not current
      const k = App.store.boards.find((b) => b.templateId === 'kanban' && b.name === 'Sprint Board');
      if (k) App.openBoard(k.id, { force: true, silent: true });
      return new Promise((resolve) => {
        setTimeout(() => {
          App.snapshotCurrent();
          const st = App.currentBoard().state;
          resolve({
            templateId: App.currentBoard().templateId,
            counts: st.columns.map((c) => c.cardIds.length),
            doingTexts: st.columns[1].cardIds.map((id) => st.cards[id] && st.cards[id].text)
          });
        }, 300);
      });
    });
    assert(persisted.templateId === 'kanban', 'Reloaded Kanban board');
    assert(persisted.counts[1] === 2, 'Cards still in In Progress after reload',
      JSON.stringify(persisted.counts));
    assert(
      persisted.doingTexts.some((t) => /feature/.test(t || '')) &&
      persisted.doingTexts.some((t) => /polish/.test(t || '')),
      'Persisted card text in In Progress',
      JSON.stringify(persisted.doingTexts)
    );

    const searchTags = await page.evaluate(() => {
      const App = window.__whiteboard.App;
      return {
        searchHit: App.searchAll('feature').some((s) => s.kind === 'kanban-card'),
        tags: App.collectBoardTags(App.currentBoard())
      };
    });
    assert(searchTags.searchHit, 'Global search finds kanban card text');
    assert(searchTags.tags.indexOf('sprint') >= 0, 'Tag bar extracts #sprint from cards',
      JSON.stringify(searchTags.tags));

    // Column add/rename via instance
    const colOps = await page.evaluate(async () => {
      const App = window.__whiteboard.App;
      App.instance.addColumn('Blocked');
      await new Promise((r) => setTimeout(r, 80));
      App.snapshotCurrent();
      const st = App.currentBoard().state;
      return {
        names: st.columns.map((c) => c.name),
        ui: document.querySelectorAll('.kanban-col').length
      };
    });
    assert(colOps.names.indexOf('Blocked') >= 0, 'Add column works', colOps.names.join('|'));
    assert(colOps.ui === 4, 'UI has 4 columns', String(colOps.ui));

    // ============================================================
    console.log('--- Promote single sticky to Kanban ---');
    // Create notepad with sticky
    await page.evaluate(async () => {
      const App = window.__whiteboard.App;
      App.createBoard({
        name: 'Ideas',
        templateId: 'notepad',
        state: {
          notes: [{
            id: 'n_promo',
            x: 40, y: 40, w: 200, h: 180,
            text: 'URGENT call client #sales',
            color: 'yellow',
            colorAuto: true
          }]
        },
        open: true
      });
      await new Promise((r) => setTimeout(r, 250));
    });
    await waitBoard(page, 200);

    // Send sticky via API (same as header button)
    const promo = await page.evaluate(async () => {
      const App = window.__whiteboard.App;
      App.snapshotCurrent();
      const note = App.currentBoard().state.notes[0];
      const beforeKanbans = App.listKanbanBoards().length;
      // Force path: add to existing Sprint Board
      const k = App.listKanbanBoards().find((b) => b.name === 'Sprint Board');
      const stickyTextBefore = note.text;
      const stickyId = note.id;
      const added = App.addCardToKanbanBoard(k.id, note.text, {
        columnId: App.ensureKanbanState(k).columns[0].id,
        open: true
      });
      await new Promise((r) => setTimeout(r, 250));
      // Switch back to Ideas and verify sticky intact
      const ideas = App.store.boards.find((b) => b.name === 'Ideas');
      App.openBoard(ideas.id, { force: true });
      await new Promise((r) => setTimeout(r, 250));
      App.snapshotCurrent();
      const still = App.currentBoard().state.notes.find((n) => n.id === stickyId);
      // Check kanban has card
      const k2 = App.getBoard(k.id);
      const cardTexts = Object.values(k2.state.cards || {}).map((c) => c.text);
      return {
        beforeKanbans,
        afterKanbans: App.listKanbanBoards().length,
        added: !!added,
        stickyText: still && still.text,
        stickyIntact: still && still.text === stickyTextBefore,
        cardTexts
      };
    });
    assert(promo.added, 'Card added to existing Kanban');
    assert(promo.stickyIntact, 'Original sticky unchanged', promo.stickyText);
    assert(promo.cardTexts.some((t) => /URGENT call client/.test(t)), 'Kanban has sticky text as card');

    // Auto-create Kanban when none exist: wipe kanbans then send
    const autoCreate = await page.evaluate(async () => {
      const App = window.__whiteboard.App;
      // Remove all kanban boards
      App.store.boards = App.store.boards.filter((b) => b.templateId !== 'kanban');
      App.snapshotCurrent();
      const ideas = App.store.boards.find((b) => b.name === 'Ideas');
      App.openBoard(ideas.id, { force: true });
      await new Promise((r) => setTimeout(r, 200));
      App.snapshotCurrent();
      const note = App.currentBoard().state.notes[0];
      // Simulate empty kanban path of promptSendStickyToKanban without modal
      const created = App.createBoard({
        name: 'Kanban',
        templateId: 'kanban',
        state: window.__whiteboard.defaultKanbanState(),
        open: false,
        silent: true
      });
      const r = App.addCardToKanbanBoard(created.id, note.text, { open: true });
      await new Promise((r2) => setTimeout(r2, 200));
      return {
        created: !!created,
        cardCount: Object.keys(App.getBoard(created.id).state.cards).length,
        openTemplate: App.currentBoard().templateId
      };
    });
    assert(autoCreate.created && autoCreate.cardCount === 1, 'Can create Kanban then add card');
    assert(autoCreate.openTemplate === 'kanban', 'Opens Kanban after send');

    // ============================================================
    console.log('--- Promote whole Sticky board to Kanban ---');
    const boardPromo = await page.evaluate(async () => {
      const App = window.__whiteboard.App;
      App.createBoard({
        name: 'Retro notes',
        templateId: 'notepad',
        state: {
          notes: [
            { id: 'n1', x: 10, y: 10, w: 180, h: 140, text: 'Keep shipping #retro', color: 'blue', colorAuto: true },
            { id: 'n2', x: 200, y: 10, w: 180, h: 140, text: 'TODO fix flaky tests', color: 'yellow', colorAuto: true },
            { id: 'n3', x: 400, y: 10, w: 180, h: 140, text: 'Celebrate wins', color: 'green', colorAuto: true }
          ]
        },
        open: true
      });
      await new Promise((r) => setTimeout(r, 250));
      const beforeBoards = App.store.boards.length;
      const beforeNotes = App.currentBoard().state.notes.length;
      App.promoteNotepadBoardToKanban();
      await new Promise((r) => setTimeout(r, 300));
      App.snapshotCurrent();
      const b = App.currentBoard();
      const st = b.state;
      const firstColCards = st.columns[0].cardIds.map((id) => st.cards[id].text);
      // Original board still has notes
      const orig = App.store.boards.find((x) => x.name === 'Retro notes');
      return {
        beforeBoards,
        afterBoards: App.store.boards.length,
        openName: b.name,
        openTemplate: b.templateId,
        firstColCount: st.columns[0].cardIds.length,
        firstColCards,
        origNotes: orig && orig.state.notes.length,
        beforeNotes
      };
    });
    assert(boardPromo.afterBoards === boardPromo.beforeBoards + 1, 'New board created');
    assert(boardPromo.openTemplate === 'kanban', 'Opened new Kanban');
    assert(/Kanban/i.test(boardPromo.openName), 'Name indicates Kanban', boardPromo.openName);
    assert(boardPromo.firstColCount === 3, 'All stickies became cards in first column',
      String(boardPromo.firstColCount));
    assert(boardPromo.origNotes === boardPromo.beforeNotes, 'Original notepad board untouched',
      String(boardPromo.origNotes));
    assert(boardPromo.firstColCards.some((t) => /flaky tests/.test(t)), 'Card text preserved');

    // ============================================================
    console.log('--- Command palette / UI affordances ---');
    const cmds = await page.evaluate(() => {
      const list = window.__whiteboard.App.listCommands();
      return {
        newKanban: list.some((c) => c.id === 'new-kanban'),
        promote: list.some((c) => c.id === 'promote-board-kanban'),
        hasKanbanBoard: list.some((c) => c.id.indexOf('board:') === 0 && /kanban/i.test(c.sub || ''))
      };
    });
    assert(cmds.newKanban, 'Command palette: New Kanban board');
    // promote only when on notepad — switch first
    await page.evaluate(async () => {
      const App = window.__whiteboard.App;
      const n = App.store.boards.find((b) => b.templateId === 'notepad');
      App.openBoard(n.id, { force: true });
      await new Promise((r) => setTimeout(r, 200));
    });
    const cmds2 = await page.evaluate(() => {
      const list = window.__whiteboard.App.listCommands();
      return {
        promote: list.some((c) => c.id === 'promote-board-kanban'),
        addNote: list.some((c) => c.id === 'add-note')
      };
    });
    assert(cmds2.promote, 'Command palette: promote stickies when on Notepad');

    // Sticky header button + board promote button present
    const ui = await page.evaluate(() => ({
      stickyKanbanBtn: !!document.querySelector('.sticky-note button[title*="Send to Kanban"]'),
      boardPromoteBtn: Array.from(document.querySelectorAll('#template-tools button'))
        .some((b) => /Kanban board/i.test(b.textContent || ''))
    }));
    assert(ui.stickyKanbanBtn, 'Sticky has Send to Kanban control');
    assert(ui.boardPromoteBtn, 'Notepad toolbar has promote-to-Kanban button');

    // ============================================================
    console.log('--- Quota meter still works ---');
    await page.evaluate(() => window.__whiteboard.App.flushSave({ forceOverwrite: true, sync: true }));
    await page.waitForTimeout(100);
    const quota = await page.locator('#quota-label').textContent();
    assert(/^\d+%$/.test((quota || '').trim()), 'Quota shows percentage', quota);

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
    version: '2.3.0',
    tier: 'v2-tier4',
    file: htmlPath,
    chrome: CHROME,
    passed: results.filter((r) => r.ok).length,
    failed,
    total: results.length,
    results
  };
  const out = path.resolve(__dirname, 'verify-report-v2-tier4.json');
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
