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
	EnvKey                 = "env"
	LogLevelKey            = "log_level"
	BugsnagAPIKey          = "bugsnag.api_key" //nolint: gosec // G101 -- This is a key, not a secret.
	BugsnagReleaseStageKey = "bugsnag.release_stage"
	BugsnagLogLevelKey     = "bugsnag.log_level"
)

func NewZapLogger(config ConfigProvider) *zap.Logger {
	config.SetDefault(EnvKey, "dev")
	config.SetDefault(LogLevelKey, "debug")
	config.SetDefault(BugsnagLogLevelKey, "info")
	config.SetDefault(BugsnagReleaseStageKey, "dev")

	logLevel, err := zapcore.ParseLevel(config.GetString(LogLevelKey))
	if err != nil {
		logLevel = zapcore.InfoLevel
	}
	bugsnagLogLevel, err := zapcore.ParseLevel(config.GetString(BugsnagLogLevelKey))
	if err != nil {
		bugsnagLogLevel = zapcore.ErrorLevel
	}
	bugsnagAPIKey := config.GetString(BugsnagAPIKey)

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
			ReleaseStage:    config.GetString(EnvKey),
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
	fields []zapcore.Field
}

func (b *bugsnagCore) With(fields []zapcore.Field) zapcore.Core {
	b.fields = append(b.fields, fields...)
	return b
}

func (b *bugsnagCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if b.Enabled(entry.Level) {
		return ce.AddCore(entry, b)
	}
	return ce
}

func bugsnagSeverity(lvl zapcore.Level) any {
	switch lvl {
	case zapcore.DebugLevel, zapcore.InfoLevel:
		return bugsnag.SeverityInfo
	case zapcore.WarnLevel:
		return bugsnag.SeverityWarning
	default:
		return bugsnag.SeverityError
	}
}

func (b *bugsnagCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	passedError := errors.New(ent.Message, 1)
	meta := bugsnag.MetaData{}

	for _, f := range b.fields {
		if f.Type == zapcore.ErrorType && f.Interface != nil {
			meta.Add("zap", "error", f.Interface)
			passedError = errors.New(f.Interface.(error).Error(), 1)
		}

		enc := zapcore.NewMapObjectEncoder()
		f.AddTo(enc)
		meta.AddStruct(f.Key, enc.Fields)
	}

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
