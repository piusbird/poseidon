// Contains input validators for us
package main

func validUserAgent(agent string) bool {

	for _, value := range UserAgents {
		if value == agent {
			return true
		}
	}
	return false
}
