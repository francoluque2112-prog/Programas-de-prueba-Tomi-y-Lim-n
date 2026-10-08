package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// 1. Initialize sound effects
	initSounds()
	defer StopAllMusic()

	// 2. Setup 5 desktop folders and ensure cleanup when the game finishes
	folders, err := SetupDesktopFolders()
	if err != nil {
		fmt.Printf("[!] Error creando carpetas en el Escritorio: %v\n", err)
	}
	defer CleanupDesktopFolders(folders)

	// 3. Load PNG Sprites
	assetsDir := "Assets"
	if _, err := os.Stat(assetsDir); os.IsNotExist(err) {
		assetsDir = "../Assets"
	}

	tvPath := filepath.Join(assetsDir, "TV antiguo.png")
	tomiPath := filepath.Join(assetsDir, "Tomi.png")
	htomiPath := filepath.Join(assetsDir, "Htomi.png")
	cunaPath := filepath.Join(assetsDir, "Cuna.png")

	spriteTV, err := LoadSprite(tvPath)
	if err != nil {
		return
	}

	spriteTomi, err := LoadSprite(tomiPath)
	if err != nil {
		return
	}

	spriteHtomi, _ := LoadSprite(htomiPath)

	spriteCuna, err := LoadSprite(cunaPath)
	if err != nil {
		return
	}

	// 4. Get Screen Resolution
	screenWVal, _, _ := procGetSystemMetrics.Call(SM_CXSCREEN)
	screenHVal, _, _ := procGetSystemMetrics.Call(SM_CYSCREEN)
	screenW := int(screenWVal)
	screenH := int(screenHVal)
	if screenW <= 0 {
		screenW = 1920
	}
	if screenH <= 0 {
		screenH = 1080
	}

	game := NewGame(screenW, screenH, spriteTV, spriteTomi, spriteHtomi, spriteCuna)
	game.DesktopFolders = folders

	_ = RunTransparentGame(game)
}
