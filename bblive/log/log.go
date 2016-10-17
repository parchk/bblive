package log

import (
	"fmt"
	"strings"

	"bbllive/conf"

	"github.com/astaxie/beego/logs"
)

func init() {
	if conf.AppConf.LogLvl == 7 {
		SetLogFuncCall(true)
	}

	logfilepath := fmt.Sprintf(`{"filename":"%s"}`, conf.AppConf.LogPath)

	SetLogger("file", logfilepath)
	SetLevel(conf.AppConf.LogLvl)
}

// Log levels to control the logging output.
const (
	LevelEmergency = iota
	LevelAlert
	LevelCritical
	LevelError
	LevelWarning
	LevelNotice
	LevelInformational
	LevelDebug
)

// BeeLogger references the used application logger.
var BeeLogger = logs.GetBeeLogger()

// SetLevel sets the global log level used by the simple logger.
func SetLevel(l int) {
	logs.SetLevel(l)
}

// SetLogFuncCall set the CallDepth, default is 3
func SetLogFuncCall(b bool) {
	logs.SetLogFuncCall(b)
}

// SetLogger sets a new logger.
func SetLogger(adaptername string, config string) error {
	return logs.SetLogger(adaptername, config)
}

// Emergency logs a message at emergency level.
func Emergency(v ...interface{}) {
	logs.Emergency(generateFmtStr(len(v)), v...)
}

// Alert logs a message at alert level.
func Alert(v ...interface{}) {
	logs.Alert(generateFmtStr(len(v)), v...)
}

// Critical logs a message at critical level.
func Critical(v ...interface{}) {
	logs.Critical(generateFmtStr(len(v)), v...)
}

// Error logs a message at error level.
func Error(v ...interface{}) {
	logs.Error(generateFmtStr(len(v)), v...)
}

func Errorf(format string, v ...interface{}) {
	m := fmt.Sprintf(format, v...)
	logs.Error(m)
}

// Warning logs a message at warning level.
func Warning(v ...interface{}) {
	logs.Warning(generateFmtStr(len(v)), v...)
}

// Warn compatibility alias for Warning()
func Warn(v ...interface{}) {
	logs.Warn(generateFmtStr(len(v)), v...)
}

// Notice logs a message at notice level.
func Notice(v ...interface{}) {
	logs.Notice(generateFmtStr(len(v)), v...)
}

// Informational logs a message at info level.
func Informational(v ...interface{}) {
	logs.Informational(generateFmtStr(len(v)), v...)
}

// Info compatibility alias for Warning()
func Info(v ...interface{}) {
	logs.Info(generateFmtStr(len(v)), v...)
}

func Infof(format string, v ...interface{}) {
	m := fmt.Sprintf(format, v...)
	logs.Info(m)
}

// Debug logs a message at debug level.
func Debug(v ...interface{}) {
	logs.Debug(generateFmtStr(len(v)), v...)
}

func Debugf(format string, v ...interface{}) {
	m := fmt.Sprintf(format, v...)
	logs.Debug(m)
}

// Trace logs a message at trace level.
// compatibility alias for Warning()
func Trace(v ...interface{}) {
	logs.Trace(generateFmtStr(len(v)), v...)
}

func generateFmtStr(n int) string {
	return strings.Repeat("%v ", n)
}
