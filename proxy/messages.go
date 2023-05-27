package proxy

import (
	"log"
	"regexp"
)

type interpolationParams struct {
	serverName    string
	serverAddress string
	connectHost   string
}

func interpolateMessage(message string, params interpolationParams) string {
	regex, err := regexp.Compile("(%[nah])")
	if err != nil {
		log.Fatalln(err)
	}
	return regex.ReplaceAllStringFunc(message, func(val string) string {
		switch val {
		case "%n":
			return params.serverName
		case "%a":
			return params.serverAddress
		case "%h":
			return params.connectHost
		default:
			return val
		}
	})
}
