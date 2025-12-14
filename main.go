package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/GopherGhaznix/Bayan/config"
	"github.com/GopherGhaznix/Bayan/internal/ui"
)

func main() {

	a := app.NewWithID("bayan.ghaznix.com")
	w := a.NewWindow("Bayan CMS")

	//
	storage := a.Storage()
	files := storage.List()
	for _, file := range files {
		if file == "ssh.json" {
			key, err := storage.Open(file)
			if err != nil {
				fmt.Println(err)
			}
			json.NewDecoder(key).Decode(&config.BaseConfig)
			key.Close()
		}
	}
	//

	// Set a mobile-friendly size for testing
	if os.Getenv("mobile") == "true" {
		w.Resize(fyne.NewSize(360, 640))
	} else {
		w.Resize(fyne.NewSize(800, 600))
	}

	var selector *ui.SiteSelector

	// Define navigation structure
	// View: [SiteSelector] -> [FileExplorer] -> [Editor]

	selector = ui.NewSiteSelector(w, config.BaseConfig.WebsiteRoot, func(sitePath string) {
		// User selected a site. Open FileExplorer at sitePath/content
		contentPath := filepath.Join(sitePath, "content")
		if _, err := os.Stat(contentPath); os.IsNotExist(err) {
			// Fallback if content dir missing, just use site root
			contentPath = sitePath
		}

		var explorer *ui.FileExplorer
		explorer = ui.NewFileExplorer(w, contentPath, func(filePath string) {
			// User selected a file. Open Editor
			editor := ui.NewEditor(w, filePath, func() {
				// Close Editor -> Back to Explorer
				w.SetContent(explorer.GetUI())
			})
			w.SetContent(editor.GetUI())
		}, func() {
			// Exit Explorer -> Back to Site Selector
			w.SetContent(selector.GetUI())
		})

		w.SetContent(explorer.GetUI())
	})

	if strings.TrimSpace(config.BaseConfig.Username) == "" ||
		strings.TrimSpace(config.BaseConfig.Email) == "" ||
		len(config.BaseConfig.Key) == 0 ||
		strings.TrimSpace(config.BaseConfig.WebsiteRoot) == "" {

		fyne.LogError("No SSH config found", fmt.Errorf("no ssh config found"))
		w.SetContent(ui.AskConfigurations(a, w))
		w.ShowAndRun()

		return
	}

	w.SetContent(selector.GetUI())
	w.ShowAndRun()
}
