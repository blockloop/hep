package log

import (
	"io/ioutil"
	"log"
	"os"
)

var (
	flags = log.Ldate | log.Ltime | log.Lshortfile

	// Debug logs when verbose is enabled
	Debug = log.New(ioutil.Discard, "DEBUG: ", flags)

	// Error logs to stderr
	Error = log.New(os.Stderr, "ERROR: ", flags)
)

// Verbose enables verbose output
func Verbose(enable bool) {
	Debug.SetOutput(os.Stdout)
}
