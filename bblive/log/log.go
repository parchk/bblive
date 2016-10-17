package log

import (
	"bbllive/conf"
	"bbllive/util"
	"fmt"
	"os"
)

var l *util.FileLogger

func init() {
	var err error

	l, err = util.NewFileLogger("", conf.AppConf.LogPath, conf.AppConf.LogLvl)

	if err != nil {
		fmt.Printf("log init error :%v", err)
		os.Exit(1)
	}
}

func Debugf(format string, v ...interface{}) {
	l.Debugf(format, v...)
}

func Debug(v ...interface{}) {
	l.Debug(v...)
}

func Infof(format string, v ...interface{}) {
	l.Infof(format, v...)
}

func Info(v ...interface{}) {
	l.Info(v...)
}

func Warnf(format string, v ...interface{}) {
	l.Warnf(format, v...)
}

func Warn(v ...interface{}) {
	l.Warn(v...)
}

func Errorf(format string, v ...interface{}) {
	l.Errorf(format, v...)
}

func Error(v ...interface{}) {
	l.Error(v...)
}
