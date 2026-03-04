package storage

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

func Ungzip(fileStream io.Reader, dest string) error {
	gzReader, err := gzip.NewReader(fileStream)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)

		// Protect against zip slip vulnerability
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		default:
			return fmt.Errorf("unsupported type flag: %d for %s", header.Typeflag, header.Name)
		}
	}

	return nil
}

func UploadFile(
	file *multipart.FileHeader,
	destination string,
	unzip bool,
) (io.ReadCloser, error) {
	src, err := file.Open()
	if err != nil {
		return nil, err
	}

	var fileStream io.ReadCloser

	if unzip {
		defer src.Close()
		err := Ungzip(src, destination)
		if err != nil {
			return nil, err
		}

		yamlFile := filepath.Join(destination, "content.yaml")
		xmlFile := filepath.Join(destination, "content.xml")

		fileStream, err = os.Open(yamlFile)
		if err != nil {
			fileStream, err = os.Open(xmlFile)
			if err != nil {
				_ = os.RemoveAll(destination)
				return nil, fmt.Errorf("cannot find content file (tried yaml and xml)")
			}
		}
		defer fileStream.Close()
	} else {
		fileStream = src
	}

	return fileStream, nil
}

func CleanFolder(folder string) {
	log.Printf("Cleaning %s directory contents...", folder)
	if err := os.MkdirAll(folder, 0755); err != nil {
		log.Fatalf("Failed to create media directory: %v", err)
	}
	if entries, err := os.ReadDir(folder); err == nil {
		for _, entry := range entries {
			if entry.Name() == ".gitignore" {
				continue
			}
			path := "media/" + entry.Name()
			if err := os.RemoveAll(path); err != nil {
				log.Printf("Warning: Failed to remove %s: %v", path, err)
			}
		}
		log.Printf("%s directory cleaned successfully", folder)
	} else {
		log.Printf("Warning: Failed to read %s directory: %v", folder, err)
	}
}
