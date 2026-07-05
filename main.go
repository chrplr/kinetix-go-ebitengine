package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

type State int

const (
	StateTitle State = iota
	StatePlay
	StateGameOver
)

var (
	state       State
	game        *Game
	totalFrames int

	assets           *Assets
	audio            *Audio
	keyboardControls *KeyboardControls
	aiControls       *AIControls
	joystickControls *JoystickControls
)

func startControls() Controls {
	if keyboardControls.firePressed() {
		return keyboardControls
	}
	if joystickControls.firePressed() {
		return joystickControls
	}
	return nil
}

func update() {
	totalFrames++

	keyboardControls.update()
	joystickControls.update()

	switch state {
	case StateTitle:
		aiControls.update()
		game.Update()
		if c := startControls(); c != nil {
			game = NewGame(c, 3, assets, audio)
			state = StatePlay
			audio.StopMusic()
		}

	case StatePlay:
		if game.lives > 0 {
			game.Update()
		} else {
			game.playSound("game_over", 1)
			state = StateGameOver
		}

	case StateGameOver:
		if startControls() != nil {
			game = NewGame(aiControls, 3, assets, audio)
			state = StateTitle
			audio.PlayMusic("title_theme", 0.3)
		}
	}
}

func draw() {
	game.Draw()

	switch state {
	case StateTitle:
		assets.Blit("title", 0, 0)
		assets.Blit("startgame", 20, 80)
		assets.Blit("start"+itoa((totalFrames/4)%13), Width/2-125, 530)

	case StateGameOver:
		assets.Blit("gameover"+itoa((totalFrames/4)%15), Width/2-225, 450)
	}
}

type EbitenGame struct{}

func (eg *EbitenGame) Update() error {
	update()
	return nil
}

func (eg *EbitenGame) Draw(screen *ebiten.Image) {
	assets.SetScreen(screen)
	draw()
}

func (eg *EbitenGame) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return Width, Height
}

func main() {
	selftest := flag.Bool("selftest", false, "run headlessly across levels, then exit")
	flag.Parse()

	assets = NewAssets()
	audio = NewAudio()

	keyboardControls = &KeyboardControls{}
	aiControls = &AIControls{}
	joystickControls = &JoystickControls{}

	if *selftest {
		runSelftest()
		return
	}

	audio.PlayMusic("title_theme", 0.3)

	state = StateTitle
	game = NewGame(aiControls, 3, assets, audio)

	ebiten.SetWindowTitle("Kinetix")
	ebiten.SetWindowSize(Width, Height)

	if err := ebiten.RunGame(&EbitenGame{}); err != nil {
		fmt.Fprintf(os.Stderr, "Error running game: %v\n", err)
		os.Exit(1)
	}
}

func runSelftest() {
	game = NewGame(aiControls, 3, assets, audio)

	for i := 0; i < 4000; i++ {
		aiControls.update()
		game.Update()
	}
	fmt.Printf("after demo: level %d, score %d, %d balls, %d barrels, %d impacts, bricks left %d\n",
		game.levelNum, game.score, len(game.balls), len(game.barrels), len(game.impacts), game.bricksRemaining)

	pg := NewGame(keyboardControls, 3, assets, audio)
	for lvl := 0; lvl < len(LEVELS); lvl++ {
		pg.newLevel(lvl)
		for _, pu := range []int{PowerupExtendBat, PowerupGun, PowerupMagnet, PowerupSmallBat,
			PowerupMultiBall, PowerupFastBalls, PowerupSlowBalls, PowerupExtraLife} {
			b := NewBarrel(pg, pg.bat.X, BatTopEdge-5)
			b.btype = pu
			pg.barrels = append(pg.barrels, b)
		}
		for i := 0; i < 300; i++ {
			pg.Update()
		}
		fmt.Printf("level %d: %dx%d grid, bricks left %d, %d balls, score %d, lives %d\n",
			lvl, pg.numRows, pg.numCols, pg.bricksRemaining, len(pg.balls), pg.score, pg.lives)
	}

	fmt.Println("SELFTEST OK")
}
