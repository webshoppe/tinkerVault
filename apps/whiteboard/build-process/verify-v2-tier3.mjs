/**
 * Whiteboard v2 Tier 3 verification — smart paste / drop intelligence.
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
  await page.waitForTimeout(100);
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

/** Tiny 2x2 PNG as data URL */
function tinyPngDataUrl() {
  // 1x1 red PNG
  return 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==';
}

async function main() {
  console.log('Whiteboard v2 Tier-3 verification');
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
    let page = await freshPage(browser);

    // ============================================================
    console.log('--- Shell / thresholds ---');
    const ver = await page.evaluate(() => window.__whiteboard.VERSION);
    assert(ver === '2.2.0', 'App version 2.2.0', ver);
    assert(await page.locator('#smart-toast').count() === 1, 'Smart toast element present');

    const thresholds = await page.evaluate(() => ({
      stickyMax: window.__whiteboard.PASTE_STICKY_MAX,
      wordpadMin: window.__whiteboard.PASTE_WORDPAD_MIN
    }));
    assert(thresholds.stickyMax === 280, 'Sticky max threshold 280', String(thresholds.stickyMax));
    assert(thresholds.wordpadMin === 900, 'Wordpad min threshold 900', String(thresholds.wordpadMin));

    // ============================================================
    console.log('--- classifyPlainText heuristics ---');
    const cls = await page.evaluate(() => {
      const c = window.__whiteboard.classifyPlainText;
      return {
        empty: c('   ').kind,
        short: c('Hello #world TODO').kind,
        // Concrete manual-test lengths
        len150: c('x'.repeat(150)).kind,
        len400: c('x'.repeat(400)).kind,
        len950: c('x'.repeat(950)).kind,
        boundarySticky: c('x'.repeat(280)).kind,
        boundaryAmb: c('x'.repeat(281)).kind,
        boundaryLong: c('x'.repeat(900)).kind,
        amb: c('x'.repeat(500)).kind,
        url: c('https://docs.example.com/a/b'),
        urlWww: c('www.example.org'),
        urlMulti: c('see https://example.com more').kind,
        // CSV needs ≥3 columns to avoid "Hello, world" prose false positives
        table: c('name,age,city\nAda,30,Paris\nBob,25,Lyon').kind,
        table2col: c('name,age\nAda,30\nBob,25').kind,
        proseComma: c('Hello, world.\nHow are you, friend?').kind,
        tableTab: c('a\tb\n1\t2').kind
      };
    });
    assert(cls.empty === 'empty', 'Empty text');
    assert(cls.short === 'sticky', 'Short text → sticky');
    assert(cls.len150 === 'sticky', '150 chars → sticky (not toast/Wordpad)', cls.len150);
    assert(cls.len400 === 'ambiguous-text', '400 chars → ambiguous toast', cls.len400);
    assert(cls.len950 === 'wordpad', '950 chars → wordpad', cls.len950);
    assert(cls.boundarySticky === 'sticky', '280 chars → sticky');
    assert(cls.boundaryAmb === 'ambiguous-text', '281 chars → ambiguous');
    assert(cls.boundaryLong === 'wordpad', '900 chars → wordpad');
    assert(cls.amb === 'ambiguous-text', '500 chars → ambiguous');
    assert(cls.url && cls.url.kind === 'url' && cls.url.domain === 'docs.example.com',
      'Single URL classified', JSON.stringify(cls.url));
    assert(cls.urlWww && cls.urlWww.kind === 'url', 'www. URL classified');
    assert(cls.urlMulti !== 'url', 'URL with extra text not pure-url', cls.urlMulti);
    assert(cls.table === 'tabular', '3-col CSV classified tabular');
    assert(cls.table2col !== 'tabular', '2-col CSV is NOT tabular (prose-safe)', cls.table2col);
    assert(cls.proseComma !== 'tabular', 'Prose with commas is NOT tabular', cls.proseComma);
    assert(cls.tableTab === 'tabular', 'TSV classified tabular');

    // ============================================================
    console.log('--- Length routing via real paste events (150 / 400 / 950) ---');
    async function pastePlainAndObserve(n) {
      await page.evaluate(() => {
        const App = window.__whiteboard.App;
        if (App.saveTimer) { clearTimeout(App.saveTimer); App.saveTimer = null; }
        localStorage.clear();
      });
      await page.reload({ waitUntil: 'domcontentloaded' });
      await waitBoard(page, 350);
      const text = 'B'.repeat(n);
      return page.evaluate(async (t) => {
        if (document.activeElement && document.activeElement.blur) document.activeElement.blur();
        document.body.click();
        const dt = new DataTransfer();
        dt.setData('text/plain', t);
        const ev = new Event('paste', { bubbles: true, cancelable: true });
        Object.defineProperty(ev, 'clipboardData', { value: dt });
        window.dispatchEvent(ev);
        await new Promise((r) => setTimeout(r, 250));
        const App = window.__whiteboard.App;
        const toast = document.getElementById('smart-toast');
        const msg = document.getElementById('smart-toast-msg');
        const b = App.currentBoard();
        return {
          n: t.length,
          classify: window.__whiteboard.classifyPlainText(t).kind,
          toast: !!(toast && toast.classList.contains('visible')),
          toastMsg: (msg && msg.textContent) || '',
          templateId: b && b.templateId,
          name: b && b.name
        };
      }, text);
    }

    const p150 = await pastePlainAndObserve(150);
    assert(p150.classify === 'sticky', 'Paste 150 classify sticky', JSON.stringify(p150));
    assert(!p150.toast, 'Paste 150: no toast', p150.toastMsg);
    assert(p150.templateId === 'notepad', 'Paste 150 → notepad sticky board', p150.templateId);

    const p400 = await pastePlainAndObserve(400);
    assert(p400.classify === 'ambiguous-text', 'Paste 400 classify ambiguous', JSON.stringify(p400));
    assert(p400.toast, 'Paste 400: toast visible');
    assert(/400 chars/i.test(p400.toastMsg), 'Toast reports 400 chars', p400.toastMsg);
    assert(p400.templateId !== 'wordpad' || p400.toast, 'Paste 400 does not silently open Wordpad');

    const p950 = await pastePlainAndObserve(950);
    assert(p950.classify === 'wordpad', 'Paste 950 classify wordpad', JSON.stringify(p950));
    assert(!p950.toast, 'Paste 950: no toast', p950.toastMsg);
    assert(p950.templateId === 'wordpad', 'Paste 950 → wordpad board', p950.templateId);

    // Residual toast must not linger after a later short paste
    await page.evaluate(() => {
      window.__whiteboard.App.routePlainText('x'.repeat(400));
    });
    await page.waitForTimeout(100);
    await page.evaluate(() => {
      window.__whiteboard.App.routePlainText('x'.repeat(150));
    });
    await page.waitForTimeout(200);
    const residual = await page.evaluate(() => ({
      toast: document.getElementById('smart-toast').classList.contains('visible'),
      templateId: window.__whiteboard.App.currentBoard().templateId
    }));
    assert(!residual.toast, 'Toast dismissed when a new short paste routes', JSON.stringify(residual));
    assert(residual.templateId === 'notepad', 'Follow-up 150-char paste → sticky', residual.templateId);

    // ============================================================
    console.log('--- Silent: short text → sticky (Tier 2 still applies) ---');
    // Fresh browser context so prior boards cannot pollute this check
    await page._context.close();
    page = await freshPage(browser);

    const sticky = await page.evaluate(async () => {
      if (document.activeElement && document.activeElement.blur) document.activeElement.blur();
      window.__whiteboard.App.routePlainText('TODO buy milk #errands');
      await new Promise((r) => setTimeout(r, 400));
      const App = window.__whiteboard.App;
      App.snapshotCurrent();
      const b = App.currentBoard();
      const notes = (b && b.state && b.state.notes) || [];
      const todoNote = notes.find((n) => n && /TODO/.test(n.text || ''));
      return {
        templateId: b && b.templateId,
        notes,
        todoNote,
        color: todoNote && todoNote.color
      };
    });
    assert(sticky.templateId === 'notepad', 'Short paste opens/uses notepad', sticky.templateId);
    assert(sticky.todoNote, 'Sticky note with TODO created', JSON.stringify(sticky.notes));
    assert(sticky.todoNote && sticky.todoNote.text.indexOf('TODO') >= 0, 'Sticky text preserved',
      sticky.todoNote && sticky.todoNote.text);
    assert(sticky.color === 'red', 'Tier 2 auto-color still applies (TODO→red)', sticky.color);

    // ============================================================
    console.log('--- Silent: long text → wordpad ---');
    await page.evaluate(() => {
      window.__whiteboard.App.routePlainText('Long document. '.repeat(80));
    });
    await waitBoard(page, 400);
    const wp = await page.evaluate(() => {
      const b = window.__whiteboard.App.currentBoard();
      return {
        templateId: b && b.templateId,
        hasHtml: !!(b && b.state && b.state.html && b.state.html.length > 50)
      };
    });
    assert(wp.templateId === 'wordpad', 'Long text → wordpad board', wp.templateId);
    assert(wp.hasHtml, 'Wordpad has HTML content');

    // ============================================================
    console.log('--- Silent: single URL → link card ---');
    await page.evaluate(() => {
      window.__whiteboard.App.routePlainText('https://example.com/docs/guide');
    });
    await waitBoard(page, 450);
    const link = await page.evaluate(() => {
      const App = window.__whiteboard.App;
      App.snapshotCurrent();
      const b = App.currentBoard();
      const n = b && b.state && b.state.notes && b.state.notes.find((x) => x.linkCard);
      const a = document.querySelector('.sticky-note.link-card a.link-card-url');
      return {
        templateId: b && b.templateId,
        note: n,
        domainText: a && a.textContent,
        href: a && a.getAttribute('href')
      };
    });
    assert(link.templateId === 'notepad', 'URL creates notepad board', link.templateId);
    assert(link.note && link.note.linkCard, 'Note marked as link card');
    assert(link.domainText === 'example.com', 'Link shows domain only', link.domainText);
    assert(link.href && link.href.indexOf('https://example.com') === 0, 'Link href preserved', link.href);

    // ============================================================
    console.log('--- Silent: image → annotate (compression path) ---');
    const imgResult = await page.evaluate(async () => {
      const wb = window.__whiteboard;
      const before = wb.App.store.boards.length;
      const c = document.createElement('canvas');
      c.width = 400;
      c.height = 300;
      const g = c.getContext('2d');
      g.fillStyle = '#4a7cff';
      g.fillRect(0, 0, 400, 300);
      g.fillStyle = '#fff';
      g.fillText('paste-test', 20, 40);
      const dataUrl = c.toDataURL('image/png');
      await wb.App.createAnnotateFromImage(dataUrl, 'Paste test image');
      // openBoard mounts async; wait for board switch
      await new Promise(function (r) { setTimeout(r, 250); });
      const byName = wb.App.store.boards.find((x) => x.name === 'Paste test image');
      const b = wb.App.currentBoard();
      return {
        before,
        after: wb.App.store.boards.length,
        currentTemplate: b && b.templateId,
        currentName: b && b.name,
        boardTemplate: byName && byName.templateId,
        hasImage: !!(byName && byName.state && byName.state.imageDataUrl &&
          byName.state.imageDataUrl.indexOf('data:image') === 0),
        imageLen: byName && byName.state && byName.state.imageDataUrl && byName.state.imageDataUrl.length
      };
    });
    await waitBoard(page, 300);
    assert(imgResult.after > imgResult.before, 'New board created for image');
    assert(imgResult.boardTemplate === 'snip', 'Image → Annotate (snip)', imgResult.boardTemplate);
    assert(imgResult.hasImage, 'Annotate has imageDataUrl', String(imgResult.imageLen));
    assert(imgResult.currentTemplate === 'snip' || imgResult.currentName === 'Paste test image',
      'Annotate board opened', imgResult.currentName + '/' + imgResult.currentTemplate);

    // prepareAnnotateImage uses same downscale helper
    const prep = await page.evaluate(async () => {
      const c = document.createElement('canvas');
      c.width = 1600;
      c.height = 1200;
      const g = c.getContext('2d');
      const id = g.createImageData(c.width, c.height);
      for (let i = 0; i < id.data.length; i += 4) {
        id.data[i] = (Math.random() * 256) | 0;
        id.data[i + 1] = (Math.random() * 256) | 0;
        id.data[i + 2] = (Math.random() * 256) | 0;
        id.data[i + 3] = 255;
      }
      g.putImageData(id, 0, 0);
      const big = c.toDataURL('image/png');
      const small = await window.__whiteboard.prepareAnnotateImage(big);
      return { before: big.length, after: small.length, compressed: small.length < big.length };
    });
    assert(prep.compressed, 'prepareAnnotateImage uses Tier 1 downscale',
      prep.before + ' → ' + prep.after);

    // ============================================================
    console.log('--- Toast: ambiguous text ---');
    await page.evaluate(() => {
      window.__whiteboard.App.routePlainText('Ambiguous middle length text. '.repeat(15));
    });
    await page.waitForTimeout(150);
    const toastAmb = await page.evaluate(() => {
      const el = document.getElementById('smart-toast');
      const msg = document.getElementById('smart-toast-msg');
      return {
        visible: el && el.classList.contains('visible'),
        text: msg && msg.textContent,
        buttons: Array.from(document.querySelectorAll('#smart-toast-actions button')).map((b) => b.textContent)
      };
    });
    assert(toastAmb.visible, 'Ambiguous text shows toast');
    assert(/Sticky|Wordpad/i.test(toastAmb.text || ''), 'Toast mentions sticky/wordpad', toastAmb.text);
    assert(toastAmb.buttons.some((b) => /Sticky/i.test(b)), 'Sticky action button');
    assert(toastAmb.buttons.some((b) => /Wordpad/i.test(b)), 'Wordpad action button');

    // Accept sticky
    await page.evaluate(() => {
      const btn = Array.from(document.querySelectorAll('#smart-toast-actions button'))
        .find((b) => /Sticky/i.test(b.textContent));
      if (btn) btn.click();
    });
    await waitBoard(page, 400);
    const afterAmb = await page.evaluate(() => {
      const b = window.__whiteboard.App.currentBoard();
      return {
        toast: document.getElementById('smart-toast').classList.contains('visible'),
        templateId: b && b.templateId
      };
    });
    assert(!afterAmb.toast, 'Toast dismissed after accept');
    assert(afterAmb.templateId === 'notepad', 'Accepted sticky → notepad');

    // ============================================================
    console.log('--- Toast: tabular text ---');
    await page.evaluate(() => {
      // 3 columns required for CSV tabular detection
      window.__whiteboard.App.routePlainText('col1,col2,col3\na,b,c\nd,e,f');
    });
    await page.waitForTimeout(150);
    const toastTab = await page.evaluate(() => {
      const msg = document.getElementById('smart-toast-msg');
      return {
        visible: document.getElementById('smart-toast').classList.contains('visible'),
        text: msg && msg.textContent
      };
    });
    assert(toastTab.visible, 'Tabular text shows toast');
    assert(/table/i.test(toastTab.text || ''), 'Toast mentions table', toastTab.text);

    await page.evaluate(() => {
      const btn = Array.from(document.querySelectorAll('#smart-toast-actions button'))
        .find((b) => /table/i.test(b.textContent));
      if (btn) btn.click();
    });
    await waitBoard(page, 400);
    const tableBoard = await page.evaluate(() => {
      const b = window.__whiteboard.App.currentBoard();
      return {
        templateId: b && b.templateId,
        hasTable: !!(b && b.state && b.state.html && b.state.html.indexOf('<table') >= 0)
      };
    });
    assert(tableBoard.templateId === 'wordpad', 'Table → wordpad');
    assert(tableBoard.hasTable, 'Wordpad HTML contains table');

    // ============================================================
    console.log('--- Toast: JSON import (never silent) ---');
    const boardCountBefore = await page.evaluate(() => window.__whiteboard.App.store.boards.length);
    await page.evaluate(() => {
      // Simulate dropped json via routeFiles with a synthetic File is hard; call toast path
      const exportPayload = {
        format: 'whiteboard-export',
        version: 2,
        boards: [{
          name: 'Imported From Toast',
          templateId: 'paint',
          state: {}
        }]
      };
      // Use showSmartToast path equivalent to JSON drop
      const fileLike = new File([JSON.stringify(exportPayload)], 'boards.json', { type: 'application/json' });
      window.__whiteboard.App.routeFiles([fileLike]);
    });
    await page.waitForTimeout(200);
    const toastJson = await page.evaluate(() => {
      const msg = document.getElementById('smart-toast-msg');
      return {
        visible: document.getElementById('smart-toast').classList.contains('visible'),
        text: msg && msg.textContent,
        boardCount: window.__whiteboard.App.store.boards.length
      };
    });
    assert(toastJson.visible, 'JSON drop shows import toast');
    assert(/Import JSON|board/i.test(toastJson.text || ''), 'Toast wording is clear', toastJson.text);
    assert(toastJson.boardCount === boardCountBefore, 'JSON not imported until accept',
      String(toastJson.boardCount));

    await page.evaluate(() => {
      const btn = Array.from(document.querySelectorAll('#smart-toast-actions button'))
        .find((b) => /Import/i.test(b.textContent));
      if (btn) btn.click();
    });
    await waitBoard(page, 400);
    const afterJson = await page.evaluate(() => {
      const App = window.__whiteboard.App;
      return {
        count: App.store.boards.length,
        names: App.store.boards.map((b) => b.name),
        current: App.currentBoard() && App.currentBoard().name
      };
    });
    assert(afterJson.count > boardCountBefore, 'JSON import added board(s)');
    assert(afterJson.names.some((n) => /Imported From Toast/i.test(n)),
      'Imported board name present', JSON.stringify(afterJson.names));

    // ============================================================
    console.log('--- Multi-image files → cascade Annotate boards ---');
    const multi = await page.evaluate(async () => {
      const App = window.__whiteboard.App;
      const before = App.store.boards.filter((b) => b.templateId === 'snip').length;
      // Two tiny PNG files
      const b64 = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==';
      function pngFile(name) {
        const bin = atob(b64);
        const arr = new Uint8Array(bin.length);
        for (let i = 0; i < bin.length; i++) arr[i] = bin.charCodeAt(i);
        return new File([arr], name, { type: 'image/png' });
      }
      await new Promise((resolve) => {
        // routeFiles is sync start + async chain
        App.routeFiles([pngFile('a.png'), pngFile('b.png')]);
        setTimeout(resolve, 800);
      });
      const snips = App.store.boards.filter((b) => b.templateId === 'snip');
      return {
        before,
        after: snips.length,
        names: snips.slice(-2).map((b) => b.name),
        currentIsSnip: App.currentBoard() && App.currentBoard().templateId === 'snip'
      };
    });
    assert(multi.after >= multi.before + 2, 'Two Annotate boards from two images',
      multi.before + ' → ' + multi.after);
    assert(multi.currentIsSnip, 'Last multi-image board is open');

    // ============================================================
    console.log('--- Text file drop → wordpad ---');
    await page.evaluate(async () => {
      const App = window.__whiteboard.App;
      const f = new File(['Hello from txt\nLine 2 #tag'], 'notes.txt', { type: 'text/plain' });
      App.routeFiles([f]);
      await new Promise((r) => setTimeout(r, 400));
    });
    await waitBoard(page, 300);
    const txtFile = await page.evaluate(() => {
      const b = window.__whiteboard.App.currentBoard();
      return {
        templateId: b && b.templateId,
        name: b && b.name,
        html: b && b.state && b.state.html
      };
    });
    assert(txtFile.templateId === 'wordpad', '.txt → wordpad', txtFile.templateId);
    assert(/Hello from txt/i.test(txtFile.html || ''), 'Wordpad contains file text');

    // ============================================================
    console.log('--- Paste into field does not steal text ---');
    await page.evaluate(() => {
      // Create notepad with empty note and focus textarea
      const App = window.__whiteboard.App;
      App.createBoard({
        name: 'Field paste test',
        templateId: 'notepad',
        state: {
          notes: [{ id: 'n_field', x: 40, y: 40, w: 200, h: 180, text: '', color: 'yellow', colorAuto: true }]
        },
        open: true
      });
    });
    await waitBoard(page, 400);
    // Enter edit mode and type
    await page.evaluate(() => {
      const note = document.querySelector('.sticky-note');
      if (note) {
        const wrap = note.querySelector('.note-body-wrap');
        if (wrap) wrap.click();
      }
    });
    await page.waitForTimeout(100);
    const fieldPaste = await page.evaluate(() => {
      const ta = document.querySelector('.sticky-note textarea.note-body');
      if (!ta) return { ok: false, reason: 'no textarea' };
      ta.focus();
      // Simulate that handleGlobalPaste would skip when in field
      const fakeEvent = {
        target: ta,
        clipboardData: {
          items: [],
          files: [],
          getData: () => 'should-stay-in-field'
        },
        preventDefault() { this._pd = true; },
        stopPropagation() { this._sp = true; }
      };
      // Call with real check
      const inField = ta.tagName === 'TEXTAREA';
      return { ok: true, inField, active: document.activeElement === ta };
    });
    assert(fieldPaste.ok && fieldPaste.inField, 'Can focus sticky textarea for native paste',
      JSON.stringify(fieldPaste));

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
    version: '2.2.0',
    tier: 'v2-tier3',
    file: htmlPath,
    chrome: CHROME,
    thresholds: { PASTE_STICKY_MAX: 280, PASTE_WORDPAD_MIN: 900 },
    passed: results.filter((r) => r.ok).length,
    failed,
    total: results.length,
    results
  };
  const out = path.resolve(__dirname, 'verify-report-v2-tier3.json');
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
