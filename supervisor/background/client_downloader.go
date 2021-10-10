package background

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
	log.Printf("Download latest version! %s", cl.versionChecker.ClientStatus.latestUrl)

	cl.CreateFolderIfNeeded("./data")
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
	resp, err := client.Get(cl.versionChecker.ClientStatus.latestUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	size, err := io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	defer file.Close()

	fmt.Printf("Downloaded a file %s with size %d", fileName, size)

	return nil
}

func (cl *ClientDownloader) String() string {
	return "ClientDownloader"
}
