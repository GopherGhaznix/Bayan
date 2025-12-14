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
	"github.com/GopherGhaznix/Bayan/resources"
)

// FileExplorer is a UI component for navigating files
type FileExplorer struct {
	RootPath    string // The starting root path
	CurrentPath string
	OnOpenFile  func(string) // Callback when a file is selected
	OnExit      func()       // Callback to exit explorer

	window    fyne.Window // Reference to window for dialogs
	container *fyne.Container
	list      *widget.List
	files     []os.DirEntry
	pathLabel *widget.Label
	upBtn     *widget.Button // Reference to update visibility

}

// NewFileExplorer creates a new file explorer starting at root path
func NewFileExplorer(w fyne.Window, root string, onOpenFile func(string), onExit func()) *FileExplorer {
	e := &FileExplorer{
		RootPath:    root,
		CurrentPath: root,
		OnOpenFile:  onOpenFile,
		OnExit:      onExit,
		window:      w,
	}

	e.pathLabel = widget.NewLabelWithStyle(root, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	e.pathLabel.Truncation = fyne.TextTruncateEllipsis

	e.list = widget.NewList(
		func() int {
			return len(e.files)
		},
		func() fyne.CanvasObject {
			return widget.NewButtonWithIcon("", theme.FolderIcon(), nil)
		},
		func(id widget.ListItemID, o fyne.CanvasObject) {
			btn := o.(*widget.Button)
			btn.Importance = widget.LowImportance
			entry := e.files[id]
			if entry.IsDir() {
				btn.SetIcon(theme.FolderIcon())
				btn.SetText(entry.Name())
			} else {
				btn.SetIcon(theme.DocumentIcon())
				btn.SetText(strings.TrimRight(entry.Name(), ".md"))
			}
			btn.OnTapped = func() {
				e.onItemTapped(id)
			}
			btn.Alignment = widget.ButtonAlignLeading
		},
	)

	// Top bar with Up button and Path
	e.upBtn = widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		e.navigateUp()
	})

	// Home button to exit explorer (if we have a callback)
	homeBtn := widget.NewButtonWithIcon("", resources.GlobeIcon(), func() {
		if e.OnExit != nil {
			e.OnExit()
		}
	})

	// Action buttons
	newFolderBtn := widget.NewButtonWithIcon("", theme.FolderNewIcon(), e.showNewFolderDialog)
	newFolderBtn.Importance = widget.HighImportance

	newFileBtn := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), e.showNewFileDialog)
	newFileBtn.Importance = widget.HighImportance

	apptabs := container.NewAppTabs(
		container.NewTabItemWithIcon(
			"Content",
			theme.DocumentIcon(),
			container.NewPadded(
				container.NewBorder(
					container.NewBorder(
						nil, nil,
						container.NewHBox(homeBtn, e.upBtn),
						container.NewHBox(newFolderBtn, newFileBtn),
						e.pathLabel,
					),
					nil,
					nil,
					nil,
					e.list,
				),
			),
		),
		container.NewTabItemWithIcon(
			"Themes",
			theme.ColorPaletteIcon(),
			container.NewCenter(
				widget.NewLabel("comming soon!"),
			),
		),
		container.NewTabItemWithIcon(
			"Git Changes",
			theme.HistoryIcon(),
			container.NewCenter(
				widget.NewLabel("comming soon!"),
			),
		),
	)

	if os.Getenv("mobile") == "true" {
		apptabs.SetTabLocation(container.TabLocationBottom)
	} else {
		apptabs.SetTabLocation(container.TabLocationLeading)
	}
	e.container = container.NewStack(apptabs)

	// Refresh content (will check visibility)
	e.refreshDir()

	return e
}

// GetUI returns the container for this component
func (e *FileExplorer) GetUI() fyne.CanvasObject {
	return e.container
}

func (e *FileExplorer) refreshDir() {
	entries, err := os.ReadDir(e.CurrentPath)
	if err != nil {
		fyne.LogError("Failed to read dir", err)
		return
	}

	// Filter and sort: Folders first, then MD files. Ignore hidden.
	var distinctDirs, distinctFiles []os.DirEntry
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue // Skip hidden files
		}
		if entry.IsDir() {
			distinctDirs = append(distinctDirs, entry)
		} else if strings.HasSuffix(strings.ToLower(entry.Name()), ".md") {
			distinctFiles = append(distinctFiles, entry)
		}
	}

	sort.Slice(distinctDirs, func(i, j int) bool { return distinctDirs[i].Name() < distinctDirs[j].Name() })
	sort.Slice(distinctFiles, func(i, j int) bool { return distinctFiles[i].Name() < distinctFiles[j].Name() })

	e.files = append(distinctDirs, distinctFiles...)
	e.pathLabel.SetText(filepath.Base(e.CurrentPath)) // Show only current folder name for brevity

	// Check root for Up button visibility
	if e.CurrentPath == e.RootPath {
		e.upBtn.Hide()
	} else {
		e.upBtn.Show()
	}

	e.list.Refresh()
	e.list.ScrollToTop() // Reset scroll
}

func (e *FileExplorer) onItemTapped(id widget.ListItemID) {
	if id >= len(e.files) {
		return
	}
	entry := e.files[id]
	fullPath := filepath.Join(e.CurrentPath, entry.Name())

	if entry.IsDir() {
		e.CurrentPath = fullPath
		e.refreshDir()
	} else {
		if e.OnOpenFile != nil {
			e.OnOpenFile(fullPath)
		}
	}
}

func (e *FileExplorer) navigateUp() {
	if e.CurrentPath == e.RootPath {
		return
	}
	parent := filepath.Dir(e.CurrentPath)
	if parent == e.CurrentPath {
		return // System root
	}
	// Extra safety: ensure we don't go above root path if somehow we got here
	// This simple check might be enough if we just string compare,
	// assuming RootPath is absolute or relative same way.

	e.CurrentPath = parent
	e.refreshDir()
}

func (e *FileExplorer) showNewFolderDialog() {
	newFolderDialog := dialog.NewEntryDialog("New Folder", "Folder Name", func(name string) {
		if name == "" {
			return
		}
		path := filepath.Join(e.CurrentPath, name)
		if err := os.Mkdir(path, 0755); err != nil {
			dialog.ShowError(err, e.window)
			return
		}
		e.refreshDir()
	}, e.window)

	newFolderDialog.Resize(fyne.NewSize(400, 170))
	newFolderDialog.Show()
}

func (e *FileExplorer) showNewFileDialog() {
	newFileDialog := dialog.NewEntryDialog("New Markdown File", "File Name (without .md)", func(name string) {
		if name == "" {
			return
		}
		if !strings.HasSuffix(name, ".md") {
			name += ".md"
		}
		path := filepath.Join(e.CurrentPath, name)

		// Create empty file
		f, err := os.Create(path)
		if err != nil {
			dialog.ShowError(err, e.window)
			return
		}
		f.Close()

		e.refreshDir()
		// Optionally open it immediately
		if e.OnOpenFile != nil {
			e.OnOpenFile(path)
		}
	}, e.window)

	newFileDialog.Resize(fyne.NewSize(400, 170))
	newFileDialog.Show()
}
