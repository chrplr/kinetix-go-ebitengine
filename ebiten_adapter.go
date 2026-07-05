package main

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	_ "image/png"
	"io"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	ebaudio "github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
)

//go:embed images/*.png
var imagesFS embed.FS

//go:embed sounds/*.ogg music/*.ogg
var audioFS embed.FS

type Assets struct {
	images   map[string]*ebiten.Image
	screen   *ebiten.Image
	clipRect *image.Rectangle
}

func NewAssets() *Assets {
	as := &Assets{
		images: make(map[string]*ebiten.Image),
	}
	entries, err := imagesFS.ReadDir("images")
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".png") {
			continue
		}
		key := strings.TrimSuffix(name, ".png")
		data, err := imagesFS.ReadFile("images/" + name)
		if err != nil {
			panic(err)
		}
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			panic(err)
		}
		as.images[key] = ebiten.NewImageFromImage(img)
	}
	return as
}

func (as *Assets) SetScreen(screen *ebiten.Image) {
	as.screen = screen
}

func (as *Assets) Blit(name string, x, y float64) {
	img, ok := as.images[name]
	if !ok {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)

	target := as.screen
	if as.clipRect != nil {
		target = target.SubImage(*as.clipRect).(*ebiten.Image)
	}
	target.DrawImage(img, op)
}

func (as *Assets) Size(name string) (float64, float64) {
	img, ok := as.images[name]
	if !ok {
		return 0, 0
	}
	b := img.Bounds()
	return float64(b.Dx()), float64(b.Dy())
}

func (as *Assets) SetClip(x, y, w, h float64) {
	r := image.Rect(int(x), int(y), int(x+w), int(y+h))
	as.clipRect = &r
}

func (as *Assets) ClearClip() {
	as.clipRect = nil
}

type Audio struct {
	ctx          *ebaudio.Context
	sounds       map[string][]byte
	musicPlayer  *ebaudio.Player
	currentMusic string
}

func NewAudio() *Audio {
	const sampleRate = 44100
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Warning: failed to initialize audio context: %v\n", r)
		}
	}()
	ctx := ebaudio.NewContext(sampleRate)
	aud := &Audio{
		ctx:    ctx,
		sounds: make(map[string][]byte),
	}

	entries, err := audioFS.ReadDir("sounds")
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".ogg") {
			continue
		}
		key := strings.TrimSuffix(name, ".ogg")
		data, err := audioFS.ReadFile("sounds/" + name)
		if err != nil {
			panic(err)
		}
		stream, err := vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
		if err != nil {
			panic(err)
		}
		decoded, err := io.ReadAll(stream)
		if err != nil {
			panic(err)
		}
		aud.sounds[key] = decoded
	}

	return aud
}

func (aud *Audio) PlaySound(name string, count int) {
	if aud == nil || aud.ctx == nil {
		return
	}
	idx := 0
	if count > 1 {
		idx = randInt(0, count-1)
	}
	key := name + itoa(idx)
	data, ok := aud.sounds[key]
	if !ok {
		return
	}
	p := aud.ctx.NewPlayerFromBytes(data)
	p.Play()
}

func (aud *Audio) PlayMusic(name string, volume float64) {
	if aud == nil || aud.ctx == nil {
		return
	}
	if aud.currentMusic == name {
		return
	}
	aud.StopMusic()

	data, err := audioFS.ReadFile("music/" + name + ".ogg")
	if err != nil {
		return
	}
	stream, err := vorbis.DecodeWithSampleRate(aud.ctx.SampleRate(), bytes.NewReader(data))
	if err != nil {
		return
	}
	loopStream := ebaudio.NewInfiniteLoop(stream, stream.Length())
	p, err := aud.ctx.NewPlayer(loopStream)
	if err != nil {
		return
	}
	p.SetVolume(volume)
	p.Play()
	aud.musicPlayer = p
	aud.currentMusic = name
}

func (aud *Audio) StopMusic() {
	if aud == nil || aud.ctx == nil {
		return
	}
	if aud.musicPlayer != nil {
		aud.musicPlayer.Close()
		aud.musicPlayer = nil
	}
	aud.currentMusic = ""
}

func keyLeft() bool  { return ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) }
func keyRight() bool { return ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) }
func keySpace() bool { return ebiten.IsKeyPressed(ebiten.KeySpace) }

func getGamepad() (ebiten.GamepadID, bool) {
	ids := ebiten.AppendGamepadIDs(nil)
	if len(ids) == 0 {
		return 0, false
	}
	return ids[0], true
}

func padLeft() bool {
	id, ok := getGamepad()
	if !ok {
		return false
	}
	return ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonLeftLeft)
}

func padRight() bool {
	id, ok := getGamepad()
	if !ok {
		return false
	}
	return ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonLeftRight)
}

func padAxisX() float64 {
	id, ok := getGamepad()
	if !ok {
		return 0
	}
	return ebiten.GamepadAxisValue(id, 0)
}

func padButton0() bool {
	id, ok := getGamepad()
	if !ok {
		return false
	}
	return ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonRightBottom)
}
