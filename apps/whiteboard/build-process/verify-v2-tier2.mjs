/**
 * Whiteboard v2 Tier 2 verification — real Chromium via playwright-core.
 * Features: auto-checklist, auto-color (+ settings), auto-tags, command palette.
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

async function openNotepad(page) {
  await page.evaluate(async () => {
    const App = window.__whiteboard.App;
    App.store.boards.push({
      id: 'board_notes_t2',
      name: 'Tier2 Notes',
      templateId: 'notepad',
      state: { notes: [] },
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
    });
    App.openBoard('board_notes_t2', { force: true });
    await App.flushSave({ forceOverwrite: true, sync: true });
  });
  await waitBoard(page, 400);
}

async function main() {
  console.log('Whiteboard v2 Tier-2 verification');
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
    console.log('--- Shell / shortcuts ---');
    let page = await freshPage(browser);
    const ver = await page.evaluate(() => window.__whiteboard.VERSION);
    assert(ver === '2.1.0', 'App version 2.1.0', ver);
    assert(await page.locator('#btn-settings').count() === 1, 'Settings button present');
    assert(await page.locator('#btn-cmd-palette').count() === 1, 'Command palette button present');
    assert(await page.locator('#tag-bar').count() === 1, 'Tag bar present');
    assert(await page.locator('#global-search').count() === 1, 'Content search still present');

    // Ctrl+K still focuses content search (not command palette)
    await page.keyboard.press('Control+k');
    await page.waitForTimeout(100);
    const focusedSearch = await page.evaluate(() => document.activeElement && document.activeElement.id);
    assert(focusedSearch === 'global-search', 'Ctrl+K focuses content search', focusedSearch);

    // Ctrl+/ opens command palette
    await page.keyboard.press('Control+/');
    await page.waitForTimeout(150);
    assert(await page.locator('.cmd-backdrop').count() === 1, 'Ctrl+/ opens command palette');
    assert(await page.locator('#cmd-input').count() === 1, 'Command palette input present');
    await page.keyboard.press('Escape');
    await page.waitForTimeout(100);
    assert(await page.locator('.cmd-backdrop').count() === 0, 'Escape closes command palette');

    // ============================================================
    console.log('--- Auto-checklist ---');
    await openNotepad(page);

    const checklist = await page.evaluate(() => {
      const wb = window.__whiteboard;
      // Unit: parsers
      const u = {
        has: wb.hasChecklistLines('- one\nplain'),
        no: wb.hasChecklistLines('plain only'),
        toggleStar: wb.toggleChecklistAtLine('* buy milk', 0),
        toggleBox: wb.toggleChecklistAtLine('[ ] task', 0),
        uncheck: wb.toggleChecklistAtLine('[x] done', 0),
        parse: wb.parseChecklistLine('- hello')
      };

      // Integration: create note with checklist text
      const App = wb.App;
      App.instance.addNote({
        text: '- unchecked item\n[ ] box item\nplain line\n[x] already done',
        x: 40, y: 40, w: 240, h: 220
      });
      App.snapshotCurrent();
      return u;
    });
    await page.waitForTimeout(200);

    assert(checklist.has === true, 'Detects checklist lines');
    assert(checklist.no === false, 'Plain text is not checklist');
    assert(checklist.toggleStar.indexOf('[x]') === 0 || checklist.toggleStar.indexOf('[x]') >= 0,
      'Toggle * line to [x]', checklist.toggleStar);
    assert(checklist.toggleBox.indexOf('[x]') !== -1, 'Toggle [ ] to [x]', checklist.toggleBox);
    assert(checklist.uncheck.indexOf('[ ]') !== -1, 'Toggle [x] to [ ]', checklist.uncheck);
    assert(checklist.parse.type === 'check' && checklist.parse.checked === false, 'Parse bullet line');

    const clUi = await page.evaluate(() => {
      const note = document.querySelector('.sticky-note');
      if (!note) return { ok: false, reason: 'no note' };
      const cl = note.querySelector('.note-checklist');
      const boxes = cl ? cl.querySelectorAll('input[type="checkbox"]') : [];
      const struck = cl ? cl.querySelectorAll('.cl-line.checked') : [];
      return {
        ok: true,
        hasChecklist: cl && !cl.classList.contains('hidden'),
        boxCount: boxes.length,
        checkedCount: struck.length,
        firstChecked: boxes[0] ? boxes[0].checked : null
      };
    });
    assert(clUi.ok && clUi.hasChecklist, 'Checklist view rendered for checklist note', JSON.stringify(clUi));
    assert(clUi.boxCount >= 3, 'At least 3 checkboxes rendered', String(clUi.boxCount));
    assert(clUi.checkedCount >= 1, 'Checked line has strikethrough class', String(clUi.checkedCount));

    // Click a checkbox and verify text source of truth updates
    await page.evaluate(() => {
      const cb = document.querySelector('.sticky-note .note-checklist input[type="checkbox"]');
      if (cb && !cb.checked) {
        cb.click();
      }
    });
    await page.waitForTimeout(150);
    const afterToggle = await page.evaluate(() => {
      const App = window.__whiteboard.App;
      App.snapshotCurrent();
      const notes = App.currentBoard().state.notes;
      return notes[0] && notes[0].text;
    });
    assert(/\[x\]/i.test(afterToggle || ''), 'Checkbox toggle updates underlying text', afterToggle);

    // ============================================================
    console.log('--- Auto-color ---');
    const colors = await page.evaluate(() => {
      const wb = window.__whiteboard;
      const rules = wb.defaultAutoColorRules();
      return {
        todo: wb.resolveAutoColor('Please TODO this', rules),
        urgent: wb.resolveAutoColor('URGENT: ship', rules),
        asap: wb.resolveAutoColor('need this ASAP', rules),
        done: wb.resolveAutoColor('DONE yesterday', rules),
        check: wb.resolveAutoColor('All ✓ complete', rules),
        questions: wb.resolveAutoColor('What about timing?? And who??', rules),
        plain: wb.resolveAutoColor('Just a normal note.', rules)
      };
    });
    assert(colors.todo && colors.todo.color === 'red', 'TODO → red', JSON.stringify(colors.todo));
    assert(colors.urgent && colors.urgent.color === 'red', 'URGENT → red');
    assert(colors.asap && colors.asap.color === 'red', 'ASAP → red');
    assert(colors.done && colors.done.color === 'green', 'DONE → green', JSON.stringify(colors.done));
    assert(colors.check && colors.check.color === 'green', '✓ → green');
    assert(colors.questions && colors.questions.color === 'yellow', 'Multiple ?? → yellow', JSON.stringify(colors.questions));
    assert(colors.plain === null, 'Plain statement has no auto color');

    // Live note color class
    await page.evaluate(() => {
      const App = window.__whiteboard.App;
      App.instance.setState({ notes: [] });
      App.instance.addNote({ text: 'TODO finish report #work', x: 50, y: 50 });
    });
    await page.waitForTimeout(200);
    const colorClass = await page.evaluate(() => {
      const notes = document.querySelectorAll('.sticky-note');
      const el = notes[notes.length - 1] || document.querySelector('.sticky-note');
      const App = window.__whiteboard.App;
      App.snapshotCurrent();
      const st = App.currentBoard().state.notes;
      return {
        classes: el ? el.className : '',
        isRed: el && el.classList.contains('color-red'),
        auto: el && el.classList.contains('auto-colored'),
        count: notes.length,
        stateColor: st[0] && st[0].color,
        stateText: st[0] && st[0].text
      };
    });
    assert(colorClass.count === 1, 'Single note after clear+add', String(colorClass.count));
    assert(colorClass.isRed || colorClass.stateColor === 'red', 'TODO note gets red background',
      JSON.stringify(colorClass));

    // Manual override via color cycle button
    await page.evaluate(() => {
      const btn = document.querySelector('.sticky-note .note-header button[title*="Cycle color"]');
      if (btn) btn.click();
    });
    await page.waitForTimeout(100);
    const overridden = await page.evaluate(() => {
      const App = window.__whiteboard.App;
      App.snapshotCurrent();
      const n = App.currentBoard().state.notes[0];
      return {
        colorAuto: n.colorAuto,
        color: n.color,
        locked: n.lockedTriggerKey
      };
    });
    assert(overridden.colorAuto === false, 'Manual color sets colorAuto false', JSON.stringify(overridden));

    // Change trigger text → re-enable auto
    await page.evaluate(() => {
      const App = window.__whiteboard.App;
      const n = App.currentBoard().state.notes[0];
      // Simulate text change that switches trigger to DONE
      App.instance.setState({
        notes: [{
          ...n,
          text: 'DONE finish report #work',
          colorAuto: false,
          lockedTriggerKey: n.lockedTriggerKey
        }]
      });
      App.instance.reapplyAutoColors();
      // reapply only applies when colorAuto true — need to update via instance path
      // Call apply by re-setting through live path: open edit
      App.snapshotCurrent();
    });
    // Force: add note and change text via live input path
    await page.evaluate(() => {
      const App = window.__whiteboard.App;
      App.instance.setState({ notes: [] });
      App.instance.addNote({
        id: 'n_override_1',
        text: 'TODO locked',
        x: 40, y: 40,
        color: 'red',
        colorAuto: false,
        lockedTriggerKey: 'urgent:TODO'
      });
      // Now change text to DONE (trigger change) via setState + re-run logic
      const st = App.instance.getState();
      const n = st.notes[0];
      n.text = 'DONE unlocked';
      // colorAuto false + lockedTriggerKey urgent:TODO, new key done:DONE → should re-enable
      App.instance.setState({ notes: [n] });
      // setState calls applyAutoColor which should re-enable
      App.snapshotCurrent();
    });
    await page.waitForTimeout(100);
    const reenabled = await page.evaluate(() => {
      const App = window.__whiteboard.App;
      App.snapshotCurrent();
      const n = App.currentBoard().state.notes[0];
      return { color: n.color, colorAuto: n.colorAuto, text: n.text };
    });
    assert(reenabled.color === 'green', 'Trigger change re-enables auto-color to green', JSON.stringify(reenabled));

    // Settings panel editable
    await page.locator('#btn-settings').click();
    await page.waitForTimeout(150);
    assert(await page.locator('.settings-rules').count() === 1, 'Settings panel opens');
    await page.evaluate(() => {
      const inp = document.querySelector('.settings-rules input[type="text"]');
      if (inp) {
        inp.value = 'TODO, URGENT, ASAP, CRITICAL';
        inp.dispatchEvent(new Event('input', { bubbles: true }));
      }
    });
    await page.locator('.modal-actions button.active').click();
    await page.waitForTimeout(150);
    const settingsSaved = await page.evaluate(() => {
      const rules = window.__whiteboard.App.store.settings.autoColorRules;
      const urgent = rules.find((r) => r.id === 'urgent');
      return urgent && urgent.keywords && urgent.keywords.indexOf('CRITICAL') >= 0;
    });
    assert(settingsSaved, 'Settings save updated keyword map with CRITICAL');

    // CRITICAL now maps to red
    const crit = await page.evaluate(() => {
      const rules = window.__whiteboard.App.store.settings.autoColorRules;
      return window.__whiteboard.resolveAutoColor('This is CRITICAL', rules);
    });
    assert(crit && crit.color === 'red', 'Custom keyword CRITICAL → red', JSON.stringify(crit));

    // ============================================================
    console.log('--- Auto-tags ---');
    await page.evaluate(() => {
      const App = window.__whiteboard.App;
      App.instance.setState({
        notes: [
          { id: 'n1', x: 20, y: 20, w: 180, h: 140, text: 'Ship #alpha and #beta', color: 'yellow', colorAuto: true },
          { id: 'n2', x: 220, y: 20, w: 180, h: 140, text: 'Only #alpha here', color: 'blue', colorAuto: true },
          { id: 'n3', x: 420, y: 20, w: 180, h: 140, text: 'No tags at all', color: 'green', colorAuto: true }
        ]
      });
      // remount notes
      const root = document.getElementById('board-root');
      App.openBoard(App.currentBoardId, { force: true });
    });
    await waitBoard(page, 450);
    await page.waitForTimeout(200);

    const tags = await page.evaluate(() => {
      const App = window.__whiteboard.App;
      App.refreshTagBar();
      const chips = Array.from(document.querySelectorAll('#tag-bar-chips .tag-chip')).map((c) => c.textContent);
      return {
        barVisible: document.getElementById('tag-bar').classList.contains('visible'),
        chips,
        collected: App.collectBoardTags(App.currentBoard())
      };
    });
    assert(tags.barVisible, 'Tag bar visible when tags present');
    assert(tags.collected.indexOf('alpha') >= 0 && tags.collected.indexOf('beta') >= 0,
      'Collects #alpha and #beta', JSON.stringify(tags.collected));
    assert(tags.chips.some((c) => c === '#alpha'), 'Chip for #alpha', JSON.stringify(tags.chips));

    // Click filter
    await page.evaluate(() => {
      const chip = Array.from(document.querySelectorAll('#tag-bar-chips .tag-chip'))
        .find((c) => c.textContent === '#beta');
      if (chip) chip.click();
    });
    await page.waitForTimeout(100);
    const filtered = await page.evaluate(() => {
      const notes = Array.from(document.querySelectorAll('.sticky-note'));
      return {
        filter: window.__whiteboard.App.activeTagFilter,
        dimmed: notes.filter((n) => n.classList.contains('tag-filtered-out')).length,
        total: notes.length,
        activeChip: !!document.querySelector('#tag-bar-chips .tag-chip.active')
      };
    });
    assert(filtered.filter === 'beta', 'Active filter is beta', filtered.filter);
    assert(filtered.dimmed === 2, 'Two notes dimmed when filtering #beta', JSON.stringify(filtered));
    assert(filtered.activeChip, 'Active chip styled');

    // Toggle off
    await page.evaluate(() => {
      const chip = Array.from(document.querySelectorAll('#tag-bar-chips .tag-chip'))
        .find((c) => c.textContent === '#beta');
      if (chip) chip.click();
    });
    await page.waitForTimeout(100);
    const cleared = await page.evaluate(() => ({
      filter: window.__whiteboard.App.activeTagFilter,
      dimmed: document.querySelectorAll('.sticky-note.tag-filtered-out').length
    }));
    assert(cleared.filter === null, 'Click again clears filter');
    assert(cleared.dimmed === 0, 'No notes dimmed after clear');

    // Wordpad + annotate tags collected when those boards active
    await page.evaluate(async () => {
      const App = window.__whiteboard.App;
      App.store.boards.push({
        id: 'board_wp_t2',
        name: 'WP Tags',
        templateId: 'wordpad',
        state: { html: '<p>Discuss #roadmap and #q3</p>' },
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      });
      App.store.boards.push({
        id: 'board_snip_t2',
        name: 'Snip Tags',
        templateId: 'snip',
        state: {
          tool: 'text',
          textLayers: [{ text: 'Label #screenshot', x: 0, y: 0 }],
          imageDataUrl: null
        },
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      });
      await App.flushSave({ forceOverwrite: true, sync: true });
    });
    const wpTags = await page.evaluate(() => {
      const App = window.__whiteboard.App;
      const b = App.store.boards.find((x) => x.id === 'board_wp_t2');
      return App.collectBoardTags(b);
    });
    const snipTags = await page.evaluate(() => {
      const App = window.__whiteboard.App;
      const b = App.store.boards.find((x) => x.id === 'board_snip_t2');
      return App.collectBoardTags(b);
    });
    assert(wpTags.indexOf('roadmap') >= 0 && wpTags.indexOf('q3') >= 0,
      'Wordpad tags extracted', JSON.stringify(wpTags));
    assert(snipTags.indexOf('screenshot') >= 0, 'Annotate textLayers tags extracted', JSON.stringify(snipTags));

    // ============================================================
    console.log('--- Command palette ---');
    await page.locator('#btn-cmd-palette').click();
    await page.waitForTimeout(150);
    assert(await page.locator('.cmd-backdrop').count() === 1, 'Palette opens from button');

    const cmds = await page.evaluate(() => {
      const list = window.__whiteboard.App.listCommands();
      return {
        total: list.length,
        hasBoards: list.some((c) => c.group === 'Boards'),
        hasTheme: list.some((c) => c.id === 'theme'),
        hasExport: list.some((c) => c.id === 'export-all'),
        hasSearch: list.some((c) => c.id === 'search'),
        hasSettings: list.some((c) => c.id === 'settings'),
        boardNames: list.filter((c) => c.id.indexOf('board:') === 0).map((c) => c.title)
      };
    });
    assert(cmds.total >= 15, 'Palette lists many commands', String(cmds.total));
    assert(cmds.hasBoards && cmds.boardNames.length >= 1, 'Lists boards', JSON.stringify(cmds.boardNames));
    assert(cmds.hasTheme, 'Lists toggle theme');
    assert(cmds.hasExport, 'Lists export all');
    assert(cmds.hasSearch, 'Lists content search action');
    assert(cmds.hasSettings, 'Lists settings');

    // Fuzzy filter
    await page.locator('#cmd-input').fill('theme');
    await page.waitForTimeout(100);
    const fuzzy = await page.evaluate(() => {
      const items = Array.from(document.querySelectorAll('.cmd-item .cmd-title')).map((e) => e.textContent);
      return items;
    });
    assert(fuzzy.some((t) => /theme/i.test(t)), 'Fuzzy filter finds theme', JSON.stringify(fuzzy));

    // Run theme command via Enter
    const themeBefore = await page.evaluate(() => document.documentElement.getAttribute('data-theme'));
    await page.keyboard.press('Enter');
    await page.waitForTimeout(150);
    const themeAfter = await page.evaluate(() => document.documentElement.getAttribute('data-theme'));
    assert(themeBefore !== themeAfter, 'Palette action toggles theme', themeBefore + ' → ' + themeAfter);
    assert(await page.locator('.cmd-backdrop').count() === 0, 'Palette closes after run');

    // Open palette, navigate to a board
    await page.keyboard.press('Control+/');
    await page.waitForTimeout(100);
    await page.locator('#cmd-input').fill('Tier2 Notes');
    await page.waitForTimeout(100);
    await page.keyboard.press('Enter');
    await waitBoard(page, 400);
    const onNotes = await page.evaluate(() => {
      const b = window.__whiteboard.App.currentBoard();
      return b && b.name;
    });
    assert(onNotes === 'Tier2 Notes', 'Palette switches board by name', onNotes);

    // Confirm Ctrl+K still content search, not palette
    await page.keyboard.press('Control+k');
    await page.waitForTimeout(80);
    assert(await page.locator('.cmd-backdrop').count() === 0, 'Ctrl+K does not open command palette');
    assert(
      await page.evaluate(() => document.activeElement && document.activeElement.id === 'global-search'),
      'Ctrl+K still focuses content search'
    );

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
    version: '2.1.0',
    tier: 'v2-tier2',
    file: htmlPath,
    chrome: CHROME,
    passed: results.filter((r) => r.ok).length,
    failed,
    total: results.length,
    results
  };
  const out = path.resolve(__dirname, 'verify-report-v2-tier2.json');
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
