# Markdown Viewer Test Fixture

This document exercises every **Tier 1** feature of the viewer.

## Formatting

You can write *italic*, **bold**, ***both***, `inline code`, and ~~strikethrough~~.
Visit [Nous Research](https://nousresearch.com) for more.

### Nested Lists

- Top level item one
  - Nested child A
    - Deeply nested
  - Nested child B
- Top level item two
  1. Ordered sub one
  2. Ordered sub two

## Task List

- [x] Implement markdown parser
- [x] Add syntax highlighting
- [ ] Write the summary
- [ ] Profit

## Code Block

```javascript
function greet(name) {
  // print a friendly greeting
  const message = `Hello, ${name}!`;
  console.log(message);
  return message.length;
}
```

## Table

| Feature    | Tier | Status |
| ---------- | ---- | ------ |
| Tables     | 1    | OK     |
| Task list  | 1    | OK     |
| Highlights | 1    | OK     |

## Blockquote

> This is a blockquote.
> It can span multiple lines.
>
> > And nest inside another.

## Links and Images

An inline image reference: ![alt text](https://example.com/x.png)

### Heading for TOC anchor

Some trailing text to test search highlighting and the table of contents.
