
"use strict";
/* ===================== APP STATE ===================== */
const $ = (id) => document.getElementById(id);
const fileInput = $('fileInput');
const content = $('content');
const tocList = $('tocList');
const tabsEl = $('tabs');

const store = {
  tabs: [],            // {id, name, text, scroll}  — id is a unique per-load key (NOT the filename)
  active: -1,
  theme: 'light',
  tocOpen: true,
  rawMode: false,
  editMode: false,
};
const LS_KEY = 'mdv_state_v1';
// Stack of recently-closed tabs for "reopen last closed" (Ctrl/Cmd+Shift+T).
let lastClosed = [];

// Unique per-load identifier for a tab. File pickers/drag-drop only expose the
// basename (no folder), so two different files can share a name. We therefore
// key tabs by an id that is unique for every load, never by filename alone —
// otherwise two distinct README.md files would collide into one tab.
let __tabSeq = 0;
function genTabId() {
  if (window.crypto && typeof crypto.randomUUID === 'function') return 'tab-' + crypto.randomUUID();
  __tabSeq += 1;
  return 'tab-' + Date.now().toString(36) + '-' + __tabSeq + '-' + Math.random().toString(36).slice(2, 8);
}

/* ===================== PERSISTENCE ===================== */
function saveState() {
  try {
    const slim = {
      tabs: store.tabs.map(t => ({ name: t.name, text: t.text, scroll: t.scroll || 0 })),
      active: store.active, theme: store.theme, tocOpen: store.tocOpen, rawMode: store.rawMode, editMode: store.editMode,
    };
    localStorage.setItem(LS_KEY, JSON.stringify(slim));
  } catch (e) { /* quota — ignore */ }
}
function loadState() {
  try {
    const raw = localStorage.getItem(LS_KEY);
    if (!raw) return false;
    const s = JSON.parse(raw);
    store.theme = s.theme || 'light';
    store.tocOpen = s.tocOpen !== false;
    store.rawMode = !!s.rawMode;
    store.editMode = !!s.editMode;
    store.tabs = (s.tabs || [])
      .filter(t => t && typeof t.text === 'string')
      .map(t => Object.assign({ id: genTabId(), scroll: 0 }, t));
    store.active = Math.min(Math.max(s.active || 0, 0), Math.max(store.tabs.length - 1, 0));
    return store.tabs.length > 0;
  } catch (e) { return false; }
}

/* ===================== MARKED CONFIG ===================== */
marked.setOptions({ gfm: true, breaks: false, headerIds: false, mangle: false });

function slugify(text) {
  return text.toLowerCase().trim()
    .replace(/[^\w\s-]/g, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-') || 'section';
}

/* ===================== RENDER ===================== */
function renderMarkdown(mdText) {
  const html = marked.parse(mdText);
  content.innerHTML = html;
  applyHighlights();
  addCopyButtons();
  addAnchors();
  buildTOC();
  clearSearch();
}

function applyHighlights() {
  content.querySelectorAll('pre code').forEach((block) => {
    try { hljs.highlightElement(block); } catch (e) {}
  });
}

function addCopyButtons() {
  content.querySelectorAll('pre').forEach((pre) => {
    if (pre.querySelector('.copy-btn')) return;
    const btn = document.createElement('button');
    btn.className = 'copy-btn';
    btn.type = 'button';
    btn.textContent = 'Copy';
    const code = pre.querySelector('code');
    const text = code ? code.textContent : pre.textContent;
    btn.addEventListener('click', () => copyText(text, btn));
    pre.appendChild(btn);
  });
}

function addAnchors() {
  const seen = {};
  content.querySelectorAll('h1,h2,h3,h4,h5,h6').forEach((h) => {
    const base = slugify(h.textContent);
    let id = base;
    if (seen[base] != null) { seen[base] += 1; id = base + '-' + seen[base]; }
    else { seen[base] = 0; }
    h.id = id;
    const a = document.createElement('a');
    a.className = 'anchor';
    a.href = '#' + id;
    a.textContent = '#';
    a.title = 'Link to this section';
    h.insertBefore(a, h.firstChild);
  });
}

function buildTOC() {
  tocList.innerHTML = '';
  const heads = content.querySelectorAll('h1,h2,h3,h4,h5,h6');
  if (!heads.length) {
    const li = document.createElement('li');
    li.innerHTML = '<span style="color:var(--fg-muted)">No headings</span>';
    tocList.appendChild(li);
    return;
  }
  heads.forEach((h) => {
    const li = document.createElement('li');
    li.className = 'lvl-' + h.tagName[1];
    const a = document.createElement('a');
    a.href = '#' + h.id;
    a.textContent = h.textContent.replace(/^#\s*/, '');
    a.addEventListener('click', (e) => {
      e.preventDefault();
      h.scrollIntoView({ behavior: 'smooth', block: 'start' });
      history.replaceState(null, '', '#' + h.id);
    });
    li.appendChild(a);
    tocList.appendChild(li);
  });
}

/* ===================== CLIPBOARD ===================== */
async function copyText(text, btn) {
  let ok = false;
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text);
      ok = true;
    }
  } catch (e) {}
  if (!ok) {
    try {
      const ta = document.createElement('textarea');
      ta.value = text;
      ta.style.position = 'fixed';
      ta.style.opacity = '0';
      document.body.appendChild(ta);
      ta.select();
      ok = document.execCommand('copy');
      ta.remove();
    } catch (e) {}
  }
  if (btn) {
    const old = btn.textContent;
    btn.textContent = ok ? 'Copied!' : 'Failed';
    btn.classList.toggle('copied', ok);
    setTimeout(() => { btn.textContent = old; btn.classList.remove('copied'); }, 1400);
  }
  return ok;
}

/* ===================== TABS ===================== */
function renderTabs() {
  tabsEl.innerHTML = '';
  // Build a display label per tab. When two tabs share a filename (e.g. two
  // README.md from different folders), append a numeric suffix so the user can
  // tell them apart. Labels are derived purely for display; tab identity is the
  // stable id, never the filename (see openTab).
  const counts = {};
  const labels = store.tabs.map(t => {
    const base = t.name || 'untitled';
    counts[base] = (counts[base] || 0) + 1;
    return counts[base] > 1 ? base + ' (' + counts[base] + ')' : base;
  });
  store.tabs.forEach((t, i) => {
    const el = document.createElement('div');
    el.className = 'tab' + (i === store.active ? ' active' : '');
    const label = labels[i];
    const name = document.createElement('span');
    name.className = 'name';
    name.textContent = label;
    name.title = t.name || 'untitled';
    el.appendChild(name);
    const close = document.createElement('span');
    close.className = 'close';
    close.textContent = '×';
    close.title = 'Close tab';
    close.addEventListener('click', (e) => { e.stopPropagation(); closeTab(i); });
    el.appendChild(close);
    el.addEventListener('click', () => { setActive(i); });
    tabsEl.appendChild(el);
  });
  tabsEl.style.display = store.tabs.length ? 'flex' : 'none';
}

function openTab(name, text, activate = true) {
  store.tabs.push({ id: genTabId(), name: name, text: text, scroll: 0 });
  if (activate) store.active = store.tabs.length - 1;
  renderTabs();
  if (activate) showActive();
  saveState();
  updateFileHint();
}

function saveScroll() {
  const t = store.tabs[store.active];
  if (t) t.scroll = window.scrollY || 0;
}
function restoreScroll() {
  const t = store.tabs[store.active];
  window.scrollTo(0, t && t.scroll ? t.scroll : 0);
}

function closeTab(i) {
  saveScroll();
  const removed = store.tabs[i];
  // Remember closed tabs (most-recent first) for "reopen last closed".
  if (removed) lastClosed.unshift({ name: removed.name, text: removed.text, scroll: removed.scroll || 0 });
  if (lastClosed.length > 25) lastClosed.length = 25;
  store.tabs.splice(i, 1);
  if (store.active >= store.tabs.length) store.active = store.tabs.length - 1;
  if (store.active < 0) store.active = 0;
  renderTabs();
  showActive();
  saveState();
  updateFileHint();
}

function reopenLastClosed() {
  if (!lastClosed.length) { toast('No recently closed tabs'); return; }
  const t = lastClosed.shift();
  store.tabs.push({ id: genTabId(), name: t.name, text: t.text, scroll: t.scroll || 0 });
  const idx = store.tabs.length - 1;
  store.active = idx;
  renderTabs();
  showActive();
  saveState();
  updateFileHint();
}

function setActive(i) {
  saveScroll();
  store.active = i;
  renderTabs();
  showActive();
  updateTitle();   // refresh the browser tab title to the now-active filename
  saveState();
}

function showActive() {
  const t = store.tabs[store.active];
  if (!t) {
    content.innerHTML = '<div id="drop-hint"><h1>Drop a .md file here</h1><p>or click <strong>Open</strong> to choose a file.</p></div>';
    tocList.innerHTML = '';
    return;
  }
  if (store.editMode) {
    // Edit mode: show the source textarea, keep the (last-rendered) content hidden.
    const ed = $('edit-area');
    ed.value = t.text;
    if (store.rawMode) { showRaw(t.text); } else { renderMarkdown(t.text); }
    restoreScroll();
    return;
  }
  if (store.rawMode) {
    showRaw(t.text);
  } else {
    renderMarkdown(t.text);
  }
  restoreScroll();
}

function showRaw(text) {
  content.innerHTML = '';
  const pre = document.createElement('pre');
  pre.id = 'raw-view';
  pre.textContent = text;
  content.appendChild(pre);
  tocList.innerHTML = '';
}

/* ===================== SEARCH ===================== */
let searchMarks = [];
let searchIdx = -1;
function clearSearchMarks() {
  searchMarks.forEach(m => {
    const parent = m.parentNode;
    if (parent) parent.replaceChild(document.createTextNode(m.textContent), m);
    parent && parent.normalize && parent.normalize();
  });
  searchMarks = [];
  searchIdx = -1;
}
function clearSearch() {
  clearSearchMarks();
  $('searchHint').textContent = '';
  $('searchInput').value = '';
}
function runSearch() {
  const q = $('searchInput').value.trim();
  clearSearchMarks();
  if (!q) { $('searchHint').textContent = ''; return; }
  const walker = document.createTreeWalker(content, NodeFilter.SHOW_TEXT, {
    acceptNode(n) {
      if (!n.nodeValue.trim()) return NodeFilter.FILTER_REJECT;
      if (n.parentNode && (n.parentNode.closest('pre') || n.parentNode.closest('style') || n.parentNode.closest('script'))) return NodeFilter.FILTER_REJECT;
      return NodeFilter.FILTER_ACCEPT;
    }
  });
  const targets = [];
  let n;
  while ((n = walker.nextNode())) targets.push(n);
  const lower = q.toLowerCase();
  targets.forEach(node => {
    const text = node.nodeValue;
    const idx = text.toLowerCase().indexOf(lower);
    if (idx === -1) return;
    const frag = document.createDocumentFragment();
    let last = 0, i = idx;
    while (i !== -1) {
      if (i > last) frag.appendChild(document.createTextNode(text.slice(last, i)));
      const mark = document.createElement('mark');
      mark.className = 'hl';
      mark.textContent = text.slice(i, i + q.length);
      frag.appendChild(mark);
      searchMarks.push(mark);
      last = i + q.length;
      i = text.toLowerCase().indexOf(lower, last);
    }
    if (last < text.length) frag.appendChild(document.createTextNode(text.slice(last)));
    node.parentNode.replaceChild(frag, node);
  });
  if (searchMarks.length) {
    searchIdx = 0;
    searchMarks[0].classList.add('current');
    searchMarks[0].scrollIntoView({ block: 'center' });
    $('searchHint').textContent = (searchIdx + 1) + '/' + searchMarks.length;
  } else {
    $('searchHint').textContent = '0/0';
  }
}
function searchNext(dir) {
  if (!searchMarks.length) return;
  searchMarks[searchIdx].classList.remove('current');
  searchIdx = (searchIdx + dir + searchMarks.length) % searchMarks.length;
  searchMarks[searchIdx].classList.add('current');
  searchMarks[searchIdx].scrollIntoView({ block: 'center' });
  $('searchHint').textContent = (searchIdx + 1) + '/' + searchMarks.length;
}

/* ===================== SAVE ACTIVE TAB AS .md ===================== */
// Client-side only: build a Blob from the active tab's (possibly edited) text
// and trigger a download via a programmatic anchor click. No server, no File
// System Access API (unreliable under file://). This always creates a NEW file;
// it cannot overwrite the original (file:// gives us no write path to it), so we
// default the download name to the tab's current filename unchanged.
function saveActiveFile() {
  const t = store.tabs[store.active];
  if (!t) { toast('Open a file first.'); return; }
  const blob = new Blob([t.text], { type: 'text/markdown;charset=utf-8' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = t.name || 'document.md';
  document.body.appendChild(a);
  a.click();
  a.remove();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
  toast('Saved a copy: ' + (t.name || 'document.md'));
}

/* ===================== EXPORT ===================== */
function exportStandalone() {
  const t = store.tabs[store.active];
  if (!t) { alert('Open a file first.'); return; }
  // Build a self-contained HTML with current rendered content + current theme hljs css.
  const dark = store.theme === 'dark';
  const hljsCss = dark ? `html[data-theme="dark"] pre code.hljs{display:block;overflow-x:auto;padding:1em}html[data-theme="dark"] code.hljs{padding:3px 5px}html[data-theme="dark"] /*!
  Theme: GitHub Dark
  Description: Dark theme as seen on github.com
  Author: github.com
  Maintainer: @Hirse
  Updated: 2021-05-15

  Outdated base version: https://github.com/primer/github-syntax-dark
  Current colors taken from GitHub's CSS
*/.hljs{color:#c9d1d9;background:#0d1117}html[data-theme="dark"] .hljs-doctag,.hljs-keyword,.hljs-meta .hljs-keyword,.hljs-template-tag,.hljs-template-variable,.hljs-type,.hljs-variable.language_{color:#ff7b72}html[data-theme="dark"] .hljs-title,.hljs-title.class_,.hljs-title.class_.inherited__,.hljs-title.function_{color:#d2a8ff}html[data-theme="dark"] .hljs-attr,.hljs-attribute,.hljs-literal,.hljs-meta,.hljs-number,.hljs-operator,.hljs-selector-attr,.hljs-selector-class,.hljs-selector-id,.hljs-variable{color:#79c0ff}html[data-theme="dark"] .hljs-meta .hljs-string,.hljs-regexp,.hljs-string{color:#a5d6ff}html[data-theme="dark"] .hljs-built_in,.hljs-symbol{color:#ffa657}html[data-theme="dark"] .hljs-code,.hljs-comment,.hljs-formula{color:#8b949e}html[data-theme="dark"] .hljs-name,.hljs-quote,.hljs-selector-pseudo,.hljs-selector-tag{color:#7ee787}html[data-theme="dark"] .hljs-subst{color:#c9d1d9}html[data-theme="dark"] .hljs-section{color:#1f6feb;font-weight:700}html[data-theme="dark"] .hljs-bullet{color:#f2cc60}html[data-theme="dark"] .hljs-emphasis{color:#c9d1d9;font-style:italic}html[data-theme="dark"] .hljs-strong{color:#c9d1d9;font-weight:700}html[data-theme="dark"] .hljs-addition{color:#aff5b4;background-color:#033a16}html[data-theme="dark"] .hljs-deletion{color:#ffdcd7;background-color:#67060c}` : `pre code.hljs{display:block;overflow-x:auto;padding:1em}code.hljs{padding:3px 5px}/*!
  Theme: GitHub
  Description: Light theme as seen on github.com
  Author: github.com
  Maintainer: @Hirse
  Updated: 2021-05-15

  Outdated base version: https://github.com/primer/github-syntax-light
  Current colors taken from GitHub's CSS
*/.hljs{color:#24292e;background:#fff}.hljs-doctag,.hljs-keyword,.hljs-meta .hljs-keyword,.hljs-template-tag,.hljs-template-variable,.hljs-type,.hljs-variable.language_{color:#d73a49}.hljs-title,.hljs-title.class_,.hljs-title.class_.inherited__,.hljs-title.function_{color:#6f42c1}.hljs-attr,.hljs-attribute,.hljs-literal,.hljs-meta,.hljs-number,.hljs-operator,.hljs-selector-attr,.hljs-selector-class,.hljs-selector-id,.hljs-variable{color:#005cc5}.hljs-meta .hljs-string,.hljs-regexp,.hljs-string{color:#032f62}.hljs-built_in,.hljs-symbol{color:#e36209}.hljs-code,.hljs-comment,.hljs-formula{color:#6a737d}.hljs-name,.hljs-quote,.hljs-selector-pseudo,.hljs-selector-tag{color:#22863a}.hljs-subst{color:#24292e}.hljs-section{color:#005cc5;font-weight:700}.hljs-bullet{color:#735c0f}.hljs-emphasis{color:#24292e;font-style:italic}.hljs-strong{color:#24292e;font-weight:700}.hljs-addition{color:#22863a;background-color:#f0fff4}.hljs-deletion{color:#b31d28;background-color:#ffeef0}`;
  const bodyHtml = content.innerHTML;
  const doc = `<!DOCTYPE html>
<html lang="en" data-theme="${store.theme}">
<head>
<meta charset="utf-8">
<title>${escapeHtml(t.name)}</title>
<style>
:root{--bg:#fff;--bg-elev:#f6f8fa;--fg:#1f2328;--fg-muted:#656d76;--border:#d0d7de;--accent:#0969da;--code-bg:#f6f8fa;--table-stripe:#f6f8fa;--blockquote-border:#d0d7de;}
html[data-theme="dark"]{--bg:#0d1117;--bg-elev:#161b22;--fg:#e6edf3;--fg-muted:#8b949e;--border:#30363d;--accent:#58a6ff;--code-bg:#161b22;--table-stripe:#161b22;--blockquote-border:#30363d;}
body{background:var(--bg);color:var(--fg);font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Helvetica,Arial,sans-serif;font-size:16px;line-height:1.6;margin:0;}
main{max-width:900px;margin:0 auto;padding:24px 32px 80px;}
h1,h2,h3,h4,h5,h6{margin:24px 0 16px;font-weight:600;line-height:1.25;} h1{font-size:2em;border-bottom:1px solid var(--border);padding-bottom:.3em;} h2{font-size:1.5em;border-bottom:1px solid var(--border);padding-bottom:.3em;} h3{font-size:1.25em;} h4{font-size:1em;} h5{font-size:.875em;} h6{font-size:.85em;color:var(--fg-muted);}
a{color:var(--accent);text-decoration:none;} a:hover{text-decoration:underline;}
img{max-width:100%;} hr{border:none;border-top:1px solid var(--border);margin:24px 0;}
blockquote{margin:0 0 16px;padding:0 1em;color:var(--fg-muted);border-left:.25em solid var(--blockquote-border);}
ul,ol{margin:0 0 16px;padding-left:2em;} li{margin:4px 0;} li.task-list-item{list-style:none;margin-left:-1.4em;} li.task-list-item input{margin-right:6px;}
table{border-collapse:collapse;margin:0 0 16px;display:block;width:max-content;max-width:100%;overflow:auto;} th,td{border:1px solid var(--border);padding:6px 13px;} th{background:var(--bg-elev);font-weight:600;} tr:nth-child(2n) td{background:var(--table-stripe);}
code{background:var(--code-bg);border-radius:6px;padding:.2em .4em;font-size:85%;font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;}
pre{background:var(--code-bg);border:1px solid var(--border);border-radius:8px;padding:14px 16px;overflow:auto;margin:0 0 16px;} pre code{background:transparent;padding:0;font-size:14px;display:block;}
${hljsCss}
</style>
</head>
<body>
<main>
${bodyHtml}
</main>
</body>
</html>`;
  const blob = new Blob([doc], { type: 'text/html' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = t.name.replace(/\.(md|markdown|txt)$/i, '') + '.html';
  document.body.appendChild(a);
  a.click();
  a.remove();
  setTimeout(() => URL.revokeObjectURL(url), 2000);
}
function escapeHtml(s){return s.replace(/[&<>"]/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;'}[c]));}

/* ===================== FILE LOADING ===================== */
let __toastTimer = null;
function toast(msg) {
  const el = document.getElementById('toast');
  if (!el) return;
  el.textContent = msg;
  el.classList.add('show');
  clearTimeout(__toastTimer);
  __toastTimer = setTimeout(() => el.classList.remove('show'), 2600);
}
function readFile(file, activate) {
  // Only treat text-like files as markdown. Dropping an image, PDF, or other
  // binary would otherwise render garbage; warn and skip instead.
  const isTextLike = /\.(md|markdown|txt)$/i.test(file.name) || file.type === 'text/markdown' || (file.type || '').startsWith('text/');
  if (!isTextLike) {
    toast('Skipped "' + file.name + '": not a Markdown/text file.');
    return;
  }
  const reader = new FileReader();
  reader.onload = () => openTab(file.name, String(reader.result || ''), activate !== false);
  reader.onerror = () => toast('Failed to read file: ' + file.name);
  reader.readAsText(file);
}

fileInput.addEventListener('change', (e) => {
  const files = e.target.files;
  if (files && files.length) {
    // Activate only the last opened tab; earlier ones open but stay inactive.
    for (let k = 0; k < files.length; k++) {
      readFile(files[k], k === files.length - 1);
    }
  }
  fileInput.value = '';
});

/* drag and drop anywhere */
function preventDefaults(e){ e.preventDefault(); e.stopPropagation(); }
['dragenter','dragover','dragleave','drop'].forEach(ev => {
  window.addEventListener(ev, preventDefaults, false);
});
let dragDepth = 0;
window.addEventListener('dragenter', () => { dragDepth++; $('drop-hint') && ($('drop-hint').classList && $('drop-hint').classList.add('dragover')); document.body.style.outline = '3px dashed var(--accent)'; });
window.addEventListener('dragleave', () => { dragDepth = Math.max(0, dragDepth - 1); if (dragDepth === 0) { document.body.style.outline = ''; } });
window.addEventListener('drop', (e) => {
  dragDepth = 0; document.body.style.outline = '';
  const files = e.dataTransfer && e.dataTransfer.files;
  if (files && files.length) {
    // Each dropped file opens as its own tab; activate only the last one.
    // readFile() itself filters out non-text files.
    for (let k = 0; k < files.length; k++) {
      readFile(files[k], k === files.length - 1);
    }
  }
});

/* ===================== TOOLBAR ACTIONS ===================== */
$('themeBtn').addEventListener('click', () => {
  store.theme = store.theme === 'light' ? 'dark' : 'light';
  document.documentElement.setAttribute('data-theme', store.theme);
  $('themeBtn').textContent = store.theme === 'light' ? '🌙 Dark' : '☀️ Light';
  saveState();
});

$('tocBtn').addEventListener('click', () => {
  store.tocOpen = !store.tocOpen;
  $('toc').classList.toggle('collapsed', !store.tocOpen);
  saveState();
});

$('rawBtn').addEventListener('click', () => {
  store.rawMode = !store.rawMode;
  $('rawBtn').textContent = store.rawMode ? '👁 Rendered' : '📝 Raw';
  showActive();
  saveState();
});

$('editBtn').addEventListener('click', () => {
  store.editMode = !store.editMode;
  document.body.classList.toggle('edit-mode', store.editMode);
  $('editBtn').textContent = store.editMode ? '👁 View' : '✏️ Edit';
  if (store.editMode) {
    const t = store.tabs[store.active];
    $('edit-area').value = t ? t.text : '';
    $('edit-area').focus();
  } else {
    // Leaving edit mode: re-render the (possibly edited) source.
    showActive();
  }
  saveState();
});

// Live re-render as the user edits the active tab's Markdown.
$('edit-area').addEventListener('input', () => {
  const t = store.tabs[store.active];
  if (!t) return;
  t.text = $('edit-area').value;
  // Re-render a hidden preview? We keep content hidden in edit mode; just store
  // the text and re-render content so switching back / exporting is current.
  renderMarkdown(t.text);
  saveState();
});

$('copyBtn').addEventListener('click', () => {
  const t = store.tabs[store.active];
  if (!t) { toast('Open a file first.'); return; }
  copyText(t.text, $('copyBtn'));
});

$('saveBtn').addEventListener('click', saveActiveFile);

$('exportBtn').addEventListener('click', exportStandalone);

let searchTimer = null;
$('searchInput').addEventListener('input', () => {
  clearTimeout(searchTimer);
  searchTimer = setTimeout(runSearch, 150);
});
$('searchInput').addEventListener('keydown', (e) => {
  if (e.key === 'Enter') { e.preventDefault(); searchNext(e.shiftKey ? -1 : 1); }
  if (e.key === 'Escape') { clearSearch(); }
});

function updateFileHint() {
  const t = store.tabs[store.active];
  $('fileHint').textContent = t ? ('📄 ' + t.name) : '';
  updateTitle();
}

// Browser tab title = the active tab's filename, nothing else. document.title
// updates are fully supported under file://, so a per-tab dynamic title is
// reliably achievable (no static fallback needed).
function updateTitle() {
  const t = store.tabs[store.active];
  document.title = t ? t.name : 'Markdown Viewer';
}

/* ===================== INTERNAL .md LINKS ===================== */
// Intercept clicks on same-origin relative links ending in .md/.markdown/.txt
// and load the target into the existing multi-file tab system instead of
// letting the browser navigate to raw, unformatted text.
function resolvePath(href) {
  const url = new URL(href, window.location.href);
  if (url.origin !== window.location.origin) return null;
  const clean = url.pathname.split('#')[0].split('?')[0];
  if (!/\.(md|markdown|txt)$/i.test(clean)) return null;
  return url;
}

function fetchMd(href, linkEl) {
  fetch(href)
    .then(r => { if (!r.ok) throw new Error('HTTP ' + r.status); return r.text(); })
    .then(text => openTab(decodeURIComponent(href.split('/').pop().split('#')[0]) || 'document.md', text))
    .catch(err => {
      // Fetch is blocked in many file:// contexts (e.g. Chrome) or failed.
      // Never fall through to raw-text navigation; show an on-page message.
      showLinkError(linkEl, href, err);
    });
}

function showLinkError(linkEl, href, err) {
  const msg = document.createElement('div');
  msg.className = 'link-error';
  msg.setAttribute('role', 'alert');
  msg.innerHTML = '⚠ Could not open <code>' + escapeHtml(href) + '</code> directly.<br>' +
    'Use the <strong>Open</strong> button or drag the file into the viewer instead.';
  const host = document.getElementById('link-error-host');
  if (host) { host.innerHTML = ''; host.appendChild(msg); host.style.display = 'block'; }
}

document.addEventListener('click', (e) => {
  if (e.defaultPrevented || e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;
  const a = e.target.closest('a');
  if (!a) return;
  const href = a.getAttribute('href');
  if (!href || href.charAt(0) === '#') return; // in-page anchors: leave alone
  let target;
  try { target = resolvePath(href); } catch (_) { return; } // absolute/other -> default nav
  if (!target) return; // external http(s) or non-md -> default nav
  e.preventDefault();
  fetchMd(target.href, a);
}, true);

/* ===================== KEYBOARD SHORTCUTS ===================== */
// Ctrl/Cmd+T  open file picker
// Ctrl/Cmd+W  close active tab
// Ctrl/Cmd+Shift+T  reopen last closed tab
// Ctrl/Cmd+Alt+Left/Right  (or Ctrl/Cmd+PageUp/PageDown)  switch tabs
document.addEventListener('keydown', (e) => {
  const mod = e.ctrlKey || e.metaKey;
  if (!mod) return;
  const k = e.key.toLowerCase();
  if (k === 't' && !e.shiftKey && !e.altKey) {
    e.preventDefault(); fileInput.click(); return;
  }
  if (k === 'w' && !e.shiftKey && !e.altKey) {
    e.preventDefault(); if (store.tabs.length) closeTab(store.active); return;
  }
  if (k === 't' && e.shiftKey && !e.altKey) {
    e.preventDefault(); reopenLastClosed(); return;
  }
  if (e.altKey && (e.key === 'ArrowLeft' || e.key === 'ArrowRight')) {
    e.preventDefault();
    const delta = e.key === 'ArrowLeft' ? -1 : 1;
    const n = store.tabs.length;
    if (n) setActive((store.active + delta + n) % n);
    return;
  }
  if ((k === 'pagedown' || k === 'pageup') && !e.altKey) {
    e.preventDefault();
    const delta = k === 'pagedown' ? 1 : -1;
    const n = store.tabs.length;
    if (n) setActive((store.active + delta + n) % n);
    return;
  }
});

/* ===================== BOOT ===================== */
function boot() {
  const had = loadState();
  document.documentElement.setAttribute('data-theme', store.theme);
  $('themeBtn').textContent = store.theme === 'light' ? '🌙 Dark' : '☀️ Light';
  $('toc').classList.toggle('collapsed', !store.tocOpen);
  $('rawBtn').textContent = store.rawMode ? '👁 Rendered' : '📝 Raw';
  $('editBtn').textContent = store.editMode ? '👁 View' : '✏️ Edit';
  document.body.classList.toggle('edit-mode', store.editMode);
  renderTabs();
  if (had) { showActive(); }
  updateFileHint();
}
boot();
