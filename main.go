package main

import (
	"os"

	"github.com/FacileStudio/sonar/cmd"
	"github.com/FacileStudio/tronc/healthcheck"
)

func main() {
	if healthcheck.Handle(os.Args) {
		return
	}
	cmd.Execute()
}
