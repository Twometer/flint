package proxy

import (
	"log"
	"regexp"
)

var messageRegex = (func() *regexp.Regexp {
	regex, err := regexp.Compile("(%[nah])")
	if err != nil {
		log.Fatalln("fatal: error compiling message regex:", err)
	}
	return regex
})()

type interpolationParams struct {
	serverName    string
	serverAddress string
	connectHost   string
}

func interpolateMessage(message string, params interpolationParams) string {
	return messageRegex.ReplaceAllStringFunc(message, func(val string) string {
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
