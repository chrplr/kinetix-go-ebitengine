# Kinetix — Go / Ebitengine port

An implementation of **Kinetix** (from *Code the Classics Volume 2*, Raspberry Pi
Press) built on [Ebitengine](https://ebitengine.org), provided as a side-by-side
counterpart to [`kinetix-go`](https://github.com/chrplr/kinetix-go), which uses
[go-sdl3](https://github.com/Zyko0/go-sdl3) + the
[pgzgo](https://github.com/chrplr/pgzgo) harness.

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

## Provenance & license

Ported to Go from the Python original in *Code the Classics Volume 2*. The game
design and original assets are © their respective authors / Raspberry Pi Press —
add the appropriate license before redistributing.

The Go source code of this port is distributed under the MIT License.
