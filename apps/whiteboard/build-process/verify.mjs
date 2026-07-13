/**
 * Real-browser verification for whiteboard.html (Tier 2).
 * Windows Node + Google Chrome via playwright-core.
 */
import { chromium } from 'playwright-core';
import path from 'path';
import fs from 'fs';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const htmlPath = path.resolve(__dirname, 'whiteboard.html');
const fileUrl = 'file:///' + htmlPath.replace(/\\/g, '/');
const CHROME = process.env.CHROME_PATH || 'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe';
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
  await page.waitForFunction(() => {
    const root = document.getElementById('board-root');
    return root && !root.classList.contains('switching');
  }, { timeout: 5000 }).catch(() => {});
  await page.waitForTimeout(80);
}

async function main() {
  console.log('Whiteboard Tier-2 verification');
  console.log('File:', fileUrl);
  console.log('Chrome:', CHROME);
  console.log('');

  const browser = await chromium.launch({
    executablePath: CHROME,
    headless: true,
    args: ['--allow-file-access-from-files', '--disable-web-security']
  });

  try {
    const context = await browser.newContext({ viewport: { width: 1280, height: 860 } });
    const page = await context.newPage();
    const consoleErrors = [];
    page.on('console', (msg) => {
      if (msg.type() === 'error') consoleErrors.push(msg.text());
    });
    page.on('pageerror', (err) => consoleErrors.push(String(err)));

    // Fresh start
    await page.goto(fileUrl, { waitUntil: 'domcontentloaded' });
    await page.evaluate(() => localStorage.clear());
    await page.reload({ waitUntil: 'domcontentloaded' });
    await waitBoard(page, 300);

    // ---- Shell ----
    assert(await page.title() === 'Whiteboard', 'Page title');
    assert(await page.locator('#app').count() === 1, 'App shell present');
    assert(await page.locator('#board-select').count() === 1, 'Board select present');
    assert(await page.locator('#btn-new-board').count() === 1, 'New board button');
    assert(await page.locator('#btn-theme').count() === 1, 'Theme toggle present');
    assert(await page.locator('#btn-file-menu').count() === 1, 'File menu present');

    const regIds = await page.evaluate(() =>
      window.__whiteboard.TemplateRegistry.list().map((t) => t.id).sort()
    );
    assert(
      regIds.join(',') === 'notepad,paint,snip,wordpad',
      'All 4 templates registered',
      regIds.join(',')
    );

    // Default board after empty storage
    let boardCount = await page.locator('#board-select option').count();
    assert(boardCount === 1, 'Default single board', `count=${boardCount}`);
    let current = await page.evaluate(() => {
      const b = window.__whiteboard.App.currentBoard();
      return b && { templateId: b.templateId, name: b.name };
    });
    assert(current && current.templateId === 'paint', 'Default board is paint', JSON.stringify(current));

    // ---- Theme ----
    const themeBefore = await page.evaluate(() => document.documentElement.getAttribute('data-theme'));
    await page.locator('#btn-theme').click();
    await page.waitForTimeout(100);
    const themeAfter = await page.evaluate(() => document.documentElement.getAttribute('data-theme'));
    assert(themeBefore !== themeAfter, 'Theme toggles', `${themeBefore} -> ${themeAfter}`);
    await page.locator('#btn-theme').click(); // back to light for screenshots

    // ---- Paint draw + undo ----
    const canvas = page.locator('.paint-canvas-wrap canvas');
    await canvas.waitFor({ state: 'visible' });
    const box = await canvas.boundingBox();
    assert(box && box.width > 100, 'Paint canvas sized', JSON.stringify(box));

    await page.mouse.move(box.x + 40, box.y + 40);
    await page.mouse.down();
    await page.mouse.move(box.x + 180, box.y + 120, { steps: 15 });
    await page.mouse.up();
    await page.waitForTimeout(200);

    const nonWhite1 = await page.evaluate(() => {
      const c = document.querySelector('.paint-canvas-wrap canvas');
      const d = c.getContext('2d').getImageData(0, 0, c.width, c.height).data;
      let n = 0;
      for (let i = 0; i < d.length; i += 4) if (d[i] < 250 || d[i + 1] < 250 || d[i + 2] < 250) n++;
      return n;
    });
    assert(nonWhite1 > 50, 'Paint stroke drawn', `pixels=${nonWhite1}`);

    // Second stroke then undo
    await page.mouse.move(box.x + 60, box.y + 160);
    await page.mouse.down();
    await page.mouse.move(box.x + 220, box.y + 160, { steps: 10 });
    await page.mouse.up();
    await page.waitForTimeout(200);
    const nonWhite2 = await page.evaluate(() => {
      const c = document.querySelector('.paint-canvas-wrap canvas');
      const d = c.getContext('2d').getImageData(0, 0, c.width, c.height).data;
      let n = 0;
      for (let i = 0; i < d.length; i += 4) if (d[i] < 250 || d[i + 1] < 250 || d[i + 2] < 250) n++;
      return n;
    });
    await page.keyboard.press('Control+z');
    await page.waitForTimeout(300);
    const nonWhiteUndo = await page.evaluate(() => {
      const c = document.querySelector('.paint-canvas-wrap canvas');
      const d = c.getContext('2d').getImageData(0, 0, c.width, c.height).data;
      let n = 0;
      for (let i = 0; i < d.length; i += 4) if (d[i] < 250 || d[i + 1] < 250 || d[i + 2] < 250) n++;
      return n;
    });
    assert(nonWhiteUndo < nonWhite2, 'Paint undo reduces ink', `${nonWhite2} -> ${nonWhiteUndo}`);

    // ---- Multi-board: create Notepad ----
    await page.locator('#btn-new-board').click();
    await page.locator('.modal').waitFor({ state: 'visible' });
    await page.locator('.modal input').fill('My Notes');
    await page.locator('.template-picker button').filter({ hasText: 'Notepad' }).click();
    await page.locator('.modal-actions button.active').click();
    await waitBoard(page, 250);

    current = await page.evaluate(() => {
      const b = window.__whiteboard.App.currentBoard();
      return b && { templateId: b.templateId, name: b.name };
    });
    assert(current.templateId === 'notepad' && current.name === 'My Notes', 'Created notepad board', JSON.stringify(current));
    boardCount = await page.locator('#board-select option').count();
    assert(boardCount === 2, 'Two boards exist', `count=${boardCount}`);

    await page.locator('#template-tools button').filter({ hasText: 'Add note' }).click();
    await page.waitForTimeout(100);
    await page.locator('.sticky-note textarea').fill('Tier 2 note');
    await page.waitForTimeout(400);

    // ---- Create Wordpad ----
    await page.locator('#btn-new-board').click();
    await page.locator('.modal').waitFor({ state: 'visible' });
    await page.locator('.modal input').fill('Doc One');
    await page.locator('.template-picker button').filter({ hasText: 'Wordpad' }).click();
    await page.locator('.modal-actions button.active').click();
    await waitBoard(page, 250);

    const editor = page.locator('.wordpad-editor');
    await editor.click();
    await page.keyboard.type('Hello Wordpad');
    await page.waitForTimeout(100);
    // Select all and bold
    await page.keyboard.press('Control+a');
    await page.locator('#template-tools button').filter({ hasText: 'B' }).first().click();
    await page.waitForTimeout(400);
    const hasBold = await page.evaluate(() => {
      const ed = document.querySelector('.wordpad-editor');
      return !!(ed.querySelector('b, strong') || /font-weight:\s*bold/i.test(ed.innerHTML));
    });
    assert(hasBold, 'Wordpad bold formatting applied');
    const plain = await page.evaluate(() => window.__whiteboard.App.instance.exportText());
    assert(/Hello Wordpad/.test(plain || ''), 'Wordpad exportText has content', plain);

    // Wordpad undo: append text (move caret to end), then undo
    await page.locator('.wordpad-editor').click();
    await page.keyboard.press('End');
    await page.keyboard.press('Control+End');
    await page.keyboard.type(' EXTRA');
    await page.waitForTimeout(500);
    const beforeUndo = await page.locator('.wordpad-editor').innerText();
    assert(/EXTRA/.test(beforeUndo), 'Wordpad typed EXTRA before undo', beforeUndo);
    await page.keyboard.press('Control+z');
    await page.waitForTimeout(400);
    const afterUndo = await page.locator('.wordpad-editor').innerText();
    assert(
      beforeUndo !== afterUndo && !/EXTRA/.test(afterUndo),
      'Wordpad undo reverts EXTRA',
      `"${beforeUndo}" -> "${afterUndo}"`
    );
    // ---- Create Annotate (snip) ----
    await page.locator('#btn-new-board').click();
    await page.locator('.modal').waitFor({ state: 'visible' });
    await page.locator('.modal input').fill('Annotate Me');
    await page.locator('.template-picker button').filter({ hasText: 'Annotate' }).click();
    await page.locator('.modal-actions button.active').click();
    await waitBoard(page, 250);
    assert(await page.locator('.snip-stage').count() === 1, 'Annotate stage mounted');
    assert(await page.locator('.empty-hint').count() >= 1, 'Annotate empty hint visible');

    // Load a tiny PNG via data URL through the app API path
    await page.evaluate(async () => {
      // 2x2 red PNG
      const c = document.createElement('canvas');
      c.width = 200; c.height = 120;
      const g = c.getContext('2d');
      g.fillStyle = '#ddeeff';
      g.fillRect(0, 0, 200, 120);
      g.fillStyle = '#cc0000';
      g.fillRect(20, 20, 80, 40);
      const url = c.toDataURL('image/png');
      // Use instance setState + remount-like load by calling internal via getState path
      const inst = window.__whiteboard.App.instance;
      inst.setState({ imageDataUrl: url, tool: 'arrow', strokeColor: '#e03131' });
      // Force reload by re-opening board
      const id = window.__whiteboard.App.currentBoardId;
      window.__whiteboard.App.openBoard(id, { force: true });
    });
    await waitBoard(page, 400);
    const snipHasImage = await page.evaluate(() => {
      const b = window.__whiteboard.App.currentBoard();
      const st = window.__whiteboard.App.instance.getState();
      return !!(st && st.imageDataUrl && st.imageDataUrl.length > 100);
    });
    assert(snipHasImage, 'Annotate board holds image data');

    // Draw arrow annotation
    const snipCanvas = page.locator('.snip-canvas-wrap canvas');
    const sbox = await snipCanvas.boundingBox();
    if (sbox) {
      await page.locator('#template-tools button').filter({ hasText: 'Arrow' }).click();
      await page.mouse.move(sbox.x + 30, sbox.y + 30);
      await page.mouse.down();
      await page.mouse.move(sbox.x + 120, sbox.y + 80, { steps: 8 });
      await page.mouse.up();
      await page.waitForTimeout(300);
    }
    const snipPng = await page.evaluate(() => window.__whiteboard.App.instance.exportPng());
    assert(typeof snipPng === 'string' && snipPng.startsWith('data:image/png'), 'Annotate exportPng works');

    // Capture button presence (API availability probe)
    const captureInfo = await page.evaluate(() => {
      const available = !!(navigator.mediaDevices && navigator.mediaDevices.getDisplayMedia);
      const btn = document.querySelector('#template-tools button');
      const captureBtn = Array.from(document.querySelectorAll('#template-tools button'))
        .find((b) => /Capture screen/i.test(b.textContent || ''));
      return {
        apiAvailable: available,
        buttonPresent: !!captureBtn,
        buttonDisabled: captureBtn ? captureBtn.disabled : null,
        isFile: location.protocol === 'file:'
      };
    });
    assert(captureInfo.buttonPresent, 'Capture screen button present', JSON.stringify(captureInfo));
    pass('Capture API probe recorded', JSON.stringify(captureInfo));

    // ---- Board rename ----
    await page.locator('#btn-board-menu').click();
    await page.locator('#board-menu button[data-action="rename"]').click();
    await page.locator('.modal').waitFor({ state: 'visible' });
    await page.locator('.modal input').fill('Annotate Renamed');
    await page.locator('.modal-actions button.active').click();
    await page.waitForTimeout(150);
    const renamed = await page.evaluate(() => window.__whiteboard.App.currentBoard().name);
    assert(renamed === 'Annotate Renamed', 'Board renamed', renamed);

    // ---- Duplicate ----
    await page.locator('#btn-board-menu').click();
    await page.locator('#board-menu button[data-action="duplicate"]').click();
    await waitBoard(page, 250);
    boardCount = await page.locator('#board-select option').count();
    assert(boardCount >= 5, 'Duplicate increased board count', `count=${boardCount}`);
    const dupName = await page.evaluate(() => window.__whiteboard.App.currentBoard().name);
    assert(/copy/i.test(dupName), 'Duplicate name has copy', dupName);

    // ---- Switch boards via select (preserve state) ----
    const options = await page.evaluate(() =>
      Array.from(document.querySelectorAll('#board-select option')).map((o) => ({
        value: o.value,
        text: o.textContent
      }))
    );
    const notesOpt = options.find((o) => /My Notes/.test(o.text));
    assert(!!notesOpt, 'My Notes board listed');
    await page.locator('#board-select').selectOption(notesOpt.value);
    await waitBoard(page, 250);
    const noteText = await page.locator('.sticky-note textarea').first().inputValue();
    assert(noteText === 'Tier 2 note', 'Notepad state preserved across board switch', noteText);

    // ---- Export JSON (evaluate payload, don't rely on download) ----
    const exportPayload = await page.evaluate(() => {
      window.__whiteboard.App.snapshotCurrent();
      return {
        format: 'whiteboard-export',
        version: 2,
        boards: window.__whiteboard.App.store.boards
      };
    });
    assert(exportPayload.boards.length >= 4, 'Export payload has boards', String(exportPayload.boards.length));

    // Import adds boards
    const beforeImport = await page.locator('#board-select option').count();
    await page.evaluate((payload) => {
      window.__whiteboard.App.importData(payload);
    }, {
      format: 'whiteboard-export',
      version: 2,
      boards: [{
        name: 'Imported Paint',
        templateId: 'paint',
        state: { color: '#e03131', sizeId: 'l', imageDataUrl: null }
      }]
    });
    await waitBoard(page, 300);
    const afterImport = await page.locator('#board-select option').count();
    assert(afterImport === beforeImport + 1, 'Import adds a board', `${beforeImport} -> ${afterImport}`);
    const imported = await page.evaluate(() => window.__whiteboard.App.currentBoard().name);
    assert(imported === 'Imported Paint', 'Opened imported board', imported);

    // ---- Persistence reload ----
    await page.waitForTimeout(600);
    const stored = await page.evaluate(() => {
      const raw = localStorage.getItem(window.__whiteboard.STORAGE_KEY);
      return raw ? JSON.parse(raw) : null;
    });
    assert(stored && stored.version === 2, 'localStorage is v2');
    assert(Array.isArray(stored.boards) && stored.boards.length >= 5, 'Persisted multiple boards', String(stored.boards && stored.boards.length));

    await page.reload({ waitUntil: 'domcontentloaded' });
    await waitBoard(page, 400);
    const afterReloadCount = await page.locator('#board-select option').count();
    assert(afterReloadCount === stored.boards.length, 'Boards survive reload', `count=${afterReloadCount}`);

    // Switch to My Notes after reload
    const opts2 = await page.evaluate(() =>
      Array.from(document.querySelectorAll('#board-select option')).map((o) => ({
        value: o.value, text: o.textContent
      }))
    );
    const notes2 = opts2.find((o) => /My Notes/.test(o.text));
    if (notes2) {
      await page.locator('#board-select').selectOption(notes2.value);
      await waitBoard(page, 250);
      const t = await page.locator('.sticky-note textarea').first().inputValue();
      assert(t === 'Tier 2 note', 'Note text survives full reload', t);
    } else {
      fail('My Notes after reload', 'option missing');
    }

    // ---- Delete board (keep at least one) ----
    const countBeforeDel = await page.locator('#board-select option').count();
    await page.locator('#btn-board-menu').click();
    await page.locator('#board-menu button[data-action="delete"]').click();
    await page.locator('.modal').waitFor({ state: 'visible' });
    await page.locator('.modal-actions button.active, .modal-actions button.danger').last().click();
    await waitBoard(page, 300);
    const countAfterDel = await page.locator('#board-select option').count();
    assert(countAfterDel === countBeforeDel - 1, 'Delete removes one board', `${countBeforeDel} -> ${countAfterDel}`);

    // ---- No external network ----
    const netPage = await context.newPage();
    const external = [];
    netPage.on('request', (req) => {
      const u = req.url();
      if (!u.startsWith('file:') && !u.startsWith('data:') && !u.startsWith('blob:')) external.push(u);
    });
    await netPage.goto(fileUrl, { waitUntil: 'domcontentloaded' });
    await netPage.waitForTimeout(500);
    await netPage.locator('#btn-new-board').click();
    await netPage.locator('.modal-actions button').filter({ hasText: 'Cancel' }).click();
    assert(external.length === 0, 'No external network requests', external.join(', ') || 'none');

    const serious = consoleErrors.filter((e) => !/favicon/i.test(e));
    assert(serious.length === 0, 'No console errors', serious.join(' | ') || 'none');

    // Screenshots
    const shotDir = path.join(__dirname, 'verify-screenshots');
    fs.mkdirSync(shotDir, { recursive: true });
    // Open paint for shot
    const paintOpt = (await page.evaluate(() =>
      Array.from(document.querySelectorAll('#board-select option')).map((o) => ({
        value: o.value, text: o.textContent
      }))
    )).find((o) => /Paint/i.test(o.text));
    if (paintOpt) {
      await page.locator('#board-select').selectOption(paintOpt.value);
      await waitBoard(page, 300);
    }
    await page.screenshot({ path: path.join(shotDir, 't2-01-paint.png'), fullPage: true });
    await page.locator('#btn-theme').click();
    await page.waitForTimeout(150);
    await page.screenshot({ path: path.join(shotDir, 't2-02-dark.png'), fullPage: true });
    pass('Screenshots written', shotDir);

  } finally {
    await browser.close();
  }

  console.log('');
  console.log(`Results: ${results.length - failed} passed, ${failed} failed, ${results.length} total`);
  const reportPath = path.join(__dirname, 'verify-report.json');
  fs.writeFileSync(reportPath, JSON.stringify({ failed, results, at: new Date().toISOString(), tier: 2 }, null, 2));
  console.log('Report:', reportPath);
  if (failed > 0) process.exit(1);
}

main().catch((e) => {
  console.error('Verification crashed:', e);
  process.exit(2);
});
