package helpers

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/gookit/color"
	providerContract "github.com/nd-tools/capyvel/contracts/providers"
	"golang.org/x/image/draw"
)

// File represents a file handler with a specific configuration.
type File struct {
	fp     providerContract.File // The file provider (e.g., storage service).
	config *FileConfig           // Configuration for file handling.
}

// FileConfig contains configuration options for file handling.
type FileConfig struct {
	ID                 string          // Unique identifier for the file handler.
	Path               string          // Path where the files are stored.
	BaseUrl            string          // Base URL for accessing the files.
	Folder             string          // Folder where the file is stored.
	DefaultCompression CompressionFile // Function to handle file compression.
}

// CompressionFile defines the signature for a function that compresses a file.
type CompressionFile func(fileReader io.Reader) (io.Reader, string, error)

// Define error constants with their corresponding messages for internal server errors (HTTP 500).
var (
	ErrFileContractNotDeclared = errors.New("file contract not declared")  // HTTP 500 Internal Server Error
	ErrFileTestFailed          = errors.New("error testing file provider") // HTTP 500 Internal Server Error
	ErrFileOpenFailed          = errors.New("failed to open file")         // HTTP 500 Internal Server Error
	ErrFileCompressionFailed   = errors.New("file compression failed")     // HTTP 500 Internal Server Error
	ErrFileSaveFailed          = errors.New("failed to save file")         // HTTP 500 Internal Server Error
	ErrFileReadFailed          = errors.New("error reading file")          // HTTP 500 Internal Server Error
	ErrFileEncodeImage         = errors.New("error encoding image")        // HTTP 500 Internal Server Error
	ErrFileDecodeImage         = errors.New("error decoding image")        // HTTP 500 Internal Server Error
	ErrFileEncodeImagePNG      = errors.New("error encoding image to PNG") // HTTP 500 Internal Server Error
)

const (
	ErrFileProviderNotDeclared = "File provider is not declared. This is a required configuration."
	ErrBaseUrlRequired         = "Configuration error: 'BaseUrl' is required and cannot be empty."
	ErrIDRequired              = "Configuration error: 'ID' is required and cannot be empty."
	ErrFolderRequired          = "Configuration error: 'Folder' is required and cannot be empty."
	ErrPathRequired            = "Configuration error: 'Path' is required and cannot be empty."
	ErrFileProviderTestFailed  = "Error testing file provider on %s: %e\n"
)

// NewFile creates a new file handler instance with the provided file provider and configuration.
// It validates that the configuration fields are properly set (all fields except DefaultCompression are required),
// and tests the file provider for errors.
func NewFile(fp providerContract.File, config FileConfig) *File {
	if fp == nil {
		color.Redf(ErrFileProviderNotDeclared)
		os.Exit(1)
	}
	config.BaseUrl = strings.ReplaceAll(config.BaseUrl, " ", "")
	if config.BaseUrl == "" {
		color.Redf(ErrBaseUrlRequired)
		os.Exit(1)
	}
	config.ID = strings.ReplaceAll(config.ID, " ", "")
	if config.ID == "" {
		color.Redf(ErrIDRequired)
		os.Exit(1)
	}
	config.Folder = strings.ReplaceAll(config.Folder, " ", "")
	if config.Folder == "" {
		color.Redf(ErrFolderRequired)
		os.Exit(1)
	}
	config.Path = strings.ReplaceAll(config.Path, " ", "")
	if config.Path == "" {
		color.Redf(ErrPathRequired)
		os.Exit(1)
	}
	if err := fp.Test(); err != nil {
		color.Redf(ErrFileProviderTestFailed, config.ID, err)
		os.Exit(1)
	}
	return &File{
		fp:     fp,
		config: &config,
	}
}

// ValidateParams checks if the folder and ID match the configuration parameters.
func (f *File) ValidateParams(id, folder string) bool {
	return f.config.Folder == strings.ReplaceAll(folder, " ", "") && f.config.ID == strings.ReplaceAll(id, " ", "")
}

// GenerateUrl generates the URL to access the file with the specified name.
func (f *File) GenerateUrl(fileName string) string {
	return fmt.Sprintf("%s/%s?folder=%s&fileName=%s", f.config.BaseUrl, f.config.ID, f.config.Folder, fileName)
}

// SaveFile saves the provided file to the configured path, applying compression if necessary.
func (f *File) SaveFile(file *multipart.FileHeader, fileName string) (string, error) {
	// Get the file extension.
	ext := strings.ToLower(filepath.Ext(file.Filename))
	// Open the file.
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrFileOpenFailed, err)
	}
	defer src.Close()

	// Apply compression if configured.
	var fileReader io.Reader = src
	if f.config.DefaultCompression != nil {
		compressedFileReader, newExt, err := f.config.DefaultCompression(src)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrFileCompressionFailed, err)
		}
		fileReader = compressedFileReader
		if newExt != "" {
			ext = newExt
		}
	}

	// Define the file path to save the file.
	path := fmt.Sprintf("%s/%s", f.config.Path, fileName+ext)
	// Save the file to the specified path.
	if err := f.fp.SaveFile(fileReader, path); err != nil {
		return "", fmt.Errorf("%w: %v", ErrFileSaveFailed, err)
	}

	// Return the file name with its extension.
	return fileName + ext, nil
}

// Read retrieves the file for the specified file name.
func (f *File) Read(fileName string) (io.ReadCloser, error) {
	path := fmt.Sprintf("%s/%s", f.config.Path, fileName)
	return f.fp.ReadFile(path)
}

// Delete deletes the file for the specified file name.
func (f *File) Delete(fileName string) error {
	path := fmt.Sprintf("%s/%s", f.config.Path, fileName)
	return f.fp.DeleteFile(path)
}

// DefaultCompressImageToJPG compresses and resizes an image to JPEG format if it exceeds the limits.
// It returns the compressed image as a reader, the file extension (".jpg"), or an error.
func DefaultCompressImageToJPG(fileReader io.Reader) (io.Reader, string, error) {
	var buffer bytes.Buffer
	// Copy the content of the file reader into a buffer.
	if _, err := io.Copy(&buffer, fileReader); err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrFileReadFailed, err)
	}
	// Decode the image (supports both JPEG and PNG).
	img, _, err := image.Decode(&buffer)
	if err != nil {
		img, err = png.Decode(fileReader)
		if err != nil {
			// If both decoding attempts fail, return the original buffer.
			return &buffer, "", nil
		}
	}

	// Resize the image to fit within the specified dimensions.
	maxWidth, maxHeight := 800, 600
	img = ResizeImage(img, maxWidth, maxHeight)

	// Compress the image to JPEG format.
	var compressed bytes.Buffer
	opts := jpeg.Options{Quality: 80}
	if err := jpeg.Encode(&compressed, img, &opts); err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrFileEncodeImage, err)
	}
	return &compressed, ".jpg", nil
}

// DefaultCompressImageToPNG compresses and resizes an image to PNG format if it exceeds the limits.
// It returns the compressed image as a reader, the file extension (".png"), or an error.
func DefaultCompressImageToPNG(fileReader io.Reader) (io.Reader, string, error) {
	var buffer bytes.Buffer
	// Copy the content of the file reader into a buffer.
	if _, err := io.Copy(&buffer, fileReader); err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrFileReadFailed, err)
	}

	// Decode the image (supports both PNG and JPEG).
	img, _, err := image.Decode(&buffer)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrFileDecodeImage, err)
	}

	// Resize the image to fit within the specified dimensions.
	maxWidth, maxHeight := 640, 480
	img = ResizeImage(img, maxWidth, maxHeight)

	// Compress the image to PNG format.
	var compressed bytes.Buffer
	if err := png.Encode(&compressed, img); err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrFileEncodeImagePNG, err)
	}

	return &compressed, ".png", nil
}

// ResizeImage resizes an image to fit within the specified maximum width and height while maintaining its aspect ratio.
// It returns the resized image.
func ResizeImage(img image.Image, maxWidth, maxHeight int) image.Image {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// If the image fits within the limits, return it as-is.
	if width <= maxWidth && height <= maxHeight {
		return img
	}

	// Calculate the new width and height maintaining the aspect ratio.
	var newWidth, newHeight int
	aspectRatio := float64(width) / float64(height)

	if aspectRatio > 1 {
		newWidth = maxWidth
		newHeight = int(float64(maxWidth) / aspectRatio)
		if newHeight > maxHeight {
			newHeight = maxHeight
			newWidth = int(float64(maxHeight) * aspectRatio)
		}
	} else {
		newHeight = maxHeight
		newWidth = int(float64(maxHeight) * aspectRatio)
		if newWidth > maxWidth {
			newWidth = maxWidth
			newHeight = int(float64(maxWidth) / aspectRatio)
		}
	}

	// Create a new image with the calculated dimensions.
	resizedImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.CatmullRom.Scale(resizedImg, resizedImg.Bounds(), img, bounds, draw.Over, nil)

	return resizedImg
}
