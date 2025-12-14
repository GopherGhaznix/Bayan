package ui

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	cms "github.com/GopherGhaznix/Bayan/internal/hugo"
)

type Editor struct {
	FullPath string
	OnClose  func()

	window    fyne.Window
	mdFile    *cms.MDFile
	container *fyne.Container // The main UI container

	// Map to hold references to UI widgets for each key
	// Key -> widget
	widgetMap map[string]fyne.CanvasObject
	fieldMap  map[string]interface{} // Reference to original value (for type checking)

	// Body Content
	bodyEntry *widget.Entry
}

func NewEditor(w fyne.Window, path string, onClose func()) *Editor {
	e := &Editor{
		FullPath:  path,
		OnClose:   onClose,
		window:    w,
		widgetMap: make(map[string]fyne.CanvasObject),
		fieldMap:  make(map[string]interface{}),
	}

	e.load()

	// Metadata Form Generation
	form := widget.NewForm()

	// Sort keys for deterministic order
	keys := make([]string, 0, len(e.mdFile.MetaData))
	for k := range e.mdFile.MetaData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		val := e.mdFile.MetaData[k]
		e.fieldMap[k] = val // Store original val for later type check if needed

		// Create widget based on type
		switch v := val.(type) {
		case bool:
			check := widget.NewCheck("", nil)
			check.Checked = v
			e.widgetMap[k] = check
			form.Append(k, check)

		case []interface{}: // YAML often decodes lists as []interface{}
			// Comma separated string
			strs := make([]string, len(v))
			for i, item := range v {
				strs[i] = fmt.Sprintf("%v", item)
			}
			entry := widget.NewEntry()
			entry.SetText(strings.Join(strs, ", "))
			e.widgetMap[k] = entry
			form.Append(k, entry)

		case []string:
			entry := widget.NewEntry()
			entry.SetText(strings.Join(v, ", "))
			e.widgetMap[k] = entry
			form.Append(k, entry)

		default:
			// Treat as string
			entry := widget.NewEntry()
			entry.SetText(fmt.Sprintf("%v", v))
			e.widgetMap[k] = entry
			form.Append(k, entry)
		}
	}

	// Add a "New Field" button? Maybe later.

	// Body Content
	e.bodyEntry = widget.NewMultiLineEntry()
	e.bodyEntry.SetText(e.mdFile.Body)
	e.bodyEntry.TextStyle = fyne.TextStyle{Monospace: true}
	e.bodyEntry.Wrapping = fyne.TextWrapWord

	// Preview Content
	preview := widget.NewRichTextFromMarkdown(e.bodyEntry.Text)
	preview.Wrapping = fyne.TextWrapWord

	// Toolbar
	closeBtn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		if e.OnClose != nil {
			e.OnClose()
		}
	})
	saveBtn := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), e.save)
	saveBtn.Importance = widget.HighImportance

	label := widget.NewLabel(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)))
	if len(label.Text) > 24 {
		label.SetText(label.Text[:24] + "...")
	}

	label.Wrapping = fyne.TextWrapBreak
	topBar := container.NewBorder(nil, nil, closeBtn, saveBtn, label)

	// Layout with Tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("Metadata", container.NewVScroll(form)),
		container.NewTabItem("Content", container.NewBorder(nil, nil, nil, nil, e.bodyEntry)),
		container.NewTabItem("Preview", container.NewVScroll(preview)),
	)

	// On tab change, update preview
	tabs.OnSelected = func(i *container.TabItem) {
		if i.Text == "Preview" {
			preview.ParseMarkdown(e.bodyEntry.Text)
		}
	}

	e.container = container.NewPadded(
		container.NewBorder(topBar, nil, nil, nil, tabs),
	)

	return e
}

func (e *Editor) GetUI() fyne.CanvasObject {
	return e.container
}

func (e *Editor) load() {
	content, err := os.ReadFile(e.FullPath)
	if err != nil {
		// If file doesn't exist, assume new empty file
		e.mdFile = &cms.MDFile{MetaData: make(map[string]interface{})}
		return
	}

	md, err := cms.ParseMD(string(content))
	if err != nil {
		fyne.LogError("Failed to parse MD", err)
		dialog.ShowError(err, e.window)
		e.mdFile = &cms.MDFile{Body: string(content), MetaData: make(map[string]interface{})}
	} else {
		e.mdFile = md
		if e.mdFile.MetaData == nil {
			e.mdFile.MetaData = make(map[string]interface{})
		}
	}
}

func (e *Editor) save() {
	// Update struct from UI
	for k, widgetObj := range e.widgetMap {
		switch w := widgetObj.(type) {
		case *widget.Check:
			e.mdFile.MetaData[k] = w.Checked
		case *widget.Entry:
			// Check if it was a list
			origVal, exists := e.fieldMap[k]
			isList := false
			if exists {
				switch origVal.(type) {
				case []interface{}, []string:
					isList = true
				}
			}

			if isList {
				parts := strings.Split(w.Text, ",")
				cleanParts := make([]string, 0, len(parts))
				for _, p := range parts {
					cleanParts = append(cleanParts, strings.TrimSpace(p))
				}
				e.mdFile.MetaData[k] = cleanParts
			} else {
				e.mdFile.MetaData[k] = w.Text
			}
		}
	}

	e.mdFile.Body = e.bodyEntry.Text

	data, err := e.mdFile.ToString()
	if err != nil {
		dialog.ShowError(err, e.window)
		return
	}

	err = os.WriteFile(e.FullPath, []byte(data), 0644)
	if err != nil {
		dialog.ShowError(err, e.window)
		return
	}

	log.Println("File saved successfully")
}
