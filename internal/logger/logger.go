package logger

import (
	"log"
	"os"
)

var LogInfo *log.Logger
var LogError *log.Logger
var LogFatal *log.Logger

// InitLoggers initializes logger for info, error and fatal messages.
func InitLoggers() {
	LogInfo = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime|log.LUTC)
	LogError = log.New(os.Stdout, "[ERROR] ", log.Ldate|log.Ltime|log.LUTC)
	LogFatal = log.New(os.Stdout, "[FATAL] ", log.Ldate|log.Ltime|log.Llongfile|log.LUTC)
}
