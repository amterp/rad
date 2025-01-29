package log

import "go.uber.org/zap"

var (
	L *zap.SugaredLogger
)

func InitLogger() {
	l, _ := zap.NewProduction()
	L = l.Sugar()
	defer L.Sync()
}
