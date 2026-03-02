// Package image provides image processing utilities for the OpenCode TUI.
//
// This package includes functions for:
//   - Validating image file sizes before display
//   - Converting images to ANSI-art string representations for terminal display
//   - Image resizing using Lanczos algorithm for quality downscaling
//
// The package uses the lipgloss library for terminal styling and the
// disintegration/imaging library for image processing.
package image

import (
	"fmt"
	"image"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/disintegration/imaging"
	"github.com/lucasb-eyer/go-colorful"
)

// ValidateFileSize checks if a file exceeds a given size limit.
// Returns true if the file is too large, along with any error encountered.
// This is useful for preventing memory issues when loading large images.
func ValidateFileSize(filePath string, sizeLimit int64) (bool, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return false, fmt.Errorf("error getting file info: %w", err)
	}

	if fileInfo.Size() > sizeLimit {
		return true, nil
	}

	return false, nil
}

// ToString converts an image to an ANSI-art string representation.
// The image is first resized to the given width using Lanczos algorithm,
// then converted to a string using Unicode block characters (▀) with
// foreground and background colors to simulate the image in terminal.
// Each pair of vertical pixels is rendered as one block character.
func ToString(width int, img image.Image) string {
	img = imaging.Resize(img, width, 0, imaging.Lanczos)
	b := img.Bounds()
	imageWidth := b.Max.X
	h := b.Max.Y
	str := strings.Builder{}

	for heightCounter := 0; heightCounter < h; heightCounter += 2 {
		for x := range imageWidth {
			c1, _ := colorful.MakeColor(img.At(x, heightCounter))
			color1 := lipgloss.Color(c1.Hex())

			var color2 lipgloss.Color
			if heightCounter+1 < h {
				c2, _ := colorful.MakeColor(img.At(x, heightCounter+1))
				color2 = lipgloss.Color(c2.Hex())
			} else {
				color2 = color1
			}

			str.WriteString(lipgloss.NewStyle().Foreground(color1).
				Background(color2).Render("▀"))
		}

		str.WriteString("\n")
	}

	return str.String()
}

// ImagePreview loads an image file and converts it to an ANSI-art string.
// The image is resized to fit the specified width and rendered using
// ToString function. Returns the ANSI string or an error if loading fails.
func ImagePreview(width int, filename string) (string, error) {
	imageContent, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer imageContent.Close()

	img, _, err := image.Decode(imageContent)
	if err != nil {
		return "", err
	}

	imageString := ToString(width, img)

	return imageString, nil
}
