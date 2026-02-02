package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	coreglib "github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
)

var client *APIClient
var config *Config

func main() {
	var err error
	config, err = LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	client = NewAPIClient(config.BaseURL, config.Username, config.Password)

	app := gtk.NewApplication("com.github.mikaelstaldal.linksaver-desktop", gio.ApplicationFlagsNone)
	app.ConnectActivate(func() { activate(app) })

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

func activate(app *gtk.Application) {
	window := gtk.NewApplicationWindow(app)
	window.SetTitle("Link Saver")
	window.SetDefaultSize(800, 600)

	mainBox := gtk.NewBox(gtk.OrientationVertical, 6)
	window.SetChild(mainBox)

	// Header bar with the Search and Add button
	headerBox := gtk.NewBox(gtk.OrientationHorizontal, 6)
	headerBox.SetMarginTop(6)
	headerBox.SetMarginBottom(6)
	headerBox.SetMarginStart(6)
	headerBox.SetMarginEnd(6)
	mainBox.Append(headerBox)

	searchEntry := gtk.NewEntry()
	searchEntry.SetPlaceholderText("Search...")
	searchEntry.SetHExpand(true)
	headerBox.Append(searchEntry)

	addLinkButton := gtk.NewButtonWithLabel("Add Link")
	addNoteButton := gtk.NewButtonWithLabel("Add Note")
	refreshButton := gtk.NewButtonWithLabel("Refresh")
	settingsButton := gtk.NewButtonWithLabel("Settings")
	headerBox.Append(addLinkButton)
	headerBox.Append(addNoteButton)
	headerBox.Append(refreshButton)
	headerBox.Append(settingsButton)

	// Scrollable list area
	scrolled := gtk.NewScrolledWindow()
	scrolled.SetVExpand(true)
	mainBox.Append(scrolled)

	listView := gtk.NewListBox()
	listView.SetSelectionMode(gtk.SelectionNone)
	scrolled.SetChild(listView)

	var refreshList func()
	refreshList = func() {
		searchTerm := searchEntry.Text()
		links, err := client.GetItems(searchTerm)
		if err != nil {
			showErrorDialog(&window.Window, "Error fetching items", err.Error())
			return
		}

		// Clear existing rows
		for {
			child := listView.FirstChild()
			if child == nil {
				break
			}
			listView.Remove(child)
		}

		for _, link := range links {
			linkRow := createItemRow(&window.Window, link, refreshList)
			listView.Append(linkRow)
		}
	}

	searchEntry.ConnectChanged(refreshList)
	addLinkButton.ConnectClicked(func() {
		newLinkDialog(&window.Window, refreshList)
	})
	addNoteButton.ConnectClicked(func() {
		addNoteDialog(&window.Window, refreshList)
	})
	refreshButton.ConnectClicked(refreshList)
	settingsButton.ConnectClicked(func() {
		showSettingsDialog(&window.Window, func() {
			client = NewAPIClient(config.BaseURL, config.Username, config.Password)
			refreshList()
		})
	})

	refreshList()

	// Drag and Drop support: accept dropped text/URIs
	dropTarget := gtk.NewDropTarget(coreglib.TypeString, gdk.ActionCopy)
	dropTarget.ConnectDrop(func(value *coreglib.Value, x, y float64) bool {
		text := value.String()
		// text may be a raw URL or a text/uri-list (newline separated, may include comments starting with '#')
		var candidate string
		for _, line := range strings.Split(text, "\n") {
			l := strings.TrimSpace(line)
			if l == "" || strings.HasPrefix(l, "#") {
				continue
			}
			candidate = l
			break
		}
		if candidate == "" {
			candidate = strings.TrimSpace(text)
		}
		if candidate == "" {
			return false
		}
		if err := client.AddLink(candidate); err != nil {
			showErrorDialog(&window.Window, "Error adding link via drag and drop", err.Error())
			return false
		}
		refreshList()
		return true
	})
	window.AddController(dropTarget)

	window.Show()
}

func createItemRow(parent *gtk.Window, link Item, onUpdate func()) *gtk.Box {
	row := gtk.NewBox(gtk.OrientationHorizontal, 12)
	row.SetMarginTop(6)
	row.SetMarginBottom(6)
	row.SetMarginStart(12)
	row.SetMarginEnd(12)
	row.SetHExpand(true)

	textVBox := gtk.NewBox(gtk.OrientationVertical, 2)
	textVBox.SetHExpand(true)
	row.Append(textVBox)

	titleLabel := gtk.NewLabel(link.Title)
	titleLabel.SetHAlign(gtk.AlignStart)
	titleLabel.SetEllipsize(pango.EllipsizeEnd)
	titleLabel.SetSingleLineMode(true)
	// Make title bold
	titleLabel.SetMarkup(fmt.Sprintf("<b>%s</b>", link.Title))
	textVBox.Append(titleLabel)

	if link.URL != "" && !link.IsNote() {
		urlLabel := gtk.NewLabel(link.URL)
		urlLabel.SetHAlign(gtk.AlignStart)
		urlLabel.SetEllipsize(pango.EllipsizeEnd)
		urlLabel.SetSingleLineMode(true)
		urlLabel.SetMarkup(fmt.Sprintf("<a href=\"%s\">%s</a>", link.URL, link.URL))
		clickGesture := gtk.NewGestureClick()
		row.AddController(clickGesture)
		clickGesture.ConnectReleased(func(n int, x, y float64) {
			err := gio.AppInfoLaunchDefaultForURI(link.URL, nil)
			if err != nil {
				showErrorDialog(parent, "Error opening link", fmt.Sprintf("URL: %s\n%v", link.URL, err))
			}
		})
		textVBox.Append(urlLabel)
	}

	if link.Description != "" {
		descriptionLabel := gtk.NewLabel(link.Description)
		descriptionLabel.SetHAlign(gtk.AlignStart)
		descriptionLabel.SetWrap(true)
		textVBox.Append(descriptionLabel)
	}

	buttonBox := gtk.NewBox(gtk.OrientationHorizontal, 6)
	buttonBox.SetHAlign(gtk.AlignEnd)
	row.Append(buttonBox)

	editButton := gtk.NewButtonWithLabel("Edit")
	editButton.SetVAlign(gtk.AlignCenter)
	editButton.ConnectClicked(func() {
		showEditDialog(parent, link, onUpdate)
	})
	buttonBox.Append(editButton)

	deleteButton := gtk.NewButtonWithLabel("Delete")
	deleteButton.SetVAlign(gtk.AlignCenter)
	deleteButton.ConnectClicked(func() {
		if err := client.DeleteItem(strconv.FormatInt(link.ID, 10)); err != nil {
			showErrorDialog(parent, "Error deleting item", err.Error())
		} else {
			onUpdate()
		}
	})
	buttonBox.Append(deleteButton)

	return row
}

func newLinkDialog(parent *gtk.Window, onSuccess func()) {
	dialog := gtk.NewDialog()
	dialog.SetTransientFor(parent)
	dialog.SetModal(true)
	dialog.SetTitle("Add New Link")

	contentArea := dialog.ContentArea()
	contentArea.SetMarginTop(12)
	contentArea.SetMarginBottom(12)
	contentArea.SetMarginStart(12)
	contentArea.SetMarginEnd(12)

	grid := gtk.NewGrid()
	grid.SetRowSpacing(6)
	grid.SetColumnSpacing(12)
	contentArea.Append(grid)

	grid.Attach(gtk.NewLabel("URL:"), 0, 1, 1, 1)
	urlEntry := gtk.NewEntry()
	grid.Attach(urlEntry, 1, 1, 1, 1)

	dialog.AddButton("Cancel", int(gtk.ResponseCancel))
	dialog.AddButton("Save", int(gtk.ResponseAccept))

	dialog.ConnectResponse(func(responseID int) {
		if responseID == int(gtk.ResponseAccept) {
			err := client.AddLink(urlEntry.Text())
			if err != nil {
				showErrorDialog(parent, "Error saving link", err.Error())
			} else {
				onSuccess()
			}
		}
		dialog.Destroy()
	})

	dialog.Show()
}

func addNoteDialog(parent *gtk.Window, onSuccess func()) {
	dialog := gtk.NewDialog()
	dialog.SetTransientFor(parent)
	dialog.SetModal(true)
	dialog.SetTitle("Add New Note")

	contentArea := dialog.ContentArea()
	contentArea.SetMarginTop(12)
	contentArea.SetMarginBottom(12)
	contentArea.SetMarginStart(12)
	contentArea.SetMarginEnd(12)

	grid := gtk.NewGrid()
	grid.SetRowSpacing(6)
	grid.SetColumnSpacing(12)
	contentArea.Append(grid)

	grid.Attach(gtk.NewLabel("Title:"), 0, 0, 1, 1)
	titleEntry := gtk.NewEntry()
	grid.Attach(titleEntry, 1, 0, 1, 1)

	grid.Attach(gtk.NewLabel("Text:"), 0, 2, 1, 1)
	textView := gtk.NewTextView()
	textView.SetWrapMode(gtk.WrapWord)
	textView.SetVExpand(true)
	scroll := gtk.NewScrolledWindow()
	scroll.SetChild(textView)
	scroll.SetMinContentHeight(200)
	scroll.SetMinContentWidth(400)
	grid.Attach(scroll, 1, 2, 1, 1)

	dialog.AddButton("Cancel", int(gtk.ResponseCancel))
	dialog.AddButton("Save", int(gtk.ResponseAccept))

	dialog.ConnectResponse(func(responseID int) {
		if responseID == int(gtk.ResponseAccept) {
			buffer := textView.Buffer()
			start, end := buffer.Bounds()
			text := buffer.Text(start, end, false)
			err := client.AddNote(titleEntry.Text(), text)
			if err != nil {
				showErrorDialog(parent, "Error saving note", err.Error())
			} else {
				onSuccess()
			}
		}
		dialog.Destroy()
	})

	dialog.Show()
}

func showEditDialog(parent *gtk.Window, link Item, onSuccess func()) {
	dialog := gtk.NewDialog()
	dialog.SetTransientFor(parent)
	dialog.SetModal(true)
	dialog.SetTitle("Edit Item")

	contentArea := dialog.ContentArea()
	contentArea.SetMarginTop(12)
	contentArea.SetMarginBottom(12)
	contentArea.SetMarginStart(12)
	contentArea.SetMarginEnd(12)

	grid := gtk.NewGrid()
	grid.SetRowSpacing(6)
	grid.SetColumnSpacing(12)
	contentArea.Append(grid)

	grid.Attach(gtk.NewLabel("Title:"), 0, 0, 1, 1)
	titleEntry := gtk.NewEntry()
	grid.Attach(titleEntry, 1, 0, 1, 1)

	grid.Attach(gtk.NewLabel("Description:"), 0, 2, 1, 1)
	textView := gtk.NewTextView()
	textView.SetWrapMode(gtk.WrapWord)
	textView.SetVExpand(true)
	scroll := gtk.NewScrolledWindow()
	scroll.SetChild(textView)
	scroll.SetMinContentHeight(200)
	scroll.SetMinContentWidth(400)
	grid.Attach(scroll, 1, 2, 1, 1)

	titleEntry.SetText(link.Title)
	textView.Buffer().SetText(link.Description)

	dialog.AddButton("Cancel", int(gtk.ResponseCancel))
	dialog.AddButton("Save", int(gtk.ResponseAccept))

	dialog.ConnectResponse(func(responseID int) {
		if responseID == int(gtk.ResponseAccept) {
			buffer := textView.Buffer()
			start, end := buffer.Bounds()
			description := buffer.Text(start, end, false)
			err := client.UpdateItem(strconv.FormatInt(link.ID, 10), titleEntry.Text(), description)
			if err != nil {
				showErrorDialog(parent, "Error saving item", err.Error())
			} else {
				onSuccess()
			}
		}
		dialog.Destroy()
	})

	dialog.Show()
}

func showSettingsDialog(parent *gtk.Window, onSuccess func()) {
	dialog := gtk.NewDialog()
	dialog.SetTransientFor(parent)
	dialog.SetModal(true)
	dialog.SetTitle("Settings")

	contentArea := dialog.ContentArea()
	contentArea.SetMarginTop(12)
	contentArea.SetMarginBottom(12)
	contentArea.SetMarginStart(12)
	contentArea.SetMarginEnd(12)

	grid := gtk.NewGrid()
	grid.SetRowSpacing(6)
	grid.SetColumnSpacing(12)
	contentArea.Append(grid)

	grid.Attach(gtk.NewLabel("API URL:"), 0, 0, 1, 1)
	urlEntry := gtk.NewEntry()
	urlEntry.SetText(config.BaseURL)
	urlEntry.SetHExpand(true)
	grid.Attach(urlEntry, 1, 0, 1, 1)

	grid.Attach(gtk.NewLabel("Username:"), 0, 1, 1, 1)
	usernameEntry := gtk.NewEntry()
	usernameEntry.SetText(config.Username)
	grid.Attach(usernameEntry, 1, 1, 1, 1)

	grid.Attach(gtk.NewLabel("Password:"), 0, 2, 1, 1)
	passwordEntry := gtk.NewEntry()
	passwordEntry.SetVisibility(false)
	passwordEntry.SetText(config.Password)
	grid.Attach(passwordEntry, 1, 2, 1, 1)

	dialog.AddButton("Cancel", int(gtk.ResponseCancel))
	dialog.AddButton("Save", int(gtk.ResponseAccept))

	dialog.ConnectResponse(func(responseID int) {
		if responseID == int(gtk.ResponseAccept) {
			config.BaseURL = urlEntry.Text()
			config.Username = usernameEntry.Text()
			config.Password = passwordEntry.Text()
			if err := config.Save(); err != nil {
				showErrorDialog(parent, "Error saving config", err.Error())
			} else {
				onSuccess()
			}
		}
		dialog.Destroy()
	})

	dialog.Show()
}

func showErrorDialog(parent *gtk.Window, title, message string) {
	dialog := gtk.NewMessageDialog(parent, gtk.DialogModal, gtk.MessageError, gtk.ButtonsClose)
	dialog.SetTitle(title)
	dialog.SetMarkup(glib.MarkupEscapeText(message))
	dialog.ConnectResponse(func(responseID int) {
		dialog.Destroy()
	})
	dialog.Show()
}
