package logger

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	mutex    sync.Mutex
	logLevel int
}

// Уровни логов
const (
	LevelInfo  = 1
	LevelWarn  = 2
	LevelError = 3
	LevelDebug = 4
)

// Цвета ANSI (для терминала)
const (
	ColorReset  = "\033[0m"
	ColorBlue   = "\033[34m"
	ColorYellow = "\033[33m"
	ColorRed    = "\033[31m"
	ColorGray   = "\033[90m"
)

func NewLogger(level int) *Logger {
	return &Logger{logLevel: level}
}

func (logger *Logger) log(level int, color string, levelText, name string, args ...interface{}) {
	if level > logger.logLevel {
		return
	}
	logger.mutex.Lock()
	defer logger.mutex.Unlock()
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	strArgs := make([]string, len(args))
	for i, arg := range args {
		strArgs[i] = fmt.Sprintf("%v", arg) // Преобразуем каждый аргумент в строку
	}
	message := strings.Join(strArgs, " ")
	//message := strings.Join(fmt.Sprint(args...), " ")
	fmt.Printf("%s[SOCKETIO-CLIENT] [%s] [%s] [%s] %s%s\n", color, timestamp, levelText, name, message, ColorReset)
}

// Info выводит информационное сообщение
func (logger *Logger) Info(name string, args ...interface{}) {
	logger.log(LevelInfo, ColorBlue, "INFO", name, args...)
}

// Warn выводит предупреждение
func (logger *Logger) Warn(name string, args ...interface{}) {
	logger.log(LevelWarn, ColorYellow, "WARN", name, args...)
}

// Error выводит ошибку
func (logger *Logger) Error(name string, args ...interface{}) {
	logger.log(LevelError, ColorRed, "ERROR", name, args...)
}

// Debug выводит отладочное сообщение
func (logger *Logger) Debug(name string, args ...interface{}) {
	logger.log(LevelDebug, ColorGray, "DEBUG", name, args...)
}

//// === Пример использования ===
//func main() {
//	logger := NewLogger(LevelDebug)
//
//	logger.Info("Это информационное сообщение.")
//	logger.Warn("Это предупреждение!")
//	logger.Error("Произошла ошибка!")
//	logger.Debug("Отладочное сообщение.")
//}
