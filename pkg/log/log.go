package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/alecthomas/kingpin"
	"github.com/sirupsen/logrus"

	"servicemanager/pkg/global"
)

const logFileName = "service-manager.log"

var level string

func init() {
	kingpin.Flag("log-level", "log level").Default("info").StringVar(&level)
}

// log writer
var f io.Writer

func Init() error {
	logPath := filepath.Join(global.CurrDir, logFileName)
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	formatter := &logrus.TextFormatter{
		PadLevelText: true,
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			return frame.Function, fmt.Sprintf("%s:%d:%s", frame.File, frame.Line, frame.Function)
		},
	}
	logrus.SetFormatter(formatter)
	logrus.SetReportCaller(true)
	logrus.SetOutput(f)
	logrus.SetLevel(logLevel())
	return nil
}

func logLevel() logrus.Level {
	switch level {
	case "info":
		return logrus.InfoLevel
	case "debug":
		return logrus.DebugLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	default:
		return logrus.InfoLevel
	}
}

func GetLogWriter() io.Writer {
	return f
}
