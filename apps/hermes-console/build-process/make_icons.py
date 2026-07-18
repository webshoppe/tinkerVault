#!/usr/bin/env python3
"""
Hermes Console PWA icon generator.

Draws a polished, original mark: a rounded tile with a vertical accent gradient,
a subtle inner highlight, and a custom geometric "H" whose crossbar carries small
upswept wings — a nod to Hermes (messenger / speed) without copying any logo.
Renders at 4x then downsamples for crisp antialiased edges.

Outputs: icon-192, icon-512, icon-maskable-512, apple-touch-icon(180).
"""
from PIL import Image, ImageDraw

# Palette (matches app UI)
BG      = (20, 22, 26)      # --bg   #14161a
TILE_TOP = (91, 154, 255)   # lighter accent
TILE_BOT = (58, 111, 216)   # --accent-2 #3a6fd8
GLYPH    = (17, 19, 23)      # near-bg, sits on the tile
GLYPH_HI = (233, 240, 255)   # optional light glyph (unused; kept for tweaking)

SS = 4  # supersample factor


def vgrad(size, top, bot):
    """Vertical gradient image."""
    base = Image.new("RGB", (1, size))
    px = base.load()
    for y in range(size):
        t = y / max(1, size - 1)
        px[0, y] = (
            round(top[0] + (bot[0] - top[0]) * t),
            round(top[1] + (bot[1] - top[1]) * t),
            round(top[2] + (bot[2] - top[2]) * t),
        )
    return base.resize((size, size))


def rounded_mask(size, radius):
    m = Image.new("L", (size, size), 0)
    d = ImageDraw.Draw(m)
    d.rounded_rectangle((0, 0, size - 1, size - 1), radius=radius, fill=255)
    return m


def draw_h_with_wings(d, cx, cy, h, w, stroke, color):
    """
    Custom H: two vertical posts + a crossbar, with two small wings sweeping
    up-and-out from the crossbar ends. Coordinates are around center (cx, cy).
    h = full glyph height, w = distance between post centers.
    """
    half_h = h / 2
    half_w = w / 2
    top = cy - half_h
    bot = cy + half_h
    lx = cx - half_w
    rx = cx + half_w
    r = stroke / 2

    # Vertical posts (rounded caps).
    d.rounded_rectangle((lx - r, top, lx + r, bot), radius=r, fill=color)
    d.rounded_rectangle((rx - r, top, rx + r, bot), radius=r, fill=color)
    # Crossbar.
    d.rounded_rectangle((lx, cy - r, rx, cy + r), radius=r, fill=color)

    # Wings: short strokes rising from each crossbar end, angled outward.
    wing_len = w * 0.52
    wing_rise = h * 0.20
    ww = stroke * 0.72
    # left wing goes up-left, right wing up-right
    d.line((lx, cy, lx - wing_len, cy - wing_rise), fill=color, width=round(ww))
    d.line((rx, cy, rx + wing_len, cy - wing_rise), fill=color, width=round(ww))
    # small feather ticks for a bit of character
    d.line((lx - wing_len * 0.55, cy - wing_rise * 0.42,
            lx - wing_len * 0.95, cy - wing_rise * 0.30),
           fill=color, width=round(ww * 0.7))
    d.line((rx + wing_len * 0.55, cy - wing_rise * 0.42,
            rx + wing_len * 0.95, cy - wing_rise * 0.30),
           fill=color, width=round(ww * 0.7))
    # round the wing tips/joins
    for (px, py) in [(lx, cy), (rx, cy)]:
        d.ellipse((px - ww/2, py - ww/2, px + ww/2, py + ww/2), fill=color)


def make(size, maskable, path):
    S = size * SS
    img = Image.new("RGBA", (S, S), (0, 0, 0, 0))

    # Background: on maskable, full-bleed dark; on normal, dark canvas too so
    # rounded corners read on any launcher background.
    canvas = Image.new("RGBA", (S, S), BG + (255,))

    # Tile inset: smaller inset (bigger tile) for normal; larger safe-zone
    # padding for maskable so the glyph stays within Android's ~80% safe area.
    inset = int(S * (0.16 if maskable else 0.07))
    tile_size = S - 2 * inset
    radius = int(tile_size * 0.24)

    # Build the gradient tile.
    grad = vgrad(tile_size, TILE_TOP, TILE_BOT).convert("RGBA")
    tmask = rounded_mask(tile_size, radius)

    # Inner top highlight for a subtle glossy feel.
    hi = Image.new("RGBA", (tile_size, tile_size), (0, 0, 0, 0))
    hd = ImageDraw.Draw(hi)
    hd.ellipse((-tile_size*0.3, -tile_size*0.85, tile_size*1.3, tile_size*0.35),
               fill=(255, 255, 255, 40))
    grad = Image.alpha_composite(grad, hi)

    canvas.paste(grad, (inset, inset), tmask)

    # Draw the glyph centered on the tile.
    d = ImageDraw.Draw(canvas)
    cx = cy = S / 2
    gh = tile_size * 0.46
    gw = tile_size * 0.40
    stroke = tile_size * 0.115
    draw_h_with_wings(d, cx, cy, gh, gw, stroke, GLYPH)

    # For non-maskable, knock out the area outside the tile so the icon has
    # transparent corners (nicer on light backgrounds / desktop).
    if not maskable:
        final = Image.new("RGBA", (S, S), (0, 0, 0, 0))
        full_mask = Image.new("L", (S, S), 0)
        fd = ImageDraw.Draw(full_mask)
        fd.rounded_rectangle((inset, inset, S - inset - 1, S - inset - 1),
                             radius=radius, fill=255)
        final.paste(canvas, (0, 0), full_mask)
        canvas = final

    out = canvas.resize((size, size), Image.LANCZOS)
    out.save(path)
    print("wrote", path, out.size)


make(192, False, "icon-192.png")
make(512, False, "icon-512.png")
make(512, True,  "icon-maskable-512.png")
make(180, False, "apple-touch-icon.png")
