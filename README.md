
# `installer`

Quickly install pre-compiled binaries from Github releases.

Installer is an HTTP server which returns shell scripts. The returned script will detect platform OS and architecture, choose from a selection of URLs, download the appropriate file, un(zip|tar|gzip) the file, find the binary (largest file) and optionally move it into your `PATH`. Useful for installing your favourite pre-compiled programs on hosts using only `curl`.

[![GoDev](https://img.shields.io/static/v1?label=godoc&message=reference&color=00add8)](https://pkg.go.dev/github.com/divyam234/installer)

## Usage

```sh
# install <user>/<repo> from github
curl https://sh-install.vercel.app/<user>/<repo>@<release> | bash
```

*Or you can use* `wget -qO- <url> | bash`

**Path API**

* `repo` Github repository belonging to `user` (**required**)
* `release` Github release name (defaults to the **latest** release)
* `move=1` When provided as query param, downloads binary directly into `/usr/local/bin/` (defaults to working directory)
* If no matching release is found you can  use `include="search term"` query param to filter release by search term.

## Examples

* https://sh-install.vercel.app/yudai/gotty@v0.0.12
* https://sh-install.vercel.app/mholt/caddy
* https://sh-install.vercel.app/rclone/rclone

    ```sh
    $ curl -s sh-install.vercel.app/mholt/caddy?move=1 | bash
    Downloading mholt/caddy v0.8.2 (https://github.com/mholt/caddy/releases/download/v0.8.2/caddy_darwin_amd64.zip)
    ######################################################################## 100.0%
    Downloaded to /usr/local/bin/caddy
    $ caddy --version
    Caddy 0.8.2
    ```

## Private repos

You'll have to set `GITHUB_TOKEN` on both your server (instance of `installer`) and client (before you run `curl https://sh-install.vercel.app/foobar?private=1 | bash`)
