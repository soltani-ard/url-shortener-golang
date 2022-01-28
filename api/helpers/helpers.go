package helpers

import (
	"os"
	"strings"
)

// EnforceHTTP : add http to input url
func EnforceHTTP(url string) string {

	if url[:4] != "http" {
		return "http://" + url
	}
	return url
}

// RemoveDomainError : remove domain
func RemoveDomainError(url string) bool {

	// localhost
	if url == os.Getenv("DOMAIN") {
		return false
	}

	// replace with empty character
	newURL := strings.Replace(url, "http://", "", 1)
	newURL = strings.Replace(newURL, "https://", "", 1)
	newURL = strings.Replace(newURL, "www.", "", 1)

	// localhost
	if newURL == os.Getenv("DOMAIN") {
		return false
	}

	return true
}
