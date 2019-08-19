package util

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *zap.SugaredLogger

func initLogs() {
	// loggerZap, _ := zap.NewProduction() // json 格式输出
	// // loggerZap, _ := zap.NewDevelopment() // format 格式输出
	// defer loggerZap.Sync() // flushes buffer, if any
	// logger = loggerZap.Sugar()

	hook := lumberjack.Logger{
		Filename:   "./logs/speed.log", // 日志文件路径
		MaxSize:    128,                // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: 30,                 // 日志文件最多保存多少个备份
		MaxAge:     7,                  // 文件最多保存多少天
		Compress:   true,               // 是否压缩
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	// 设置日志级别
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zap.DebugLevel)

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),                                           // 编码器配置
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&hook)), // 打印到控制台和文件
		atomicLevel, // 日志级别
	)

	// 开启开发模式，堆栈跟踪
	caller := zap.AddCaller()
	// 开启文件及行号
	development := zap.Development()
	// // 设置初始化字段
	// filed := zap.Fields(zap.String("serviceName", "serviceName"))
	// // 构造日志
	// loggerZap := zap.New(core, caller, development, filed)

	// 构造日志
	loggerZap := zap.New(core, caller, development)

	logger = loggerZap.Sugar()
	// logger.Info("log 初始化成功")
	// logger.Info("无法获取网址",
	// 	zap.String("url", "http://www.baidu.com"),
	// 	zap.Int("attempt", 3),
	// 	zap.Duration("backoff", time.Second))

}

func NewLogger() *zap.SugaredLogger {

	if logger == nil {
		initLogs()
	}

	return logger
}
