package logger

import (
	"log"
	"os"
)

var LogInfo *log.Logger
var LogError *log.Logger
var LogFatal *log.Logger

func InitLoggers() {
	LogInfo = log.New(os.Stdout, "INFO|", log.Ldate|log.Ltime)
	LogError = log.New(os.Stdout, "ERROR|", log.Ldate|log.Ltime)
	LogFatal = log.New(os.Stdout, "FATAL|", log.Ldate|log.Ltime)
}
