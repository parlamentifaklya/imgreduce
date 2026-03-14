package main

import (
	"fmt"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// OptimizeJPEG compresses a JPEG with the given quality
func OptimizeJPEG(inputPath, outputPath string, quality int) error {
	inFile, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer inFile.Close()

	img, err := jpeg.Decode(inFile)
	if err != nil {
		return err
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	opts := jpeg.Options{Quality: quality}
	return jpeg.Encode(outFile, img, &opts)
}

// OptimizePNGtoJPEG converts a PNG to JPEG and compresses it
func OptimizePNGtoJPEG(inputPath, outputPath string, quality int) error {
	inFile, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer inFile.Close()

	img, err := png.Decode(inFile)
	if err != nil {
		return err
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	opts := jpeg.Options{Quality: quality}
	return jpeg.Encode(outFile, img, &opts)
}

// GetFileSize returns the size of a file in bytes
func GetFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

// OptimizeImage decides what to do based on extension
func OptimizeImage(path string, quality int) (string, int64, int64, error) {
	ext := strings.ToLower(filepath.Ext(path))
	originalSize := GetFileSize(path)

	var outputPath string
	var err error

	switch ext {
	case ".jpg", ".jpeg":
		outputPath = strings.TrimSuffix(path, ext) + "_optimized" + ext
		err = OptimizeJPEG(path, outputPath, quality)
	case ".png":
		// Convert PNG to JPEG
		outputPath = strings.TrimSuffix(path, ext) + "_optimized.jpg"
		err = OptimizePNGtoJPEG(path, outputPath, quality)
	default:
		return "", 0, 0, fmt.Errorf("unsupported format")
	}

	if err != nil {
		return "", 0, 0, err
	}

	newSize := GetFileSize(outputPath)
	return outputPath, originalSize, newSize, nil
}

func main() {
	a := app.New()
	w := a.NewWindow("Image Reducer")

	status := widget.NewLabel("Drag and drop or select an image")

	// Slider to adjust compression quality
	qualitySlider := widget.NewSlider(10, 100)
	qualitySlider.Value = 60
	qualityLabel := widget.NewLabel(fmt.Sprintf("JPEG Quality: %d", int(qualitySlider.Value)))
	qualitySlider.OnChanged = func(v float64) {
		qualityLabel.SetText(fmt.Sprintf("JPEG Quality: %d", int(v)))
	}

	w.SetOnDropped(func(pos fyne.Position, uris []fyne.URI) {
		if len(uris) == 0 {
			return
		}

		path := uris[0].Path()
		status.SetText("Optimizing...")

		output, original, newsize, err := OptimizeImage(path, int(qualitySlider.Value))
		if err != nil {
			status.SetText("Error: " + err.Error())
			return
		}

		saved := original - newsize

		status.SetText(fmt.Sprintf(
			"Done!\nOutput: %s\nOriginal: %.2f MB\nNew: %.2f MB\nSaved: %.2f MB",
			output,
			float64(original)/(1024*1024),
			float64(newsize)/(1024*1024),
			float64(saved)/(1024*1024),
		))
	})

	button := widget.NewButton("Select Image", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}

			path := reader.URI().Path()
			reader.Close()

			status.SetText("Optimizing...")

			output, original, newSize, err := OptimizeImage(path, int(qualitySlider.Value))
			if err != nil {
				status.SetText("Error: " + err.Error())
				return
			}

			saved := original - newSize

			status.SetText(fmt.Sprintf(
				"Done!\nOutput: %s\nOriginal: %.2f MB\nNew: %.2f MB\nSaved: %.2f MB",
				output,
				float64(original)/(1024*1024),
				float64(newSize)/(1024*1024),
				float64(saved)/(1024*1024),
			))
		}, w)
		fd.Show()
	})

	content := container.NewVBox(
		widget.NewLabel("Image Optimizer"),
		qualityLabel,
		qualitySlider,
		button,
		status,
	)

	w.SetContent(content)
	w.Resize(fyne.NewSize(600, 350))
	w.ShowAndRun()
}