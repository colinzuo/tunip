package logp

import (
	"time"

	"go.uber.org/zap/zapcore"
)

var baseEncodingConfig = zapcore.EncoderConfig{
	TimeKey:        "timestamp",
	LevelKey:       "level",
	NameKey:        "logger",
	CallerKey:      "caller",
	MessageKey:     "message",
	StacktraceKey:  "stacktrace",
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeLevel:    zapcore.LowercaseLevelEncoder,
	EncodeTime:     ISO8601TimeEncoder,
	EncodeDuration: zapcore.NanosDurationEncoder,
	EncodeCaller:   zapcore.ShortCallerEncoder,
	EncodeName:     zapcore.FullNameEncoder,
}

func buildEncoder(cfg Config) zapcore.Encoder {
	if cfg.JSON {
		return zapcore.NewJSONEncoder(jsonEncoderConfig())
	}

	return zapcore.NewConsoleEncoder(consoleEncoderConfig())
}

func jsonEncoderConfig() zapcore.EncoderConfig {
	return baseEncodingConfig
}

func consoleEncoderConfig() zapcore.EncoderConfig {
	c := baseEncodingConfig
	c.EncodeLevel = zapcore.CapitalLevelEncoder
	c.EncodeName = bracketedNameEncoder
	return c
}

func bracketedNameEncoder(loggerName string, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + loggerName + "]")
}

func encodeTimeLayout(t time.Time, layout string, enc zapcore.PrimitiveArrayEncoder) {
	type appendTimeEncoder interface {
		AppendTimeLayout(time.Time, string)
	}

	if enc, ok := enc.(appendTimeEncoder); ok {
		enc.AppendTimeLayout(t, layout)
		return
	}

	enc.AppendString(t.Format(layout))
}

// ISO8601TimeEncoder use official timezone format
func ISO8601TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	encodeTimeLayout(t, "2006-01-02T15:04:05.000-07:00", enc)
}
