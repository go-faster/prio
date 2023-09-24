package main

import (
	"context"
	"os/exec"
	"time"

	"github.com/go-faster/errors"
	"github.com/go-faster/schedpolicy"
	"github.com/go-faster/sdk/app"
	"go.uber.org/zap"
)

func main() {
	app.Run(func(ctx context.Context, lg *zap.Logger, m *app.Metrics) (err error) {
		do := func() error {
			cmd := exec.CommandContext(ctx, "/usr/bin/sleep", "10")
			if err := cmd.Start(); err != nil {
				return errors.Wrap(err, "start")
			}
			pid := cmd.Process.Pid
			ticker := time.NewTicker(1 * time.Second)
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			for range ticker.C {
				if err := ctx.Err(); err != nil {
					break
				}
				policy, err := schedpolicy.Get(cmd.Process.Pid)
				if err != nil {
					return errors.Wrap(err, "schedpolicy")
				}
				lg.Info("schedpolicy", zap.Int("pid", pid), zap.Stringer("policy", policy))
			}
			if err := cmd.Process.Kill(); err != nil {
				return errors.Wrap(err, "kill")
			}
			return nil
		}
		if err := do(); err != nil {
			return errors.Wrap(err, "do")
		}
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			if err := do(); err != nil {
				return errors.Wrap(err, "do")
			}
		}
		return nil
	})

}
