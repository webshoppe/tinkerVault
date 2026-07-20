# Hermes Console: User Guide

Hermes Console lets you drive a running **Hermes** agent from your phone,
tablet, or any second device: send messages, watch responses, and approve or
deny approval-gated tool calls remotely (the reason it exists). It runs
entirely in your browser and talks straight to your Hermes server; there's no
account and no middle server.

---

## 1. Before you start

You need three things from your Hermes server:

| Field | What it is | Example |
|-------|-----------|---------|
| **Host** | The address of the machine running Hermes (often a tailnet name or LAN IP) | `hermes.my-tailnet.ts.net` or `192.168.1.20` |
| **Port** | The API port Hermes serves on | `8642` (the default) |
| **Bearer key** | The API key configured on that server | `sk-...` |

> The console never ships with a host, port, or key baked in; you fill these
> in once per device and they're saved locally in that browser.

---

## 2. Connecting

1. Open the console (see **Installing** below, or just visit the hosted URL).
2. On the **Connection** panel, enter Host, Port, and Bearer key.
3. Tap **Save & Connect**.

The dot in the top-left turns **green** when connected. If it fails:

- **Health check failed**, the host/port is wrong, or the server isn't
  reachable from this device (check you're on the same tailnet/LAN).
- **Capabilities request failed**, usually a wrong bearer key.

Your settings are remembered; next time you open the app it reconnects
automatically. Use **⚙ Settings** to change them, or **Clear stored config** to
wipe them from this device.

---

## 3. Chatting

Type in the box at the bottom and press **Send** (or **Enter**; **Shift+Enter**
makes a new line). The agent's reply appears above as it's produced. While a run
is active a **Stop** button appears; tap it to halt the run.

Hovering any message reveals a small toolbar in its corner: **Copy** copies
that message's raw text to your clipboard, and on your own messages
**Rerun** resubmits the exact same input as a brand-new run (useful for
retrying a prompt without retyping it). This toolbar currently only appears on
hover, so it's most reliable on desktop with a mouse; touch-only devices may
not reveal it the same way.

At the very bottom you'll see a small **"Powered by"** line showing which agent
/ model the server reported.

---

## 4. Approving tool calls (the main event)

When the agent wants to run a tool that requires your approval, the run pauses
and a highlighted **Approval required** card appears showing:

- the **tool name**, and
- the **arguments** it wants to run with.

Review it, then tap **Approve** or **Deny**. The decision is sent to the server
and the run continues (or stops, if denied). This works from any device the
console is open on; that's the whole point: you can be away from the machine
running Hermes and still clear its approvals.

---

## 5. Sessions

Tap **☰ Sessions** to open the sessions panel:

- **Pick a session** to load its full history into the chat and continue it;
  every message you send afterward is threaded into that conversation.
- **+ New session** starts a fresh, empty conversation.
- **Refresh** re-fetches the list.
- **⑂ Fork** (top bar, shown when a session is active) branches the current
  conversation into a new one so you can explore an alternate path without
  disturbing the original.

Without a selected session, your first message quietly starts a new one.

---

## 6. Installing (optional, recommended)

The console is a PWA, so you can install it like a native app:

- **Android (Chrome):** tap the **⬇ Install** button in the top bar (or
  Chrome's "Add to Home screen").
- **Desktop (Chrome/Edge):** tap **⬇ Install**, or use the install icon in the
  address bar.
- **iPhone/iPad (Safari):** use **Share → Add to Home Screen** (iOS doesn't
  offer an in-page install button; that's normal).

Once installed it opens in its own window and **loads offline**. Note: the app
*shell* works offline, but it still needs network access to your Hermes server
to actually do anything, offline just means the app itself launches instantly.

---

## 7. Troubleshooting

| Symptom | Likely cause / fix |
|--------|--------------------|
| Red dot, "Health check failed" | Wrong host/port, or server unreachable from this device. |
| Red dot, "Capabilities request failed" | Wrong bearer key. |
| Connected but can't type | Server didn't advertise `run_submission` in its capabilities. |
| No **Stop** button during a run | Server didn't advertise `run_stop`. |
| Response seems slow to update | Monitoring is by polling (~1.5 s), so updates arrive in small steps rather than a live stream, expected on this server (see the Dev Guide for why). |
| Approve/Deny didn't seem to register | Open the browser console; the full request and response are logged for every decision. |

---

## 8. Privacy

Everything runs in your browser. Your host, port, and bearer key are stored only
in this browser's `localStorage` and are sent only to the Hermes server you
point at. There is no analytics, no third-party server, and nothing phones home.
Use **Clear stored config** to remove your saved credentials from a device.
