package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// SetupDesktopFolders creates 5 folders on the user's Desktop for the game
func SetupDesktopFolders() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.Getenv("USERPROFILE")
	}
	desktopDir := filepath.Join(homeDir, "Desktop")
	if _, err := os.Stat(desktopDir); os.IsNotExist(err) {
		desktopDir = "Desktop"
		_ = os.MkdirAll(desktopDir, 0755)
	}

	var folderPaths []string
	for i := 1; i <= 5; i++ {
		folderName := fmt.Sprintf("Carpeta Huesito %d", i)
		folderPath := filepath.Join(desktopDir, folderName)
		_ = os.MkdirAll(folderPath, 0755)

		fileName := fmt.Sprintf("Huesito_%d_para_Tomi.txt", i)
		filePath := filepath.Join(folderPath, fileName)

		content := fmt.Sprintf(`=======================================================
 🦴 HUESITO #%d PARA TOMI 🦴
=======================================================

   (\_/)
   ( •.•)  ¡Guau! ¡Encontraste el archivo del Huesito #%d!
   / >🦴

INSTRUCCIONES:
1. Haz clic y ARRASTRA este archivo "%s"
2. Suéltalo dentro de la CUNA DE TOMI en tu pantalla.
3. ¡Cuidado con el mouse! Tomi sigue tu recorrido pasado con 5s de retraso.

¡Lleva los 5 huesitos a la cuna para ganar el juego!
=======================================================
`, i, i, fileName)

		_ = os.WriteFile(filePath, []byte(content), 0644)
		folderPaths = append(folderPaths, folderPath)
	}

	return folderPaths, nil
}

// CleanupDesktopFolders deletes the 5 folders created on the Desktop
func CleanupDesktopFolders(folders []string) {
	for _, folder := range folders {
		if folder != "" {
			_ = os.RemoveAll(folder)
		}
	}
}

// OpenDesktopFolder opens Windows Explorer at the Desktop folder
func OpenDesktopFolder() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.Getenv("USERPROFILE")
	}
	desktopDir := filepath.Join(homeDir, "Desktop")
	_ = exec.Command("explorer.exe", desktopDir).Start()
}
