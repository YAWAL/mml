package main

import (
	"encoding/json"
	"fmt"
	"github.com/YAWAL/mml/model"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"runtime"
	"time"
)

const (
	versionManifestUrl    = "https://launchermeta.mojang.com/mc/game/version_manifest.json"
	minecraftResourcesUrl = "https://resources.download.minecraft.net"
	minecraftPath         = "temp" // TODO: user should set this
)

func main() {

	now := time.Now()
	versionManifest := model.VersionManifest{}

	if err := doRequest(versionManifestUrl, &versionManifest); err != nil {
		fmt.Printf("get version manifest: %s", err.Error())
		return
	}

	//for _, item := range versionManifest.Versions {
	//	fmt.Printf("version: %s \n", item.ID)
	//}

	// TODO: refactor this
	vrs := "1.16.2"

	var version model.Version

	for _, item := range versionManifest.Versions {
		if item.ID == vrs {
			version = item
		}
	}

	versionDetails := model.VersionDetails{}
	if err := doRequest(version.URL, &versionDetails); err != nil {
		fmt.Printf("get version details: %s", err.Error())
		return
	}

	if err := downloadLibraries(versionDetails.Libraries); err != nil {
		fmt.Printf("get libraries: %s", err.Error())
		return
	}

	if err := downloadClient(&versionDetails.Downloads.Client); err != nil {
		fmt.Printf("get client: %s", err.Error())
		return
	}

	assets := model.AssetsData{}
	if err := doRequest(versionDetails.AssetIndex.URL, &assets); err != nil {
		fmt.Printf("get assets: %s", err.Error())
		return
	}

	if err := downloadResources(&assets); err != nil {
		fmt.Printf("get resources: %s", err.Error())
		return
	}

	fmt.Printf("exec time: %f ", time.Since(now).Seconds())
}

func doRequest(url string, out interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &out)
}

func downloadLibraries(libraries []model.Library) error {
	libPath := minecraftPath + "/libraries/"
	for _, lib := range libraries {
		err := download(lib.Downloads.Artifact.URL, libPath+lib.Downloads.Artifact.Path)
		if err != nil {
			return err
		}
		classifiers := lib.Downloads.Classifiers
		if classifiers != nil {
			switch runtime.GOOS {
			case "linux":
				err = download(classifiers.NativesLinux.URL, libPath+classifiers.NativesLinux.Path)
			case "windows":
				err = download(classifiers.NativesWindows.URL, libPath+classifiers.NativesWindows.Path)
			case "darwin":
				err = download(classifiers.NativesMacos.URL, libPath+classifiers.NativesMacos.Path)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func downloadResources(assets *model.AssetsData) error {
	resourcePath := minecraftPath + "/assets/objects/"
	for _, val := range assets.Objects {
		url := minecraftResourcesUrl + "/" + (val.Hash)[0:2] + "/" + val.Hash
		fullPath := resourcePath + (val.Hash)[0:2] + "/" + val.Hash
		err := download(url, fullPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func downloadClient(client *DownloadItem) error {
	clientPath := minecraftPath + "/path/" // TODO: change path to correct location
	if err := download(client.URL, clientPath); err != nil {
		return err
	}
	return nil
}

func download(url, fullPath string) error {
	if url == "" {
		return nil
	}
	folder := path.Dir(fullPath)
	if err := os.MkdirAll(folder, 0700); err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	rawData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(rawData)
	if err != nil {
		return err
	}

	return nil
}
