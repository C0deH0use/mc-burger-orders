package log

import (
	"log"
	"os"
)

var (
	Info    *log.Logger = log.New(os.Stdin, "[INFO]", log.Ldate|log.Ltime|log.Lshortfile)
	Warning *log.Logger = log.New(os.Stdin, "[WARNING]", log.Ldate|log.Ltime|log.Lshortfile)
	Error   *log.Logger = log.New(os.Stderr, "[ERROR]", log.Ldate|log.Ltime|log.Lshortfile)
)
