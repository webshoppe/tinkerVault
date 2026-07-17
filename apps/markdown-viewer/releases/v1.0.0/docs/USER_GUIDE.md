# User Guide

This guide walks through every feature of the Markdown Viewer in the order a first-time user would meet them. It assumes no technical background; if you just want to read a Markdown file, start at the top.

The single file that matters is **`index.html`**. Double-click it to open the viewer in your browser. Everything below happens after a `.md` file is open.

---

## 1. Opening a file

When the page first opens, it shows a "Drop a .md file here" message. There are two ways to load a file:

- **File picker:** Click the **"Choose a .md file"** label in the top-left of the toolbar. A normal file window opens; pick any `.md` (or `.markdown` / `.txt`) file and it loads.
- **Drag and drop:** Drag a `.md` file from your computer and drop it anywhere on the page. It loads immediately, no dialog needed. You can drop several files and they each become a tab (see §7).

Once loaded, the document is formatted on the right, and a table of contents appears on the left.

---

## 2. What the formatted view shows

The viewer renders standard GitHub-Flavored Markdown correctly:

- **Headings** (the big titles) each get a small `#` link next to them. Click that link to copy a direct anchor to that heading, or just click the heading text in the table of contents (see §5).
- **Bold**, *italic*, ~~strikethrough~~, and `inline code` all display as you'd expect.
- **Tables** render with a proper header row and alternating shading.
- **Task lists** (checklists like `- [x]` and `- [ ]`) show real checkboxes; ticked ones are filled in, unticked ones are empty.
- **Nested lists** (bullets or numbers indented under other bullets) keep their structure.
- **Blockquotes** (lines starting with `>`) are shown as indented, bordered quotes, including nested quotes.
- **Links** are clickable.

---

## 3. Code blocks and the Copy button

Fenced code blocks (the ones between triple backticks, like ```` ```javascript ````) are displayed with **real syntax highlighting**, keywords, strings, and function names are colored, the way a code editor does it.

Each code block has a **Copy** button in its corner. Click it and the entire code block is copied to your clipboard, so you can paste it into an editor, terminal, or chat. The button briefly changes to "Copied!" to confirm.

---

## 4. Dark and light mode

The toolbar has a button on the right that shows either 🌙 (Dark) or ☀️ (Light). Click it to switch the whole interface between a light theme and a dark theme. Your choice is remembered the next time you open the tool.

---

## 5. Table of contents

When a file is open, a list of its headings appears in a panel on the left under the heading "CONTENTS". Click any entry to jump straight to that section.

If you want more reading room, click the **☰ TOC** button in the toolbar to collapse or expand that panel.

---

## 6. Search within the document

The search box (top-right of the toolbar, placeholder "Search in document…") highlights every match inside the open document.

- Type a word, matches are highlighted.
- Press **Enter** to jump to the next match; **Shift+Enter** goes to the previous one.
- A small "1/3" style counter shows which match you're on.
- Press **Escape** to clear the search.

---

## 7. Multiple files (tabs)

You can have several files open at once. Each opened file gets a **tab** at the top of the page showing its name, with a small `×` to close it.

- Open more files with the file picker or by dropping them.
- Click a tab to switch to that file.
- Click the `×` on a tab to close it.

---

## 8. Raw / rendered toggle

The **📝 Raw** button toggles between two views of the current file:

- **Rendered** (default): the formatted, readable view.
- **Raw**: the original Markdown text exactly as written, in a plain box.

Click the button again (it then reads "👁 Rendered") to return to the formatted view.

---

## 9. Export to a standalone HTML file

The **⬇ Export** button saves the currently open, formatted document as a single `.html` file on your computer. That exported file is fully self-contained; it keeps the formatting and syntax highlighting and can be opened in any browser with no dependencies. Use it to share a nicely formatted copy of a document.

---

## 10. It remembers your file

The viewer stores the last file you opened (and your theme and TOC preference) in your browser's local storage. The next time you open `index.html`, that file reloads automatically; you don't have to pick it again.

> Note: this "remembering" is local to your browser on your computer. It does not upload anything anywhere; nothing leaves your machine.

---

## Try it yourself

Open `test/test.md` (included in this release) to see tables, a code block, a task list, and nested headings all working together.
