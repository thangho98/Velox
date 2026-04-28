# Velox — Design System (V2)

Redesign of the Velox media dashboard as a premium streaming experience (Netflix/Max tier). Ultra-minimalist dark mode, zero borders, spacing and shadow as the sole separators, high-contrast typography.

---

## 1. Design Principles

1. **Pitch-black canvas.** The root background is `#000`. Everything sits on it.
2. **No borders.** Elements are separated by whitespace, elevation (shadow), and subtle surface-tint shifts — never by stroked outlines.
3. **High contrast type.** Pure white `#FFFFFF` for primary text; muted neutrals for secondary text. No mid-greys on mid-greys.
4. **Accent is sacred.** The crimson red (`#E50A1E`) is reserved for: the wordmark, primary CTAs, live indicators, progress fills, and "Up Next" eyebrows. Never decorative.
5. **Content before chrome.** Navigation is a single horizontal top bar; no left rail. Posters, stills, and channel marks fill the frame.
6. **Motion on intent.** Hover pops (channel cards scale 1.06), hero scrubs, play-button micro-scale. No idle animation beyond the live pulse.

---

## 2. Tailwind Color Palette

Registered in `tailwind.config` — used by the frontend.

```js
colors: {
  // Surfaces — layered near-blacks, no true greys
  ink: {
    0:   '#000000',  // canvas
    50:  '#0A0A0A',  // channel card body, footer surfaces
    100: '#141414',  // pill background (inactive category)
    200: '#1C1C1C',  // secondary button
    300: '#262626',  // secondary button hover
    400: '#2E2E2E',  // divider tint (use sparingly, never as a line)
  },
  // Text / neutrals — warm-cool neutral grey
  fog: {
    400: '#525252',  // dot separators, de-emphasized meta
    500: '#737373',  // secondary text, inactive tabs, captions
    600: '#A3A3A3',  // body copy, hover-target icons
    700: '#D4D4D4',  // hero body paragraph
    800: '#E5E5E5',  // white-button hover
  },
  // Brand accent — Velox crimson
  crimson: {
    500: '#E50A1E',  // wordmark, CTAs, LIVE badge, progress fill
    600: '#C80918',  // pressed / hover of crimson buttons
  },
}
```

### Usage rules
- **Backgrounds**: default `ink-0`. Nested surfaces step up to `ink-50`, `ink-100`, `ink-200` as elevation rises. Never more than three surface levels visible at once.
- **Text**: primary `white`, secondary `fog-700` / `fog-600`, tertiary `fog-500`, disabled/eyebrows `fog-500` + uppercase + tracking.
- **Accents**: `crimson-500` for < 3 hits per screen.
- **Do not** introduce blues, yellows, or saturated greys into chrome — those colors live only inside channel artwork tiles (brand-owned).

---

## 3. Typography

- **Family:** `Inter` (300–800), loaded from Google Fonts. System-ui fallback.
- **Wordmark:** 22 px, weight 800, letter-spacing `0.22em`, color `crimson-500`.
- **H1 (hero):** 72 px, weight 800, line-height `0.95`, tracking `-0.01em`.
- **H2 (section title):** 22–28 px, weight 600, tracking `-0.01em`.
- **Eyebrow:** 11 px, uppercase, letter-spacing `0.28em–0.32em`, color `fog-500` (or `crimson-500` for live/flags).
- **Body:** 13.5–15 px, weight 400, color `fog-700`.
- **Meta / captions:** 11–12 px, `fog-500`/`fog-600`.
- **Buttons:** 13–13.5 px, weight 500–600, wide tracking.

---

## 4. Spacing & Layout Grid

- **Max content width:** `1600px`, centered.
- **Horizontal gutter:** `40px` (`px-10`).
- **Top nav height:** `72px`, sticky, `bg-black/70` + backdrop-blur.
- **Section rhythm:** `56px` (`space-y-14`) between content sections.
- **Row carousel:** items in a horizontally scrolling flex row, gap `16px` (`gap-4`), full-bleed scroll area (negative margin) but content padded back to the gutter.
- **Hero (Home):** `78vh`, min `620px`. Content anchored bottom-left, 40px from bottom gutter.
- **Hero (Live TV):** fixed 21:9 aspect ratio at the top of the content column (not full-bleed) — signals "cinematic player" without taking over.

### Card dimensions
| Card | Dimensions |
|---|---|
| Poster (Movies / Series) | 180 × 260 px |
| Continue Watching still | 300 × 170 px |
| Next Up still | 340 × 190 px |
| Channel card | aspect-ratio 4/3 + 2-line footer (≈ 48 px) |
| On Now card | aspect-video + 5-line footer |

### Channel grid
- 6 columns ≥ `lg`, 4 ≥ `md`, 3 ≥ `sm`, 2 below.
- Gap: `20px` (`gap-5`).
- No borders; every card is a self-contained tinted tile.

---

## 5. Elevation (Shadows)

Borders are banned; depth comes from two shadow tokens:

```js
boxShadow: {
  card: '0 10px 30px -12px rgba(0,0,0,0.9)',
  pop:  '0 24px 60px -18px rgba(0,0,0,0.95), 0 0 0 1px rgba(255,255,255,0.04)',
}
```

- **`shadow-card`** — resting state for posters, stills, channel cards.
- **`shadow-pop`** — hovered or focused state. The `0 0 0 1px rgba(255,255,255,0.04)` is a *hairline inner glow*, not a border — it separates the card from adjacent black when raised.

---

## 6. Components

### Top Nav
- Sticky, translucent. Logo → tab group (Home · Movies · Series · **Live TV** · Browse) → right actions (Search, Bell, Avatar).
- Active tab: white text + 2 px white underline 18 px below baseline.
- Inactive tab: `fog-500`, hover `white`.
- **Live TV tab** carries a 6 px pulsing `crimson-500` dot beside the label.

### Channel Card (Live TV)
- Tinted tile in the channel's brand color (from data), white/brand typographic mark centered.
- **LIVE badge** top-left: black pill + pulsing crimson dot + uppercase tracked label.
- Footer in `ink-50`: channel name (white) + current program (`fog-500`).
- **Hover:** `transform: scale(1.06)`, `shadow-pop`, `z-index: 5`. Transition `260ms cubic-bezier(.2,.7,.2,1)`.

### Row (Carousel)
- Horizontal scroll, no visible scrollbar (`hide-scroll`).
- Chevron buttons fade in on row hover, positioned outside the gutter, scroll ±600 px on click.

### Buttons
- **Primary:** `bg-white text-black`, 44 px tall (`h-11`), font-weight 600, wide tracking. Hover → `fog-800`.
- **Secondary:** `bg-ink-200/80`, hover `ink-300`. Same dimensions. White text.
- **Icon-only:** same background, square (`w-11 h-11`), 18 px icon.
- **No border, no radius beyond 2 px** — buttons are crisp slabs to match the cinematic tone.

### Progress bars
- 2–3 px tall, `bg-white/10` track, `bg-crimson-500` fill (content progress) or `bg-white/90` fill (live scrubber).

### Live indicator
- 6 px `crimson-500` dot + `livepulse` keyframe (1.6 s ease-in-out, opacity 1↔0.5, scale 1↔0.85).

---

## 7. Imagery & Placeholders

No real posters ship in the redesign. Placeholder tiles use:
- A tinted near-black background (`poster-tint-0…6`, warm-cool variants).
- A 14 px diagonal hairline stripe pattern at `rgba(255,255,255,0.06)`.
- Centered monospace hint text ("Poster", "Still", "Key Art · 1920×1080").

Swap in real artwork by removing the `.poster` class and placing a full-bleed `<img>` or `background-image`.

---

## 9. Accessibility notes

- All primary text ≥ `fog-700` on black passes AAA.
- Live pulse is purely decorative; badges include the text "LIVE".
- Focus rings: default browser outline is kept; consider replacing with a 2 px white ring in follow-up work.
- Tab controls use `<button>` elements; keyboard order follows DOM order.
