package logger

import (
	"os"
	"path"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var SugarLogger *zap.SugaredLogger

func InitLogger() {
	writerSyncer := getLogWriter()
	var core zapcore.Core
	if viper.GetString("Environment") == "dev" {
		core = zapcore.NewTee(
			zapcore.NewCore(getConsoleEncoder(), zapcore.AddSync(os.Stdout), zapcore.InfoLevel),
			zapcore.NewCore(getFileEncoder(), writerSyncer, zapcore.InfoLevel),
		)
	} else {
		core = zapcore.NewTee(
			zapcore.NewCore(getConsoleEncoder(), zapcore.AddSync(os.Stdout), zapcore.InfoLevel),
			zapcore.NewCore(getFileEncoder(), writerSyncer, zapcore.InfoLevel),
		)
	}
	logger := zap.New(core, zap.AddCaller())
	SugarLogger = logger.Sugar()
}
func getConsoleEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}
func getFileEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}
func getLogWriter() zapcore.WriteSyncer {
	logFilePath := viper.GetString("log.file.path")
	logFileName := viper.GetString("log.file.name")
	logFileMaxSize := viper.GetInt("log.file.maxsize")
	logFileMaxBackups := viper.GetInt("log.file.maxbackup")
	logFileMaxAge := viper.GetInt("log.file.maxage")
	logFile := path.Join(logFilePath, logFileName)
	lumberJackLogger := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    logFileMaxSize,
		MaxBackups: logFileMaxBackups,
		MaxAge:     logFileMaxAge,
		Compress:   true,
		LocalTime:  true,
	}
	return zapcore.AddSync(lumberJackLogger)
}
