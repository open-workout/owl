# tools/playground/

A browser-based IDE for authoring OWL workout projects: a file tree,
tabbed editing, and up to three side-by-side panes. React + TypeScript
+ Vite.

## Current scope (MVP)

- **File tree** (left sidebar) — flat list of in-memory files, each an
  `.owl`, `.json`, or `.md` file. A new project starts **empty**; use
  "+ New File" to create files (extension determines kind). Files
  support rename (double-click) and delete.
- **Tabs** — clicking a file opens it in the focused pane. `.owl` and
  `.json` files open as a single source tab; `.md` files open as
  **two** tabs, "name.md" (raw source, editable) and "name.md
  (Preview)" (rendered markdown, via `react-markdown` + `remark-gfm`).
- **Split view** — 1 to 3 vertical panes ("Split" / "Close Pane" in
  each pane's tab bar). Any file can be opened independently in any
  pane.
- **Editing** — plain monospace `<textarea>` per file, held in React
  state (`src/state/WorkspaceContext.tsx`). **Nothing persists**: a
  reload loses all files. That's a deliberate v1 simplification, not
  an oversight.

**No compiler is wired in.** `.json` files are just files a user
pastes or types into — there is no "Compile" action turning an `.owl`
source into canonical JSON yet. See "Future work" below.

## Not built yet (visible stubs)

The top bar has two buttons that render disabled with a "Coming soon"
tooltip — they fix the UI shape for later work without functioning yet:

- **Publish** — will publish the current project to the (not yet
  built) marketplace.
- **Send to Phone** — will hand the current project to the
  OpenWorkout phone app via a QR code.

Neither has any backend, marketplace, or QR logic behind it today.

## Future work

- **Compilation.** Per the root [`CLAUDE.md`](../../CLAUDE.md) and
  [`reference/README.md`](../../reference/README.md), compilation runs
  server-side in Go — the intended shape here is a "Compile" action
  that calls a local `owlc` HTTP-server mode (or a thin dev-only
  endpoint), not an in-browser/WASM compiler. If client-side
  `Compile`/`Resolve` ever becomes a real requirement, that's what
  [`bindings/js/`](../../bindings/js/) is the placeholder for — nothing
  in this app imports it today.
- **Forking.** Opening a project tree "forked" from someone else's
  published program (once Publish/marketplace exist) — today, every
  project starts empty. The file/pane state (`src/types.ts`,
  `src/state/WorkspaceContext.tsx`) is a plain in-memory model, so
  seeding it from fetched files instead of starting empty is a small
  extension, not a redesign.
- **A real code editor.** The `.owl`/`.json` textarea has no syntax
  highlighting. CodeMirror or Monaco are the natural upgrades when
  that's worth the dependency weight.
- **Persistence.** Currently none — no `localStorage`, no backend save.

## Running locally

```sh
cd tools/playground
npm install
npm run dev        # dev server with hot reload
npm run typecheck  # tsc --noEmit
npm run build      # typecheck + production build to dist/
```
