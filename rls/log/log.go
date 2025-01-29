package log

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	L *zap.SugaredLogger
)

func InitLogger() {
	logFilePath := filepath.Join(os.TempDir(), "rls.log")

	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic("failed to open log file: " + err.Error())
	}

	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(logFile),
		zapcore.InfoLevel,
	)

	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(os.Stderr),
		zapcore.InfoLevel,
	)

	core := zapcore.NewTee(fileCore, consoleCore)

	logger := zap.New(core)
	L = logger.Sugar()

	defer L.Sync()
}
