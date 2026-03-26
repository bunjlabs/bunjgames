package storage

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type ArchiveType int

const (
	ArchiveNone ArchiveType = iota
	ArchiveTarGz
	ArchiveZip
)

func Unzip(filePath string, dest string) error {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		name, err := url.PathUnescape(f.Name)
		if err != nil {
			name = f.Name
		}
		target := filepath.Join(dest, name)

		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}

		if _, err := io.Copy(outFile, rc); err != nil {
			outFile.Close()
			rc.Close()
			return err
		}
		outFile.Close()
		rc.Close()
	}

	return nil
}

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
	archiveType ArchiveType,
) (io.ReadCloser, error) {
	src, err := file.Open()
	if err != nil {
		return nil, err
	}

	if archiveType == ArchiveNone {
		return src, nil
	}

	defer src.Close()

	switch archiveType {
	case ArchiveTarGz:
		if err := Ungzip(src, destination); err != nil {
			return nil, err
		}
	case ArchiveZip:
		tmpFile, err := os.CreateTemp("", "upload-*.zip")
		if err != nil {
			return nil, err
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		if _, err := io.Copy(tmpFile, src); err != nil {
			return nil, err
		}

		if err := Unzip(tmpFile.Name(), destination); err != nil {
			return nil, err
		}
	}

	yamlFile := filepath.Join(destination, "content.yaml")
	xmlFile := filepath.Join(destination, "content.xml")

	fileStream, err := os.Open(yamlFile)
	if err != nil {
		fileStream, err = os.Open(xmlFile)
		if err != nil {
			_ = os.RemoveAll(destination)
			return nil, fmt.Errorf("cannot find content file (tried yaml and xml)")
		}
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
