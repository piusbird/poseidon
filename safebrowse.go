// Contains input validators for us
package main

import (
	"errors"
	"log"
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
		return false, errors.New("Unsupported Schemea")

	}
	if u.User != nil {
		return false, errors.New("Authentication unsupported")
	}
	return true, nil

}
func getProxyUrl(hostname string, schema string) (string, error) {
	log.Println(schema + hostname)
	u, err := url.Parse(schema + hostname)
	log.Println(u.String())
	if err != nil {
		return "", err
	}
	return u.String(), nil

}
