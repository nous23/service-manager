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
