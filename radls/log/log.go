package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	L *zap.SugaredLogger
)

func InitLogger(w io.Writer) {
	logFilePath := filepath.Join(os.TempDir(), "radls.log")
	fmt.Fprintln(w, "Log file path: ", logFilePath)

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
		zapcore.AddSync(w),
		zapcore.InfoLevel,
	)

	core := zapcore.NewTee(fileCore, consoleCore)

	logger := zap.New(core)
	L = logger.Sugar()

	defer L.Sync()
}
