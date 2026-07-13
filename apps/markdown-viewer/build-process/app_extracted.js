
"use strict";
/* ===================== APP STATE ===================== */
const $ = (id) => document.getElementById(id);
const fileInput = $('fileInput');
const content = $('content');
const tocList = $('tocList');
const tabsEl = $('tabs');

const store = {
  tabs: [],            // {name, text}
  active: -1,
  theme: 'light',
  tocOpen: true,
  rawMode: false,
};
const LS_KEY = 'mdv_state_v1';

/* ===================== PERSISTENCE ===================== */
function saveState() {
  try {
    const slim = {
      tabs: store.tabs.map(t => ({ name: t.name, text: t.text })),
      active: store.active, theme: store.theme, tocOpen: store.tocOpen, rawMode: store.rawMode,
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
    store.tabs = (s.tabs || []).filter(t => t && typeof t.text === 'string');
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
  store.tabs.forEach((t, i) => {
    const el = document.createElement('div');
    el.className = 'tab' + (i === store.active ? ' active' : '');
    const name = document.createElement('span');
    name.textContent = t.name;
    name.title = t.name;
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
  const existing = store.tabs.findIndex(t => t.name === name);
  if (existing >= 0) {
    store.tabs[existing].text = text;
    if (activate) store.active = existing;
  } else {
    store.tabs.push({ name, text });
    if (activate) store.active = store.tabs.length - 1;
  }
  renderTabs();
  if (activate) showActive();
  saveState();
  updateFileHint();
}

function closeTab(i) {
  store.tabs.splice(i, 1);
  if (store.active >= store.tabs.length) store.active = store.tabs.length - 1;
  if (store.active < 0) store.active = 0;
  renderTabs();
  showActive();
  saveState();
  updateFileHint();
}

function setActive(i) {
  store.active = i;
  renderTabs();
  showActive();
  saveState();
}

function showActive() {
  const t = store.tabs[store.active];
  if (!t) {
    content.innerHTML = '<div id="drop-hint"><h1>Drop a .md file here</h1><p>or click <strong>Open</strong> to choose a file.</p></div>';
    tocList.innerHTML = '';
    return;
  }
  if (store.rawMode) {
    showRaw(t.text);
  } else {
    renderMarkdown(t.text);
  }
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
function readFile(file) {
  const reader = new FileReader();
  reader.onload = () => openTab(file.name, String(reader.result || ''));
  reader.onerror = () => alert('Failed to read file: ' + file.name);
  reader.readAsText(file);
}

fileInput.addEventListener('change', (e) => {
  const f = e.target.files && e.target.files[0];
  if (f) readFile(f);
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
    for (const f of files) {
      if (/\.(md|markdown|txt)$/i.test(f.name) || f.type === 'text/markdown') readFile(f);
      else readFile(f); // try anyway
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
}

/* ===================== BOOT ===================== */
function boot() {
  const had = loadState();
  document.documentElement.setAttribute('data-theme', store.theme);
  $('themeBtn').textContent = store.theme === 'light' ? '🌙 Dark' : '☀️ Light';
  $('toc').classList.toggle('collapsed', !store.tocOpen);
  $('rawBtn').textContent = store.rawMode ? '👁 Rendered' : '📝 Raw';
  renderTabs();
  if (had) { showActive(); }
  updateFileHint();
}
boot();
