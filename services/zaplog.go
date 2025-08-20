package services

import (
	"os"
	"runtime/debug"

	"github.com/bugsnag/bugsnag-go/v2"
	"github.com/bugsnag/bugsnag-go/v2/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	ENV                   = "env"
	LOG_LEVEL             = "log_level"
	BUGSNAG_API_KEY       = "bugsnag.api_key"
	BUGSNAG_RELEASE_STAGE = "bugsnag.release_stage"
	BUGSNAG_LOG_LEVEL     = "bugsnag.log_level"
)

func NewZapLogger(config ConfigProvider) *zap.Logger {
	config.SetDefault(ENV, "dev")
	config.SetDefault(LOG_LEVEL, "debug")
	config.SetDefault(BUGSNAG_LOG_LEVEL, "info")
	config.SetDefault(BUGSNAG_RELEASE_STAGE, "dev")

	logLevel, err := zapcore.ParseLevel(config.GetString(LOG_LEVEL))
	if err != nil {
		logLevel = zapcore.InfoLevel
	}
	bugsnagLogLevel, err := zapcore.ParseLevel(config.GetString(BUGSNAG_LOG_LEVEL))
	if err != nil {
		bugsnagLogLevel = zapcore.ErrorLevel
	}
	bugsnagAPIKey := config.GetString(BUGSNAG_API_KEY)

	console := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(os.Stdout),
		logLevel,
	)

	if bugsnagAPIKey != "" {
		moduleName := "main"
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Path != "" {
			moduleName = info.Main.Path
		}

		bugsnag.Configure(bugsnag.Configuration{
			APIKey:          bugsnagAPIKey,
			ReleaseStage:    config.GetString(ENV),
			ProjectPackages: []string{"main", moduleName},
		})

		bc := &bugsnagCore{
			LevelEnabler: bugsnagLogLevel,
		}
		return zap.New(zapcore.NewTee(console, bc))
	}

	return zap.New(console)
}

type bugsnagCore struct {
	zapcore.LevelEnabler
}

func (b *bugsnagCore) With(fields []zapcore.Field) zapcore.Core {
	return b
}

func (b *bugsnagCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if b.Enabled(entry.Level) {
		return ce.AddCore(entry, b)
	}
	return ce
}

type severity interface{}

func bugsnagSeverity(lvl zapcore.Level) severity {
	switch lvl {
	case zapcore.DebugLevel:
		return bugsnag.SeverityInfo
	case zapcore.InfoLevel:
		return bugsnag.SeverityInfo
	case zapcore.WarnLevel:
		return bugsnag.SeverityWarning
	case zapcore.ErrorLevel:
		return bugsnag.SeverityError
	case zapcore.DPanicLevel:
		return bugsnag.SeverityError
	case zapcore.PanicLevel:
		return bugsnag.SeverityError
	case zapcore.FatalLevel:
		return bugsnag.SeverityError
	default:
		return bugsnag.SeverityError
	}
}

func (b *bugsnagCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	var passedError = errors.New(ent.Message, 1)
	meta := bugsnag.MetaData{}

	for _, f := range fields {
		if f.Type == zapcore.ErrorType && f.Interface != nil {
			meta.Add("zap", "error", f.Interface)
			passedError = errors.New(f.Interface.(error).Error(), 1)
		}

		enc := zapcore.NewMapObjectEncoder()
		f.AddTo(enc)
		meta.AddStruct(f.Key, enc.Fields)
	}

	meta.Add("zap", "message", ent.Message)
	meta.Add("zap", "caller", ent.Caller.String())
	meta.Add("zap", "level", ent.Level.String())
	meta.Add("zap", "timestamp", ent.Time.String())

	err := bugsnag.Notify(passedError, bugsnagSeverity(ent.Level), bugsnag.ErrorClass{Name: ent.Message}, meta)

	if err != nil {
		return err
	}

	return nil
}

func (b *bugsnagCore) Sync() error {
	return nil
}
