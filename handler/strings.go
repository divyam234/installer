package handler

import (
	"regexp"
	"strings"
)

var (
	archRe    = regexp.MustCompile(`(armv8|armv7|x64|arm64|arm|386|686|amd64|x86_64|aarch64|linux64|win64)`)
	fileExtRe = regexp.MustCompile(`(\.tar)?(\.[a-z][a-z0-9]+)$`)
	posixOSRe = regexp.MustCompile(`(darwin|linux|(net|free|open)bsd|mac|osx|windows|win)`)
)

func getOS(s string) string {
	s = strings.ToLower(s)
	o := posixOSRe.FindString(s)
	if o == "mac" || o == "osx" {
		o = "darwin"
	}
	if o == "win" {
		o = "windows"
	}
	return o
}

func getArch(s string) string {
	s = strings.ToLower(s)
	a := archRe.FindString(s)
	//arch modifications
	if a == "linux64" || a == "x86_64" || a == "win64" || a == "x64" {
		a = "amd64"
	} else if a == "32" || a == "686" {
		a = "386"
	} else if a == "aarch64" || a == "armv8" {
		a = "arm64"
	}
	return a
}

func getFileExt(s string) string {
	return fileExtRe.FindString(s)
}

func splitHalf(s, by string) (string, string) {
	i := strings.Index(s, by)
	if i == -1 {
		return s, ""
	}
	return s[:i], s[i+len(by):]
}
