package handler

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

func (h *Handler) execute(q Query) (Result, error) {
	//load from cache
	key := q.cacheKey()
	h.cacheMut.Lock()
	if h.cache == nil {
		h.cache = map[string]Result{}
	}
	cached, ok := h.cache[key]
	h.cacheMut.Unlock()
	//cache hit
	if ok && time.Since(cached.Timestamp) < cacheTTL {
		return cached, nil
	}
	//do real operation
	ts := time.Now()

	release, assets, err := h.getAssetsNoCache(q)

	if err != nil {
		return Result{}, err
	}
	//success
	if q.Release == "" && release != "" {
		log.Printf("detected release: %s", release)
		q.Release = release
	}
	result := Result{
		Timestamp: ts,
		Query:     q,
		Assets:    assets,
		M1Asset:   assets.HasM1(),
	}
	//success store results
	h.cacheMut.Lock()
	h.cache[key] = result
	h.cacheMut.Unlock()
	return result, nil
}

func (h *Handler) getAssetsNoCache(q Query) (string, Assets, error) {
	user := q.User
	repo := q.Program
	release := q.Release
	//not cached - ask github
	log.Printf("fetching asset info for %s/%s@%s", user, repo, release)
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", user, repo)
	ghas := ghAssets{}
	if release == "" || release == "latest" {
		url += "/latest"
		ghr := ghRelease{}
		if err := h.get(url, q.Token, &ghr); err != nil {
			return release, nil, err
		}
		release = ghr.TagName //discovered
		ghas = ghr.Assets
	} else {
		ghrs := []ghRelease{}
		if err := h.get(url, q.Token, &ghrs); err != nil {
			return release, nil, err
		}
		found := false
		for _, ghr := range ghrs {
			if ghr.TagName == release {
				found = true
				if err := h.get(ghr.AssetsURL, q.Token, &ghas); err != nil {
					return release, nil, err
				}
				ghas = ghr.Assets
				break
			}
		}
		if !found {
			return release, nil, fmt.Errorf("release tag '%s' not found", release)
		}
	}
	if len(ghas) == 0 {
		return release, nil, errors.New("no assets found")
	}

	assets := Assets{}
	index := map[string]bool{}
	for _, ga := range ghas {
		url := ga.BrowserDownloadURL
		if q.Private {
			url = ga.URL
		}
		//only binary containers are supported
		//TODO deb,rpm etc
		fext := getFileExt(ga.Name)
		if fext == "" && ga.Size > 1024*1024 {
			fext = ".bin" // +1MB binary
		}
		switch fext {
		case ".bin", ".zip", ".tar.bz", ".tar.bz2", ".bz2", ".gz", ".tar.gz", ".tgz", ".tar.xz":
			// valid
		default:
			log.Printf("fetched asset has unsupported file type: %s (ext '%s')", ga.Name, fext)
			continue
		}

		//filter name by query
		if q.Include != "" {
			skip := true
			includes := strings.Split(q.Include, ",")
			for _, include := range includes {
				if strings.Contains(ga.Name, include) {
					skip = false
				}
			}
			if skip {
				continue
			}

		}
		//match
		os := getOS(ga.Name)
		arch := getArch(ga.Name)
		//windows not supported yet
		if os == "windows" {
			log.Printf("fetched asset is for windows: %s", ga.Name)
			//TODO: powershell
			// EG: iwr https://deno.land/x/install/install.ps1 -useb | iex
			continue
		}
		//unknown os, cant use
		if os == "" {
			log.Printf("fetched asset has unknown os: %s", ga.Name)
			continue
		}
		if arch == "" {
			continue
		}
		log.Printf("fetched asset: %s", ga.Name)
		asset := Asset{
			OS:   os,
			Arch: arch,
			Name: ga.Name,
			URL:  url,
			Type: fext,
		}
		assets = append(assets, asset)

	}
	if len(assets) == 0 {
		return release, nil, errors.New("no downloads found for this release")
	}

	//there can only be 1 file for each OS/Arch

	filterAssets := Assets{}

	for _, asset := range assets {
		if index[asset.Key()] {
			continue
		}
		index[asset.Key()] = true
		filterAssets = append(filterAssets, asset)
	}

	return release, filterAssets, nil
}

type ghAssets []ghAsset

type ghAsset struct {
	BrowserDownloadURL string `json:"browser_download_url"`
	ContentType        string `json:"content_type"`
	CreatedAt          string `json:"created_at"`
	DownloadCount      int    `json:"download_count"`
	ID                 int    `json:"id"`
	Label              string `json:"label"`
	Name               string `json:"name"`
	Size               int    `json:"size"`
	State              string `json:"state"`
	UpdatedAt          string `json:"updated_at"`
	Uploader           struct {
		ID    int    `json:"id"`
		Login string `json:"login"`
	} `json:"uploader"`
	URL string `json:"url"`
}

func (g ghAsset) IsChecksumFile() bool {
	return checksumRe.MatchString(strings.ToLower(g.Name)) && g.Size < 64*1024 //maximum file size 64KB
}

type ghRelease struct {
	Assets    []ghAsset `json:"assets"`
	AssetsURL string    `json:"assets_url"`
	Author    struct {
		ID    int    `json:"id"`
		Login string `json:"login"`
	} `json:"author"`
	Body            string      `json:"body"`
	CreatedAt       string      `json:"created_at"`
	Draft           bool        `json:"draft"`
	HTMLURL         string      `json:"html_url"`
	ID              int         `json:"id"`
	Name            interface{} `json:"name"`
	Prerelease      bool        `json:"prerelease"`
	PublishedAt     string      `json:"published_at"`
	TagName         string      `json:"tag_name"`
	TarballURL      string      `json:"tarball_url"`
	TargetCommitish string      `json:"target_commitish"`
	UploadURL       string      `json:"upload_url"`
	URL             string      `json:"url"`
	ZipballURL      string      `json:"zipball_url"`
}
