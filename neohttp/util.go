package neohttp

import (
	"github.com/pkg/errors"
	"github.com/tjbrockmeyer/cypher"
	"log"
)

func errMsg(err error, message string) error {
	return errors.WithMessage(err, message)
}

func writeLog(prefix, format string, args ...interface{}) {
	log.Printf(prefix+"cypher/neohttp: "+format, args...)
}

func debugLog(format string, args ...interface{}) {
	if cypher.Debug {
		writeLog("[DEBUG] ", format, args...)
	}
}
