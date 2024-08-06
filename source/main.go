package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("CSV Sorter")

	selectButton := widget.NewButton("Select CSV File", func() {
		// Get the current working directory
		pwd, err := os.Getwd()
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to get current working directory: %v", err), myWindow)
			return
		}

		// Convert the pwd to a ListableURI
		uri := storage.NewFileURI(pwd)
		lister, err := storage.ListerForURI(uri)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to create lister for pwd: %v", err), myWindow)
			return
		}

		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			filePath := reader.URI().Path()
			processCSV(filePath, myWindow)
		}, myWindow)

		fd.SetFilter(storage.NewExtensionFileFilter([]string{".csv"}))
		fd.SetLocation(lister)

		// Resize the file dialog to a larger size
		fd.Resize(fyne.NewSize(800, 800))

		fd.Show()
	})

	// Create a centered container for the button
	centeredContainer := container.NewCenter(selectButton)

	// Create a padding container to add some space around the edges
	paddedContainer := container.NewPadded(centeredContainer)

	myWindow.SetContent(paddedContainer)
	myWindow.Resize(fyne.NewSize(600, 600)) // Increase the main window size
	myWindow.ShowAndRun()
}

func processCSV(filePath string, window fyne.Window) {
	file, err := os.Open(filePath)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to open the file: %v", err), window)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to read the CSV file: %v", err), window)
		return
	}

	headers := records[0]
	tierIndex := findIndex(headers, "Tier")
	statusIndex := findIndex(headers, "Patron Status")
	nameIndex := findIndex(headers, "Name")

	var data [][]string
	for _, record := range records[1:] {
		patronStatus := strings.TrimSpace(record[statusIndex])
		if strings.EqualFold(patronStatus, "Active Patron") {
			data = append(data, record)
		}
	}

	sort.SliceStable(data, func(i, j int) bool {
		return data[i][tierIndex] < data[j][tierIndex]
	})

	var sb strings.Builder
	previousTier := ""

	for _, record := range data {
		currentTier := record[tierIndex]
		if currentTier != previousTier {
			sb.WriteString("---------------------------\n")
			previousTier = currentTier
		}
		sb.WriteString(record[nameIndex] + "\n")
	}

	// Get the directory of the selected file
	dir := filepath.Dir(filePath)
	// Get the base name of the selected file without extension
	baseName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filepath.Base(filePath)))
	// Create the output path with the base name and _sort.txt suffix
	outputPath := filepath.Join(dir, baseName+"_sort.txt")

	err = os.WriteFile(outputPath, []byte(sb.String()), 0644)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to write to the file: %v", err), window)
		return
	}

	dialog.ShowInformation("Success", fmt.Sprintf("'Name' column exported as %s.", filepath.Base(outputPath)), window)
}

func findIndex(headers []string, name string) int {
	for i, header := range headers {
		if header == name {
			return i
		}
	}
	return -1
}
