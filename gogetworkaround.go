package main

import (
	"github.com/ghnail/rhood/rhood"
	"fmt"
)

func main() {
	// We can't place main routine (rhood.go) here. It will be compiled to
	// 'rhood' file, but we already have dir with same name, so it is located
	// at cmd/rhood/rhood.go.
	// We need this gogetworkaround.go to avoid error message 'no buildable Go source files in ...'
	// from command 'go get github.com/ghnail/rhood'
	_ = rhood.GetConfVal("test")
	fmt.Println("The actual executable is in dir cmd/rhood. This is just a bootstrap for go get.")
}
