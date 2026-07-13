# Philosophy & Built With

## Built on a real desk, not a cloud dev box

Everything in this repo was written, tested, and packaged on ordinary
consumer hardware, not rented cloud compute or a CI cluster:

- **Build machine:** Intel i7-7700 (2017-era desktop CPU), RTX 3060
  12GB, 32GB DDR4, Windows 10 IoT Enterprise LTSC
- **Same physical rig, isolated build tooling:** one app's build
  environment runs in its own WSL2 distro (Ubuntu 24.04) on that same
  machine, not a separate box, see below

No dev cloud, no rented GPU, no build farm. If you're running similar
homelab-tier hardware, this is what it's actually capable of.

## How each app actually got built

Different apps here were built through different local pipelines,
deliberately, as an experiment in what small self-hosted setups can
produce:

- **Markdown Viewer** was built by an LLM agent running through a
  locally-hosted gateway, using an external model provider as the
  backend, in an isolated profile with no access to git credentials
  or account-level auth of any kind. All git operations for the
  actual repo were done by a human, separately.
- **Whiteboard** was built through an agentic coding CLI running
  inside an isolated WSL2 distro (Ubuntu 24.04) on that same rig,
  driven interactively one feature tier at a time, with independent
  manual verification in a real browser after every automated test
  pass.

## Verification discipline

The one rule that mattered more than any tool choice: **automated
"tests passed" is not the same as verified working.** One internal
build shipped a fully broken drag-and-drop interaction with a clean
automated test report, because the test moved application state
directly instead of exercising real pointer events. It looked done.
It wasn't. A human clicking the actual feature in a real browser
caught it; nothing else would have.

Every release in this repo gets checked in more than one browser
engine before being called finished, and "the build succeeded" is
never treated as equivalent to "the feature works."

## Why single-file, offline-first apps

No CDN, no account, no server, no install. That constraint was chosen
on purpose: an app built this way still works exactly the same in
five years, off a USB stick, with nothing to set up and nothing that
can silently stop working because a service somewhere went away.

## Multi-agent, one human in the loop

Different apps here were built by different tools and different
models, but the process was the same every time: an agent drafts and
runs autonomously, and a human verifies the actual result before
anything ships. The tool changes. That discipline doesn't.
