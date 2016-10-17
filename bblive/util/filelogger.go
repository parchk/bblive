package util

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
)

type FileLogger struct {
	prefix string
	logger *log.Logger
	level  int
}

func (l *FileLogger) GetHead() string {

	_, file, line, ok := runtime.Caller(3)

	if !ok {
		file = "???"
		line = 0
	}

	_, filename := path.Split(file)

	head := "[" + filename + ":" + strconv.FormatInt(int64(line), 10) + "]"

	return head
}

func (l *FileLogger) Debugf(format string, v ...interface{}) {
	if l.level > 0 {
		return
	}

	head := l.GetHead()

	l.logger.Printf("DEBUG "+head+format, v...)
}

func (l *FileLogger) Debug(v ...interface{}) {
	if l.level > 0 {
		return
	}

	head := l.GetHead()

	l.logger.Println("DEBUG"+head, fmt.Sprintln(v...))
}

func (l *FileLogger) Infof(format string, v ...interface{}) {
	if l.level > 1 {
		return
	}

	head := l.GetHead()

	l.logger.Printf("INFO "+head+format, v...)
}

func (l *FileLogger) Info(v ...interface{}) {
	if l.level > 1 {
		return
	}

	head := l.GetHead()

	l.logger.Println("INFO"+head, fmt.Sprintln(v...))
}

func (l *FileLogger) Warnf(format string, v ...interface{}) {
	if l.level > 2 {
		return
	}

	head := l.GetHead()

	l.logger.Printf("WARN "+head+format, v...)
}

func (l *FileLogger) Warn(v ...interface{}) {
	if l.level > 2 {
		return
	}

	head := l.GetHead()

	l.logger.Println("WARN"+head, fmt.Sprintln(v...))
}

func (l *FileLogger) Errorf(format string, v ...interface{}) {
	if l.level > 3 {
		return
	}

	head := l.GetHead()

	l.logger.Printf("ERROR "+head+format, v...)
}

func (l *FileLogger) Error(v ...interface{}) {
	if l.level > 3 {
		return
	}

	head := l.GetHead()

	l.logger.Println("ERROR"+head, fmt.Sprintln(v...))
}

func NewFileLogger(prefix, logfile string, level int) (*FileLogger, error) {
	if err := os.MkdirAll(filepath.Dir(logfile), os.ModePerm); err != nil {
		return nil, err
	}
	var logger *log.Logger
	if "" == logfile || "stdout" == logfile {
		logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	} else {
		fi, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		logger = log.New(fi, "", log.Ldate|log.Ltime)
	}

	return &FileLogger{prefix, logger, level}, nil
}
