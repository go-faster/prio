package main

import (
	"os"
	"strconv"
	"time"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/term"
)

// consoleColorLevelEncoder is single-character color encoder for zapcore.Level.
func consoleColorLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch l {
	case zapcore.DebugLevel:
		enc.AppendString(color.New(color.FgCyan).Sprint("D"))
	case zapcore.InfoLevel:
		enc.AppendString(color.New(color.FgBlue).Sprint("I"))
	case zapcore.WarnLevel:
		enc.AppendString(color.New(color.FgYellow).Sprint("W"))
	case zapcore.ErrorLevel:
		enc.AppendString(color.New(color.FgRed).Sprint("E"))
	default:
		enc.AppendString("U")
	}
}

// consoleDeltaEncoder colorfully encodes delta from start in seconds and milliseconds.
func consoleDeltaEncoder(now time.Time) zapcore.TimeEncoder {
	return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		duration := t.Sub(now)
		seconds := duration / time.Second
		milliseconds := (duration % time.Second) / time.Millisecond
		secColor := color.New(color.Faint)
		msecColor := color.New(color.FgHiBlack)
		enc.AppendString(secColor.Sprintf("%03d", seconds) + msecColor.Sprintf(".%02d", milliseconds/10))
	}
}

func NewConsole() zap.Config {
	cfg := zap.NewDevelopmentConfig()
	cfg.DisableStacktrace = true
	cfg.DisableCaller = true
	cfg.EncoderConfig.EncodeLevel = consoleColorLevelEncoder
	cfg.EncoderConfig.EncodeTime = consoleDeltaEncoder(time.Now())
	cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	cfg.EncoderConfig.ConsoleSeparator = " "
	cfg.EncoderConfig.EncodeName = func(s string, encoder zapcore.PrimitiveArrayEncoder) {
		name := s
		const maxChars = 6
		if len(name) > maxChars {
			name = name[:maxChars]
		}
		format := "%-" + strconv.Itoa(maxChars) + "s"
		encoder.AppendString(color.New(color.FgHiBlue).Sprintf(format, name))
	}
	return cfg
}

func New() zap.Config {
	if term.IsTerminal(int(os.Stderr.Fd())) {
		// Interactive terminal, using console output.
		return NewConsole()
	}

	zapCfg := zap.NewProductionConfig()
	// Not so useful, we have go-faster/errors for this that will
	// show in logs.
	zapCfg.DisableStacktrace = true
	// Human-readable timestamps so ops can read them.
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// Disable sampling.
	zapCfg.Sampling = nil

	return zapCfg
}
