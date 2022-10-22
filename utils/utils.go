package utils

import (
	"fmt"
	"log"
	"os"
)

/* Must prints the foratting error string with the error inside of it */
func Must(msg string, err error) {
	if err != nil {
		log.Fatal(fmt.Errorf(msg, err))
		os.Exit(1)
	}
}
