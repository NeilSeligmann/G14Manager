package background

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/Masterminds/semver"
	"github.com/NeilSeligmann/G15Manager/util"
)

var repoServer string = "NeilSeligmann/G15Manager"
var repoClient string = "NeilSeligmann/G15Manager-client"

type VersionCheckerStatus struct {
	currentVersion    *semver.Version `json:"currentVersion"`
	isUpdateAvailable bool            `json:"isUpdateAvailable"`
	hasCheckFailed    bool            `json:"hasCheckFailed"`
	latestVersion     string          `json:"latestVersion"`
	latestUrl         string          `json:"latestUrl"`
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
			isUpdateAvailable: false,
			hasCheckFailed:    false,
			currentVersion:    &semver.Version{},
			latestVersion:     current,
			latestUrl:         "",
		},
		ClientStatus: &VersionCheckerStatus{
			isUpdateAvailable: false,
			hasCheckFailed:    false,
			currentVersion:    &semver.Version{},
			latestVersion:     "",
			latestUrl:         "",
		},
	}, nil
}

func (v *VersionChecker) String() string {
	return "VersionChecker"
}

func (v *VersionChecker) Serve(haltCtx context.Context) error {
	log.Println("[VersionChecker] starting checker loop")

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
				v.ServerStatus.latestUrl = latest.Assets[0].DownloadUrl
				v.ServerStatus.latestVersion = latest.version.String()
				v.ServerStatus.hasCheckFailed = false

				if latest.version != nil {
					if latest.version.GreaterThan(v.ServerStatus.currentVersion) {
						log.Printf("[VersionChecker] new server version found: %s\n", latest.version.String())

						v.ServerStatus.isUpdateAvailable = true

						v.notifier <- util.Notification{
							Message: fmt.Sprintf("A new version of G15Manager is available: %s", latest.version.String()),
						}
					}
				}
			} else {
				log.Printf("[VersionChecker] error checking for a new server version: %+v\n", err)
				v.ServerStatus.hasCheckFailed = true
			}

			// Client Update Check
			log.Println("[VersionChecker] checking for client")
			latest, err = v.getLatestVersion(repoClient)

			if err == nil && latest.version != nil {
				log.Printf("assets: %v", len(latest.Assets))
				v.ClientStatus.latestUrl = latest.Assets[0].DownloadUrl
				v.ClientStatus.latestVersion = latest.version.String()
				v.ClientStatus.hasCheckFailed = false

				if latest.version.GreaterThan(v.ClientStatus.currentVersion) {
					log.Printf("[VersionChecker] new client version found: %s\n", latest.version.String())
					v.ClientStatus.isUpdateAvailable = true
				}
			} else {
				log.Printf("[VersionChecker] error checking for a new client version: %+v\n", err)
				v.ClientStatus.hasCheckFailed = true
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
