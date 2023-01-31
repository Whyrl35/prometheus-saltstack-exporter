package build

import (
	"os"
)

// Variable with default value, overwrite by makefile during build
var Version string = "development"
var Time string = "now"
var User string = "me"

/*
 * Print version information on the screen and exit
 */
func VersionInformation() {
	println("Version:\t", Version)
	println("Build-Time:\t", Time)
	println("Build-User:\t", User)
	os.Exit(0)
}
