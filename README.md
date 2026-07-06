# Kinetix — Go / Ebitengine port

An implementation of **Kinetix** (from *Code the Classics Volume 2*, Raspberry Pi
Press) built on [Ebitengine](https://ebitengine.org), provided as a side-by-side
counterpart to [`kinetix-go`](https://github.com/chrplr/kinetix-go), which uses
[go-sdl3](https://github.com/Zyko0/go-sdl3) + the
[pgzgo](https://github.com/chrplr/pgzgo) harness.

**▶ Play it in your browser: <https://chrplr.github.io/kinetix-go-ebitengine/>**

## Why two versions?

Both repositories run the **same game** on two different pure-Go game stacks, so you
can compare the stacks directly. The entire gameplay is shared *verbatim* — these 13
files are byte-identical between the two repos:

```
game.go  ball.go  bat.go  barrel.go  bullet.go  brickcollide.go  controls.go
actor.go  impact.go  constants.go  vec.go  util.go  rng.go
```

Only the **backend** differs:

| | This repo (Ebitengine) | [`kinetix-go`](https://github.com/chrplr/kinetix-go) (go-sdl3 + pgzgo) |
|---|---|---|
| Entry point | `main.go` | `main.go` |
| Backend glue | `ebiten_adapter.go` (in-repo shim) | `harness.go` + the external `pgzgo` module |
| Graphics / audio / input | Ebitengine (pure Go engine) | SDL3 via `purego` bindings |
| CGo | none | none |
| System libraries | none | SDL3/SDL3_image/SDL3_mixer, bundled & extracted at runtime |
| WebAssembly | **yes** (see below) | not out of the box |

Ebitengine is a batteries-included, pure-Go engine (its own game loop, image,
audio and input), so `ebiten_adapter.go` is a thin shim exposing the same
`Assets`/audio/input API that the shared game code expects. `kinetix-go` instead
wraps the SDL3 C libraries through `purego` (no CGo — the libraries are bundled and
extracted at startup), with `pgzgo` supplying the game loop and helpers. Neither
version needs a system SDL install.

Diffing the two `main.go` + glue files is the quickest way to see exactly what each
stack asks of you.

## Run (native)

```sh
go run .
```

## Controls

Keyboard (arrow keys / `A`,`D` to steer, `Space` to launch/fire) or a standard
gamepad — identical to `kinetix-go`.

## WebAssembly (play in a browser)

Ebitengine compiles to WebAssembly, so the game runs in a browser. The loader
(`index.html`) and Go's JS support shim (`wasm_exec.js`) are included; only the wasm
binary needs building — it is a build artifact and is not committed.

```sh
# 1. Compile the game to WebAssembly
GOOS=js GOARCH=wasm go build -o kinetix.wasm .

# 2. Serve this directory over HTTP — a wasm module cannot be loaded from a
#    file:// URL, so opening index.html directly will not work.
python3 -m http.server 8000
```

Then open <http://localhost:8000>. Any static HTTP server works (`python3 -m
http.server`, `npx serve`, …); the only requirement is that the files are served
over HTTP.

If you upgrade your Go toolchain, refresh the shim so it matches the compiler:

```sh
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" .
```

You don't have to build it to share it, though: every push to `main` compiles the
wasm and deploys `index.html` + `wasm_exec.js` + `kinetix.wasm` to GitHub Pages via
[`.github/workflows/pages.yml`](.github/workflows/pages.yml), so the binary never
needs to live in the repository.

## Ebitengine vs. go-sdl3 + pgzgo

This game exists as two repositories that share their gameplay code verbatim and
differ only in the backend: **[kinetix-go](https://github.com/chrplr/kinetix-go)**
uses go-sdl3 with the [pgzgo](https://github.com/chrplr/pgzgo) harness, and
**[kinetix-go-ebitengine](https://github.com/chrplr/kinetix-go-ebitengine)** uses
[Ebitengine](https://ebitengine.org) (and is playable
[in your browser](https://chrplr.github.io/kinetix-go-ebitengine/)). Comparing them
is really a comparison of the two Go game stacks. Where they differ:

| Dimension | Comes out ahead | Why |
|-----------|-----------------|-----|
| Web / mobile reach | **Ebitengine** | Compiles to WebAssembly (see the live demo) plus iOS/Android; SDL3-via-purego has no real wasm story. |
| Build & dependencies | **Ebitengine** | Pure Go — `go build` just works and cross-compiles cleanly. go-sdl3 bundles SDL's C libraries and extracts them at runtime. |
| Maturity & ecosystem | **Ebitengine** | Years old, widely used and documented; go-sdl3 is a young (v0.1.x) binding over battle-tested SDL. |
| Built-in audio | **go-sdl3 + pgzgo** | SDL3_mixer offers looping tracks, gain and fades out of the box; Ebitengine's audio is lower-level. |
| Low-level control | **go-sdl3 + pgzgo** | Direct renderer, blend-mode and clip-rect access, with a small harness you fully own. |
| Headless testing | **go-sdl3 + pgzgo** | SDL's dummy driver runs the game with no display, so CI can `-selftest` the real loop. |

**Bottom line:** for something you want to ship or share, Ebitengine is the more
pragmatic default — its maturity and one-command web/mobile builds are hard to beat.
Reach for go-sdl3 + pgzgo when you want low-level SDL control, richer built-in audio,
or a minimal, transparent stack you own end to end.

Two things this pair shows beyond the scorecard:

1. **The engine choice barely touches the game.** Only `main.go` and the backend
   glue differ — the 13 gameplay files are identical. Keeping game logic
   engine-agnostic (plain structs behind an assets/audio/input seam) lets you swap
   backends later.
2. **The two APIs converged.** The Ebitengine adapter is thin because both stacks
   descend from the same `update()` / `draw()` / `screen` loop (Pygame Zero → pgzgo;
   Ebitengine's `ebiten.Game`). Good 2D game APIs keep rediscovering the same shape.

## Provenance & license

Ported to Go from the Python original in *Code the Classics Volume 2*. The game
design and original assets are © their respective authors / Raspberry Pi Press.

The Go source code of this port is distributed under the MIT License.
