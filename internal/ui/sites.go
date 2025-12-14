package ui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/GopherGhaznix/Bayan/config"
	"github.com/GopherGhaznix/Bayan/internal/gitlib"
	"github.com/GopherGhaznix/Bayan/resources"
)

// SiteSelector is the main screen to choose a website
type SiteSelector struct {
	WebsitesRoot string
	OnSelectSite func(string) // Returns path to site (e.g., .../websites/Mysite)

	window    fyne.Window
	container *fyne.Container
	list      *widget.List
	sites     []os.DirEntry
}

func NewSiteSelector(w fyne.Window, root string, onSelect func(string)) *SiteSelector {
	s := &SiteSelector{
		WebsitesRoot: root,
		OnSelectSite: onSelect,
		window:       w,
	}

	s.list = widget.NewList(
		func() int {
			return len(s.sites)
		},
		func() fyne.CanvasObject {
			return widget.NewButtonWithIcon("", resources.GlobeIcon(), nil)
		},
		func(id widget.ListItemID, o fyne.CanvasObject) {
			btn := o.(*widget.Button)
			btn.Importance = widget.LowImportance
			entry := s.sites[id]
			btn.SetText(entry.Name())
			btn.OnTapped = func() {
				s.onSiteTapped(id)
			}
			btn.Alignment = widget.ButtonAlignLeading
		},
	)

	s.refreshSites()

	// Toolbar
	addBtn := widget.NewButtonWithIcon("New Site", theme.ContentAddIcon(), s.showNewSiteDialog)
	addBtn.Importance = widget.HighImportance
	topBar := container.NewBorder(
		nil,
		nil,
		nil,
		addBtn,
		widget.NewLabelWithStyle("Select Website", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	s.container = container.NewPadded(
		container.NewBorder(topBar, nil, nil, nil, container.NewScroll(s.list)),
	)

	return s
}

func (s *SiteSelector) GetUI() fyne.CanvasObject {
	return s.container
}

func (s *SiteSelector) refreshSites() {
	// Create root if not exists
	if _, err := os.Stat(s.WebsitesRoot); os.IsNotExist(err) {
		os.MkdirAll(s.WebsitesRoot, 0755)
	}

	entries, err := os.ReadDir(s.WebsitesRoot)
	if err != nil {
		fyne.LogError("Failed to read websites dir", err)
		return
	}

	var dirs []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			dirs = append(dirs, entry)
		}
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name() < dirs[j].Name() })

	s.sites = dirs
	s.list.Refresh()
}

func (s *SiteSelector) onSiteTapped(id widget.ListItemID) {
	if id >= len(s.sites) {
		return
	}
	entry := s.sites[id]
	fullPath := filepath.Join(s.WebsitesRoot, entry.Name())

	if s.OnSelectSite != nil {
		s.OnSelectSite(fullPath)
	}
}

func (s *SiteSelector) showNewSiteDialog() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Site name")

	repoEntry := widget.NewEntry()
	repoEntry.SetPlaceHolder("git@github.com:user/repo.git")

	cloneEntry := widget.NewCheck("", func(b bool) {})

	items := []*widget.FormItem{
		widget.NewFormItem("Site Name", nameEntry),
		widget.NewFormItem("Repo URL", repoEntry),
		widget.NewFormItem("Clone", cloneEntry),
	}

	cloneRepoDialog := dialog.NewForm(
		"New Website",
		"Create",
		"Cancel",
		items,
		func(ok bool) {
			if !ok {
				return
			}
			if cloneEntry.Checked {

				name := nameEntry.Text
				repoURL := repoEntry.Text

				if name == "" || repoURL == "" {
					return
				}

				path := filepath.Join(s.WebsitesRoot, name)

				if err := os.Mkdir(path, 0755); err != nil {
					dialog.ShowError(err, s.window)
					return
				}

				progress := widget.NewProgressBarInfinite()
				progressDialog := dialog.NewCustomWithoutButtons(
					"Cloning repo",
					container.NewVBox(
						widget.NewLabel("Cloning repository, please wait..."),
						progress,
					),
					s.window,
				)
				progressDialog.Show()

				// 🔹 Run clone in background
				go func() {
					err := gitlib.CloneRepoWithSSH(
						path,
						repoURL,
						[]byte(config.BaseConfig.Key),
					)

					// 🔹 UI updates ONLY
					fyne.Do(func() {
						progressDialog.Hide()

						if err != nil {
							dialog.ShowError(err, s.window)
							return
						}

						_ = os.Mkdir(filepath.Join(path, "content"), 0755)
						s.refreshSites()
					})
				}()
			} else {
				// create fresh new website
			}
		},
		s.window,
	)

	cloneRepoDialog.Resize(fyne.NewSize(400, 200))
	cloneRepoDialog.Show()

}
