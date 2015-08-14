/*=============================================================================
#     FileName: log.go
#       Author: sunminghong, allen.fantasy@gmail.com, http://weibo.com/5d13
#         Team: http://1201.us
#   LastChange: 2015-08-14 17:21:45
#      History:
=============================================================================*/

package gotools

import "github.com/sunminghong/freelog"

//--------------------
// LOG LEVEL
//--------------------

/*
// Log levels to control the logging output.
const (
	LevelAll = iota
	LevelTrace
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
    LevelPanic
	LevelFatal
	LevelOff
)*/

func SetLogger(inifile *string) {
    freelog.CallDepth = 3
    freelog.Start(inifile)
}

func Trace(v ...interface{}) {
    freelog.Trace(v...)
}

func Debug(v ...interface{}) {
    freelog.Debug(v...)
}

func Info(v ...interface{}) {
    freelog.Info(v...)
}

func Warn(v ...interface{}) {
    freelog.Warn(v...)
}

// Error logs a message at error level.
func Error(v ...interface{}) {
    freelog.Error(v...)
}

// Critical logs a message at critical level.
func Panic(v ...interface{}) {
    freelog.Panic(v...)
}

func PrintPanicStack() {
    freelog.PrintPanicStack()
}

