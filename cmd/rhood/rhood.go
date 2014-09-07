package main

import (
	"github.com/ghnail/rhood/rhood"
)

func main() {

	rhood.LoadConfig()

	go rhood.StartDownloadService()

	go rhood.Proxy()

	// Lock the goroutine, and also exit on server error
	rhood.RunGorillaMux()
}
