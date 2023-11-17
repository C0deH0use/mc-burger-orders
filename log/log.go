package log

import (
	"log"
	"os"
)

var (
	Info    *log.Logger = log.New(os.Stderr, " [INFO] ", log.LstdFlags|log.Lmsgprefix)
	Warning *log.Logger = log.New(os.Stdout, " [WARNING] ", log.LstdFlags|log.Lmsgprefix)
	Error   *log.Logger = log.New(os.Stderr, " [ERROR] ", log.LstdFlags|log.Lmsgprefix)
)
