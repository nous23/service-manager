package util

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

// record error log on both log and std output and then exit
func ExitOnError(err error) {
	_, _ = fmt.Fprint(os.Stderr, err.Error())
	log.Fatal(err.Error())
}

// exit 0
func ExitOnSuccess(msg string) {
	_, _ = fmt.Fprintf(os.Stdout, msg)
	log.Info(msg)
	os.Exit(0)
}

// check whether file exists
func Exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
