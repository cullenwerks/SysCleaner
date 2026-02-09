//go:build gui

package views

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"syscleaner/pkg/priority"
)

type priorityTableData struct {
	entries []priority.PriorityEntry
}

func (d *priorityTableData) Length() int {
	return len(d.entries)
}

func (d *priorityTableData) CreateCell() fyne.CanvasObject {
	return widget.NewLabel("")
}

func (d *priorityTableData) UpdateCell(id widget.TableCellID, cell fyne.CanvasObject) {
	label := cell.(*widget.Label)
	entry := d.entries[id.Row]

	switch id.Col {
	case 0:
		label.SetText(entry.ProcessName)
	case 1:
		label.SetText(entry.CpuPriorityName)
	case 2:
		label.SetText(entry.IoPriorityName)
	case 3:
		label.SetText(entry.PagePriorityName)
	}
}

// NewPriorityPanel creates the CPU Priority Manager tab
func NewPriorityPanel(w fyne.Window) fyne.CanvasObject {
	tableData := &priorityTableData{}

	// Create table for configured processes
	table := widget.NewTable(
		func() (int, int) { return tableData.Length(), 4 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			tableData.UpdateCell(id, cell)
		},
	)

	// Set column widths
	table.SetColumnWidth(0, 250)
	table.SetColumnWidth(1, 150)
	table.SetColumnWidth(2, 150)
	table.SetColumnWidth(3, 150)

	// Header
	header := widget.NewLabel("Process Name            CPU Priority       I/O Priority       Page Priority")
	header.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}

	// Selected row for removal
	var selectedRow = -1
	table.OnSelected = func(id widget.TableCellID) {
		selectedRow = id.Row
	}

	// Refresh table function
	refreshTable := func() {
		entries, err := priority.ListConfiguredPriorities()
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		tableData.entries = entries
		table.Refresh()
	}

	// Remove button
	removeBtn := widget.NewButton("Remove Selected", func() {
		if selectedRow < 0 || selectedRow >= len(tableData.entries) {
			dialog.ShowInformation("No Selection", "Please select a process to remove", w)
			return
		}

		processName := tableData.entries[selectedRow].ProcessName
		dialog.ShowConfirm("Confirm Removal",
			fmt.Sprintf("Remove priority settings for %s?", processName),
			func(confirmed bool) {
				if confirmed {
					if err := priority.RemoveProcessPriority(processName); err != nil {
						dialog.ShowError(err, w)
					} else {
						dialog.ShowInformation("Success",
							fmt.Sprintf("Priority settings removed for %s", processName), w)
						refreshTable()
						selectedRow = -1
					}
				}
			}, w)
	})

	configuredSection := container.NewBorder(
		widget.NewLabelWithStyle("Configured Process Priorities", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		removeBtn,
		nil, nil,
		container.NewScroll(table),
	)

	// Add new process section
	processNameEntry := widget.NewEntry()
	processNameEntry.SetPlaceHolder("Process name (e.g., LeagueClient.exe)")

	cpuSelect := widget.NewSelect([]string{"Idle", "Below Normal", "Normal", "Above Normal", "High"}, nil)
	cpuSelect.SetSelected("Normal")

	ioSelect := widget.NewSelect([]string{"Very Low", "Low", "Normal", "High"}, nil)
	ioSelect.SetSelected("Normal")

	pageSelect := widget.NewSelect([]string{"Idle", "Very Low", "Low", "Background", "Default", "Normal"}, nil)
	pageSelect.SetSelected("Normal")

	browseBtn := widget.NewButton("Browse...", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			uri := reader.URI()
			path := uri.Path()
			parts := strings.Split(path, "/")
			if len(parts) > 0 {
				filename := parts[len(parts)-1]
				processNameEntry.SetText(filename)
			}
		}, w)
	})

	applyBtn := widget.NewButton("Apply Priority", func() {
		processName := strings.TrimSpace(processNameEntry.Text)
		if processName == "" {
			dialog.ShowError(fmt.Errorf("process name cannot be empty"), w)
			return
		}

		cpuPriorityName := cpuSelect.Selected
		ioPriorityName := ioSelect.Selected
		pagePriorityName := pageSelect.Selected

		cpuVal := priority.ParseCpuPriorityName(cpuPriorityName)
		ioVal := priority.ParseIoPriorityName(ioPriorityName)
		pageVal := priority.ParsePagePriorityName(pagePriorityName)

		if err := priority.SetProcessPriority(processName, cpuVal, ioVal, pageVal); err != nil {
			dialog.ShowError(err, w)
			return
		}

		dialog.ShowInformation("Success",
			fmt.Sprintf("Priority settings applied to %s\nChanges take effect next time the process starts", processName), w)
		refreshTable()
		processNameEntry.SetText("")
	})

	addForm := container.NewVBox(
		widget.NewLabelWithStyle("Add New Process Priority", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			widget.NewLabel("Process Name:"),
			container.NewBorder(nil, nil, nil, browseBtn, processNameEntry),
		),
		container.NewGridWithColumns(2,
			widget.NewLabel("CPU Priority:"),
			cpuSelect,
		),
		container.NewGridWithColumns(2,
			widget.NewLabel("I/O Priority:"),
			ioSelect,
		),
		container.NewGridWithColumns(2,
			widget.NewLabel("Page Priority:"),
			pageSelect,
		),
		applyBtn,
	)

	// Quick presets
	presetBtn := func(name, process string, cpu, io, page int) *widget.Button {
		return widget.NewButton(name, func() {
			processNameEntry.SetText(process)
			cpuSelect.SetSelected(priority.GetCpuPriorityName(cpu))
			ioSelect.SetSelected(priority.GetIoPriorityName(io))
			pageSelect.SetSelected(priority.GetPagePriorityName(page))
		})
	}

	presetsSection := container.NewVBox(
		widget.NewLabelWithStyle("Quick Presets", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		container.NewGridWithColumns(3,
			presetBtn("League of Legends", "LeagueClient.exe", 3, 3, 5),
			presetBtn("League Game", "League of Legends.exe", 3, 3, 5),
			presetBtn("Valorant", "VALORANT-Win64-Shipping.exe", 3, 3, 5),
		),
		container.NewGridWithColumns(3,
			presetBtn("CS2", "cs2.exe", 3, 3, 5),
			presetBtn("Fortnite", "FortniteClient-Win64-Shipping.exe", 3, 3, 5),
			presetBtn("Apex Legends", "r5apex.exe", 3, 3, 5),
		),
	)

	infoLabel := widget.NewLabel("Priority settings take effect the next time the process starts. No restart required.\n" +
		"Warning: Setting CPU priority to High for non-game processes may cause system instability.")
	infoLabel.Wrapping = fyne.TextWrapWord

	content := container.NewVBox(
		widget.NewLabelWithStyle("CPU Priority Manager", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		infoLabel,
		widget.NewSeparator(),
		addForm,
		widget.NewSeparator(),
		presetsSection,
		widget.NewSeparator(),
		configuredSection,
	)

	// Load initial data
	refreshTable()

	return container.NewScroll(container.NewPadded(content))
}
