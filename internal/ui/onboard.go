package ui

import (
	"encoding/json"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/GopherGhaznix/Bayan/config"
)

func AskConfigurations(app fyne.App, w fyne.Window) fyne.CanvasObject {

	usernameEntry := widget.NewEntry()
	usernameEntry.SetIcon(theme.AccountIcon())
	usernameEntry.SetPlaceHolder("e.g. John Doe")

	emailEntry := widget.NewEntry()
	emailEntry.SetIcon(theme.MailComposeIcon())
	emailEntry.SetPlaceHolder("e.g. john@example.com")

	keyEntry := widget.NewMultiLineEntry()
	keyEntry.SetIcon(theme.InfoIcon())
	keyEntry.SetPlaceHolder("Paste your GitHub SSH public key here")

	rootPath := binding.NewString()
	rootEntry := dialog.NewFolderOpen(func(lu fyne.ListableURI, err error) {
		if err != nil || lu == nil {
			fyne.LogError("Failed to open folder", err)
			return
		}
		rootPath.Set(lu.Path())
		fmt.Println(rootPath)
	}, w)

	saveBtn := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {

		rootPathValue, err := rootPath.Get()
		if err != nil {
			dialog.ShowError(err, app.NewWindow("Error"))
			return
		}

		if rootPathValue == "" {
			dialog.ShowError(fmt.Errorf("storage path is required"), w)
			return
		}

		Configurations := config.BaseConfiguration{
			Username:    usernameEntry.Text,
			Email:       emailEntry.Text,
			Key:         []byte(keyEntry.Text),
			WebsiteRoot: rootPathValue,
		}

		file, err := app.Storage().Create("ssh.json")
		if err != nil {
			fyne.LogError("Failed to create ssh.json", err)
			dialog.ShowError(err, app.NewWindow("Error"))
			return
		}
		defer file.Close()

		err = json.NewEncoder(file).Encode(Configurations)
		if err != nil {
			fyne.LogError("Failed to write ssh.json", err)
			dialog.ShowError(err, app.NewWindow("Error"))
			return
		}

		dialog.ShowInformation("Success", "SSH Configuration Saved", w)
		app.Quit()
	})
	saveBtn.Importance = widget.HighImportance

	rootPathLabel := widget.NewLabelWithData(rootPath)
	rootPathLabel.Wrapping = fyne.TextWrapWord

	return container.NewBorder(
		widget.NewLabelWithStyle(
			"GitHub Configuration",
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true},
		),
		saveBtn,
		nil,
		nil,
		container.NewVBox(
			widget.NewLabelWithStyle("Git Username", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			usernameEntry,

			widget.NewLabelWithStyle("Git Email", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			emailEntry,

			widget.NewLabelWithStyle("SSH Key", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			keyEntry,

			widget.NewLabelWithStyle("Storage Folder (where your websites will be stored)", fyne.TextAlignLeading, fyne.TextStyle{Italic: true, Bold: true}),
			rootPathLabel,
			widget.NewButtonWithIcon("Select", theme.FolderOpenIcon(), func() {
				rootEntry.Show()
			}),
		),
	)
}
