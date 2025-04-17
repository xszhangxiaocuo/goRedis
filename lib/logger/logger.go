package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// Settings 存储Logger的配置
type Settings struct {
	Path       string `yaml:"path"`        // 日志文件的存储路径
	Name       string `yaml:"name"`        // 日志文件的名称
	Ext        string `yaml:"ext"`         // 日志文件的扩展名
	TimeFormat string `yaml:"time-format"` // 日志文件名中时间的格式
}

type logLevel int

// 输出级别
const (
	DEBUG logLevel = iota
	INFO
	WARNING
	ERROR
	FATAL
)

const (
	flags              = log.LstdFlags // 日志输出的标志
	defaultCallerDepth = 2             // 默认的调用栈深度
	bufferSize         = 1e5           // 日志通道的缓冲大小
)

type logEntry struct {
	msg   string
	level logLevel
}

var (
	levelFlags = []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"} // 日志级别的字符串表示
)

// Logger 日志记录器
type Logger struct {
	logFile   *os.File       // 日志文件
	logger    *log.Logger    // 标准库的logger
	entryChan chan *logEntry // 日志条目channel，用channel来进行日志写入操作，减少阻塞
	entryPool *sync.Pool     // 日志条目的对象池
}

var DefaultLogger = NewStdoutLogger() // 默认的日志记录器，输出到标准输出

// NewStdoutLogger 创建一个将日志输出到标准输出的记录器
func NewStdoutLogger() *Logger {
	logger := &Logger{
		logFile:   nil,
		logger:    log.New(os.Stdout, "", flags),
		entryChan: make(chan *logEntry, bufferSize),
		entryPool: &sync.Pool{
			New: func() interface{} {
				return &logEntry{}
			},
		},
	}
	go func() {
		for e := range logger.entryChan {
			_ = logger.logger.Output(0, e.msg) // 将日志输出到标准输出
			logger.entryPool.Put(e)
		}
	}()
	return logger
}

// NewFileLogger 创建一个将日志输出到标准输出和日志文件的记录器
func NewFileLogger(settings *Settings) (*Logger, error) {
	fileName := fmt.Sprintf("%s-%s.%s",
		settings.Name,
		time.Now().Format(settings.TimeFormat),
		settings.Ext)
	logFile, err := mustOpen(fileName, settings.Path)
	if err != nil {
		return nil, fmt.Errorf("logging.Join err: %s", err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	logger := &Logger{
		logFile:   logFile,
		logger:    log.New(mw, "", flags),
		entryChan: make(chan *logEntry, bufferSize),
		entryPool: &sync.Pool{
			New: func() interface{} {
				return &logEntry{}
			},
		},
	}
	go func() {
		for e := range logger.entryChan {
			if e.msg == "" {
				continue
			}
			logFilename := fmt.Sprintf("%s-%s.%s",
				settings.Name,
				time.Now().Format(settings.TimeFormat),
				settings.Ext)
			if path.Join(settings.Path, logFilename) != logger.logFile.Name() {
				logFile, err := mustOpen(logFilename, settings.Path)
				if err != nil {
					panic("open log " + logFilename + " failed: " + err.Error())
				}
				logger.logFile = logFile
				logger.logger = log.New(io.MultiWriter(os.Stdout, logFile), "", flags)
			}
			_ = logger.logger.Output(0, e.msg) // 将日志输出到标准输出和日志文件
			logger.entryPool.Put(e)
		}
	}()
	return logger, nil
}

// Setup 初始化DefaultLogger
func Setup(settings *Settings) {
	logger, err := NewFileLogger(settings)
	if err != nil {
		panic(err)
	}
	DefaultLogger = logger
}

// Output 将日志消息发送到记录器
func (logger *Logger) Output(level logLevel, callerDepth int, msg string) {
	var formattedMsg string
	_, file, line, ok := runtime.Caller(callerDepth)
	if ok {
		formattedMsg = fmt.Sprintf("[%s][%s:%d] %s", levelFlags[level], filepath.Base(file), line, msg)
	} else {
		formattedMsg = fmt.Sprintf("[%s] %s", levelFlags[level], msg)
	}
	entry := logger.entryPool.Get().(*logEntry)
	entry.msg = formattedMsg
	entry.level = level
	logger.entryChan <- entry
}

// Debug 通过DefaultLogger记录调试信息
func Debug(v ...interface{}) {
	msg := fmt.Sprintln(v...)
	DefaultLogger.Output(DEBUG, defaultCallerDepth, msg)
}

// Debugf 通过DefaultLogger记录调试信息
func Debugf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	DefaultLogger.Output(DEBUG, defaultCallerDepth, msg)
}

// Info 通过DefaultLogger记录一般信息
func Info(v ...interface{}) {
	msg := fmt.Sprintln(v...)
	DefaultLogger.Output(INFO, defaultCallerDepth, msg)
}

// Infof 通过DefaultLogger记录一般信息
func Infof(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	DefaultLogger.Output(INFO, defaultCallerDepth, msg)
}

// Warn 通过DefaultLogger记录警告信息
func Warn(v ...interface{}) {
	msg := fmt.Sprintln(v...)
	DefaultLogger.Output(WARNING, defaultCallerDepth, msg)
}

// Error 通过DefaultLogger记录错误信息
func Error(v ...interface{}) {
	msg := fmt.Sprintln(v...)
	DefaultLogger.Output(ERROR, defaultCallerDepth, msg)
}

// Errorf 通过DefaultLogger记录错误信息
func Errorf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	DefaultLogger.Output(ERROR, defaultCallerDepth, msg)
}

// Fatal 通过DefaultLogger记录致命错误信息，然后停止程序
func Fatal(v ...interface{}) {
	msg := fmt.Sprintln(v...)
	DefaultLogger.Output(FATAL, defaultCallerDepth, msg)
}
