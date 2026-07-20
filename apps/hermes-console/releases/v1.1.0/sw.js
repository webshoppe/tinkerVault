/* Hermes Console: service worker (Phase 4)
 *
 * Strategy:
 *   - App shell (same-origin static assets): cache-first, so the console
 *     opens offline / instantly. The shell is just the static HTML/CSS/JS +
 *     manifest + icons served from this origin (e.g. GitHub Pages).
 *   - API traffic (the Hermes server: Runs API + Sessions API): NETWORK-ONLY,
 *     never cached. That server lives at a different origin (a tailnet host:port
 *     the user configures at runtime), so every cross-origin request is treated
 *     as live data and passed straight through. Stale runs/sessions would be
 *     dangerous, so they must never be served from cache.
 *
 * Cache is versioned: bump SHELL_VERSION on every deploy so clients drop the
 * old shell instead of getting stuck on it.
 *
 * No push notifications (settled non-goal: public push infra conflicts with
 * the tailnet-only constraint).
 */

const SHELL_VERSION = "hermes-console-shell-v3";

// Same-origin shell assets to precache. Paths are relative to the SW scope
// (/apps/hermes-console/), so they resolve correctly under a subpath deploy.
const SHELL_ASSETS = [
  "./",
  "./index.html",
  "./manifest.json",
  "./icon-192.png",
  "./icon-512.png",
  "./icon-maskable-512.png",
  "./apple-touch-icon.png",
  "./favicon.ico",
  "./favicon-32.png",
  "./favicon-16.png",
  "./VERSION",
];

// ---- Install: precache the shell ----
self.addEventListener("install", (event) => {
  event.waitUntil(
    caches.open(SHELL_VERSION).then((cache) => cache.addAll(SHELL_ASSETS))
      .then(() => self.skipWaiting())
  );
});

// ---- Activate: drop any older shell caches ----
self.addEventListener("activate", (event) => {
  event.waitUntil(
    caches.keys().then((keys) =>
      Promise.all(
        keys.filter((k) => k !== SHELL_VERSION).map((k) => caches.delete(k))
      )
    ).then(() => self.clients.claim())
  );
});

// ---- Fetch routing ----
self.addEventListener("fetch", (event) => {
  const req = event.request;

  // Only handle GET; never interfere with POST/PATCH/DELETE (all API writes).
  if (req.method !== "GET") return;

  const url = new URL(req.url);

  // Cross-origin => the Hermes API server (or any other host). NETWORK-ONLY.
  // Do not cache, do not fall back to cache: live tailnet data.
  if (url.origin !== self.location.origin) {
    return; // let the browser handle it directly (no respondWith = default net)
  }

  // Same-origin: app shell. Cache-first, fall back to network, then cache the
  // fresh copy for next time.
  event.respondWith(
    caches.match(req).then((cached) => {
      if (cached) return cached;
      return fetch(req).then((res) => {
        // Only cache good, basic (same-origin) responses.
        if (res && res.status === 200 && res.type === "basic") {
          const copy = res.clone();
          caches.open(SHELL_VERSION).then((cache) => cache.put(req, copy));
        }
        return res;
      }).catch(() => {
        // Offline and not in cache: for navigations, serve the shell.
        if (req.mode === "navigate") return caches.match("./index.html");
        return Response.error();
      });
    })
  );
});
