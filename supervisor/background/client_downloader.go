package background

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type ClientDownloader struct {
	versionChecker *VersionChecker
}

func NewClientDownloader(versionChecker *VersionChecker) (*ClientDownloader, error) {
	// log.Print("New client downloader!")
	// path, err := os.Getwd()
	// if err != nil {
	// 	log.Println(err)
	// }
	// log.Printf("WD path: %s", path)

	return &ClientDownloader{
		versionChecker: versionChecker,
	}, nil
}

func (cl *ClientDownloader) CreateFolderIfNeeded(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(path, 0755)
		if errDir != nil {
			return errDir
		}
	}

	return nil
}

func (cl *ClientDownloader) DownloadLatestVersion() error {
	log.Printf("Downloading latest client version! %s\n", cl.versionChecker.ClientStatus.LatestUrl)

	cl.CreateFolderIfNeeded("./data/downloads")

	fileName := "./data/downloads/client_tmp.zip"

	// Create blank file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}

	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	// Put content on file
	resp, err := client.Get(cl.versionChecker.ClientStatus.LatestUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	size, err := io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	closeErr := file.Close()
	if closeErr != nil {
		log.Println("Failed to close temporary file!")
		log.Println(closeErr)
	}

	log.Printf("Downloaded new client with a size of %d, unzipping...\n", size)

	// Extract zip to web folder
	zipErr := Unzip(fileName, "./data/web/")
	if zipErr != nil {
		log.Print("Failed to unzip downloaded client!")
		log.Print(zipErr)
		return zipErr
	}

	log.Printf("Unzipping done, removing temporal file...\n")

	// Remove file
	removeErr := os.Remove(file.Name())
	if removeErr != nil {
		log.Println("Failed to remove temporal file!")
		log.Print(removeErr)
	} else {
		log.Println("Temporal file removed! New client version has been successfully installed!")
	}

	cl.versionChecker.setLocalClientVersion(cl.versionChecker.ClientStatus.LatestVersion)

	return nil
}

func (cl *ClientDownloader) String() string {
	return "ClientDownloader"
}

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}
