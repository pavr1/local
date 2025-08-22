package images

import (
	"fmt"
	"os"
	"path/filepath"
)

// ImageHandler handles image storage and retrieval operations
type ImageHandler struct {
	basePath string
}

// NewImageHandler creates a new image handler
func NewImageHandler(basePath string) *ImageHandler {
	return &ImageHandler{
		basePath: basePath,
	}
}

// AddImage stores an image in the specified service directory
// service: the service name (e.g., "recipes", "invoices")
// imageName: the name of the image file
// image: the image data as bytes
func (h *ImageHandler) AddImage(service, imageName string, image []byte) error {
	// Validate inputs
	if service == "" {
		return fmt.Errorf("service name cannot be empty")
	}
	if imageName == "" {
		return fmt.Errorf("image name cannot be empty")
	}
	if len(image) == 0 {
		return fmt.Errorf("image data cannot be empty")
	}

	// Create service directory path
	servicePath := filepath.Join(h.basePath, "images", service)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(servicePath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", servicePath, err)
	}

	// Create full file path
	filePath := filepath.Join(servicePath, imageName)

	// Write image data to file
	if err := os.WriteFile(filePath, image, 0644); err != nil {
		return fmt.Errorf("failed to write image file %s: %w", filePath, err)
	}

	return nil
}

// RetrieveImage retrieves an image from the specified service directory
// service: the service name (e.g., "recipes", "invoices")
// name: the name of the image file
// returns: the image data as bytes
func (h *ImageHandler) RetrieveImage(service, name string) ([]byte, error) {
	// Validate inputs
	if service == "" {
		return nil, fmt.Errorf("service name cannot be empty")
	}
	if name == "" {
		return nil, fmt.Errorf("image name cannot be empty")
	}

	// Create full file path
	filePath := filepath.Join(h.basePath, "images", service, name)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("image file not found: %s", filePath)
	}

	// Read image data from file
	imageData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read image file %s: %w", filePath, err)
	}

	return imageData, nil
}

// DeleteImage deletes an image from the specified service directory
// service: the service name (e.g., "recipes", "invoices")
// name: the name of the image file
func (h *ImageHandler) DeleteImage(service, name string) error {
	// Validate inputs
	if service == "" {
		return fmt.Errorf("service name cannot be empty")
	}
	if name == "" {
		return fmt.Errorf("image name cannot be empty")
	}

	// Create full file path
	filePath := filepath.Join(h.basePath, "images", service, name)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("image file not found: %s", filePath)
	}

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete image file %s: %w", filePath, err)
	}

	return nil
}

// ImageExists checks if an image exists in the specified service directory
// service: the service name (e.g., "recipes", "invoices")
// name: the name of the image file
// returns: true if image exists, false otherwise
func (h *ImageHandler) ImageExists(service, name string) bool {
	filePath := filepath.Join(h.basePath, "images", service, name)
	_, err := os.Stat(filePath)
	return err == nil
}

// GetImagePath returns the full path to an image
// service: the service name (e.g., "recipes", "invoices")
// name: the name of the image file
// returns: the full path to the image file
func (h *ImageHandler) GetImagePath(service, name string) string {
	return filepath.Join(h.basePath, "images", service, name)
}

// GetImageURL returns the URL path for an image (for web serving)
// service: the service name (e.g., "recipes", "invoices")
// name: the name of the image file
// returns: the URL path (e.g., "/images/recipes/vanilla-ice-cream.jpg")
func (h *ImageHandler) GetImageURL(service, name string) string {
	return fmt.Sprintf("/images/%s/%s", service, name)
}
