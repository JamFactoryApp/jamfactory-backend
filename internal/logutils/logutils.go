package logutils

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func NewDefault() *log.Logger {
	return &log.Logger{
		Out: os.Stdout,
		Formatter: &log.TextFormatter{
			ForceColors:   true,
			FullTimestamp: false,
		},
	}
}

func Log(level log.Level, msg string) {
	switch level {
	case log.PanicLevel:
		log.Panic(msg)
	case log.FatalLevel:
		log.Fatal(msg)
	case log.ErrorLevel:
		log.Error(msg)
	case log.WarnLevel:
		log.Warn(msg)
	case log.InfoLevel:
		log.Info(msg)
	case log.DebugLevel:
		log.Debug(msg)
	case log.TraceLevel:
		log.Trace(msg)
	default:
		log.Print(msg)
	}
}

//func Logf(level log.Level, format string, args ...interface{}) {
//	switch level {
//	case log.PanicLevel:
//		log.Panicf(format, args...)
//	case log.FatalLevel:
//		log.Fatalf(format, args...)
//	case log.ErrorLevel:
//		log.Errorf(format, args...)
//	case log.WarnLevel:
//		log.Warnf(format, args...)
//	case log.InfoLevel:
//		log.Infof(format, args...)
//	case log.DebugLevel:
//		log.Debugf(format, args...)
//	case log.TraceLevel:
//		log.Tracef(format, args...)
//	default:
//		log.Printf(format, args...)
//	}
//}
