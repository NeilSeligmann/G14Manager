package background

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Masterminds/semver"
	"github.com/NeilSeligmann/G15Manager/util"
	"github.com/gin-gonic/gin"
)

var repoServer string = "NeilSeligmann/G15Manager"
var repoClient string = "NeilSeligmann/G15Manager-client"

var CLIENT_VER_FILE string = "./data/client_ver"

type VersionCheckerStatus struct {
	CurrentVersion    *semver.Version `json:"currentVersion"`
	IsUpdateAvailable bool            `json:"isUpdateAvailable"`
	HasCheckFailed    bool            `json:"hasCheckFailed"`
	LatestVersion     string          `json:"latestVersion"`
	LatestUrl         string          `json:"latestUrl"`
	ManualUrl         string          `json:"manualUrl"`
}

type VersionChecker struct {
	current      *semver.Version
	tick         chan time.Time
	notifier     chan<- util.Notification
	ServerStatus *VersionCheckerStatus
	ClientStatus *VersionCheckerStatus
}

type releaseAsset struct {
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	Size        int    `json:"size"`
	DownloadUrl string `json:"browser_download_url"`
}

type ReleaseStruct struct {
	TagName string         `json:"tag_name"`
	HtmlUrl string         `json:"html_url"`
	Assets  []releaseAsset `json:"assets"`
	version *semver.Version
}

func NewVersionCheck(current string, notifier chan<- util.Notification) (*VersionChecker, error) {
	sem, err := semver.NewVersion(current)
	if err != nil {
		return nil, err
	}

	tick := make(chan time.Time, 1)
	tick <- time.Now()

	return &VersionChecker{
		current:  sem,
		tick:     tick,
		notifier: notifier,
		ServerStatus: &VersionCheckerStatus{
			IsUpdateAvailable: false,
			HasCheckFailed:    false,
			CurrentVersion:    sem,
			LatestVersion:     current,
			LatestUrl:         "",
		},
		ClientStatus: &VersionCheckerStatus{
			IsUpdateAvailable: false,
			HasCheckFailed:    false,
			CurrentVersion:    &semver.Version{},
			LatestVersion:     "",
			LatestUrl:         "",
		},
	}, nil
}

func (v *VersionChecker) String() string {
	return "VersionChecker"
}

func (v *VersionChecker) Serve(haltCtx context.Context) error {
	log.Println("[VersionChecker] starting checker loop")

	_, err := v.GetLocalClientVersion()
	if err != nil {
		log.Println("Failed to load local client version:")
		log.Println(err)
	}

	go func() {
		ticker := time.NewTicker(time.Hour * 6)
		defer ticker.Stop()
		for {
			select {
			case t := <-ticker.C:
				v.tick <- t
			case <-haltCtx.Done():
				return
			}
		}
	}()

	for {
		select {
		case <-haltCtx.Done():
			log.Println("[VersionChecker] stopping checker loop")
			return nil
		case <-v.tick:
			log.Println("[VersionChecker] checking for new versions...")

			// Server Update Check
			log.Println("[VersionChecker] checking for manager/server")
			latest, err := v.getLatestVersion(repoServer)

			if err == nil {
				v.ServerStatus.ManualUrl = latest.HtmlUrl
				v.ServerStatus.LatestUrl = latest.Assets[0].DownloadUrl
				v.ServerStatus.LatestVersion = latest.version.String()
				v.ServerStatus.HasCheckFailed = false

				if latest.version != nil {
					if latest.version.GreaterThan(v.ServerStatus.CurrentVersion) {
						log.Printf("[VersionChecker] new server version found: %s\n", latest.version.String())

						v.ServerStatus.IsUpdateAvailable = true

						v.notifier <- util.Notification{
							Message: fmt.Sprintf("A new version of G15Manager is available: %s", latest.version.String()),
						}
					}
				} else {
					log.Printf("[VersionChecker] already running latest server version: %s\n", latest.version.String())
				}
			} else {
				log.Printf("[VersionChecker] error checking for a new server version: %+v\n", err)
				v.ServerStatus.HasCheckFailed = true
			}

			// Client Update Check
			log.Println("[VersionChecker] checking for client")
			latest, err = v.getLatestVersion(repoClient)

			if err == nil && latest.version != nil {
				v.ClientStatus.ManualUrl = latest.HtmlUrl
				v.ClientStatus.LatestUrl = latest.Assets[0].DownloadUrl
				v.ClientStatus.LatestVersion = latest.version.String()
				v.ClientStatus.HasCheckFailed = false

				if latest.version.GreaterThan(v.ClientStatus.CurrentVersion) {
					log.Printf("[VersionChecker] new client version found: %s\n", latest.version.String())
					v.ClientStatus.IsUpdateAvailable = true
				} else {
					log.Printf("[VersionChecker] already running latest client version: %s\n", latest.version.String())
				}
			} else {
				log.Printf("[VersionChecker] error checking for a new client version: %+v\n", err)
				v.ClientStatus.HasCheckFailed = true
			}
		}
	}
}

func (v *VersionChecker) getLatestVersion(repo string) (*ReleaseStruct, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	client := http.Client{
		Timeout: time.Second * 5,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, getErr := client.Do(req)
	if getErr != nil {
		return nil, getErr
	}
	defer res.Body.Close()

	body, redErr := ioutil.ReadAll(res.Body)
	if redErr != nil {
		return nil, redErr
	}

	r := ReleaseStruct{}
	jsonErr := json.Unmarshal(body, &r)
	if jsonErr != nil {
		return nil, jsonErr
	}

	version, verErr := semver.NewVersion(r.TagName)
	if verErr != nil {
		return nil, verErr
	}

	r.version = version

	return &r, nil
}

func (v *VersionChecker) SetLocalClientVersion(newVersion string) error {
	version, verErr := semver.NewVersion(newVersion)
	if verErr != nil {
		return verErr
	}

	err := os.WriteFile(CLIENT_VER_FILE, []byte(version.String()), 0666)
	if err != nil {
		return err
	}

	v.ClientStatus.CurrentVersion = version
	v.ClientStatus.IsUpdateAvailable = false

	return nil
}

func (v *VersionChecker) GetLocalClientVersion() (string, error) {
	content, err := ioutil.ReadFile(CLIENT_VER_FILE)
	if err != nil {
		return "", err
	}

	version, verErr := semver.NewVersion(string(content))
	if verErr != nil {
		return "", verErr
	}

	v.ClientStatus.CurrentVersion = version

	return version.String(), nil
}

func (v *VersionChecker) GetWSInfo() gin.H {
	return gin.H{
		"client": v.ClientStatus,
		"server": v.ServerStatus,
	}
}
