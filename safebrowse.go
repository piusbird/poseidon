// Contains input validators for us
package main

import (
	"errors"
	"net/url"
)

func validUserAgent(agent string) bool {

	for _, value := range UserAgents {
		if value == agent {
			return true
		}
	}
	return false
}

func validateURL(strUrl string) (bool, error) {
	u, err := url.Parse(strUrl)
	if err != nil {
		return false, err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false, errors.New("unsupported Schemea")

	}
	if u.User != nil {
		return false, errors.New("authentication unsupported")
	}
	return true, nil

}
