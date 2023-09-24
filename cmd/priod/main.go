package main

import (
	"context"
	"io"
	"time"

	"github.com/go-faster/errors"
	"github.com/go-faster/schedpolicy"
	"github.com/go-faster/sdk/app"
	"github.com/go-faster/sdk/zctx"
	"github.com/go-faster/tetragon/api/v1/tetragon"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func main() {
	app.Run(func(ctx context.Context, lg *zap.Logger, m *app.Metrics) (err error) {
		a, err := NewApp(lg, m)
		if err != nil {
			return errors.Wrap(err, "init")
		}
		return a.Run(ctx)
	}, app.WithZapConfig(newZapConfig()))
}

type App struct {
	log       *zap.Logger
	metrics   *app.Metrics
	tracer    trace.Tracer
	processed metric.Int64Counter
}

func NewApp(logger *zap.Logger, metrics *app.Metrics) (*App, error) {
	a := &App{
		metrics: metrics,
		log:     logger,
		tracer:  metrics.TracerProvider().Tracer("org.go-faster.prio"),
	}

	var err error
	m := metrics.MeterProvider().Meter("org.go-faster.prio")
	if a.processed, err = m.Int64Counter("org.go-faster.prio.execs.processed"); err != nil {
		return nil, err
	}

	return a, nil
}

// Handle exec event, setting scheduler mode based on pod labels.
func (a *App) Handle(ctx context.Context, event *tetragon.ProcessExec) error {
	ctx, span := a.tracer.Start(ctx, "Handle")
	defer span.End()
	policyStr, ok := event.GetProcess().GetPod().GetPodLabels()["prio.go-faster.io/policy"]
	if !ok {
		zctx.From(ctx).Warn("No scheduler policy set")
		return nil
	}
	policy, err := schedpolicy.PolicyString(policyStr)
	if err != nil {
		return errors.Wrap(err, "parse policy")
	}
	pid := int(event.GetProcess().GetPid().GetValue())
	if pid == 0 {
		return errors.New("pid is zero")
	}
	if err := schedpolicy.Set(pid, policy, 0); err != nil {
		return errors.Wrap(err, "set policy")
	}
	a.processed.Add(ctx, 1)
	return nil
}

func (a *App) Run(ctx context.Context) error {
	// Read flows from a tetragon server.
	dialCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	const target = "unix:///var/run/tetragon/tetragon.sock"
	zctx.From(ctx).Info("Connecting to tetragon server",
		zap.String("target", target),
	)
	tetragonConn, err := grpc.DialContext(dialCtx, target,
		grpc.WithBlock(),
		grpc.FailOnNonTempDialError(true),
		grpc.WithReturnConnectionError(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return errors.Wrap(err, "failed to dial")
	}

	client := tetragon.NewFineGuidanceSensorsClient(tetragonConn)
	version, err := client.GetVersion(ctx, &tetragon.GetVersionRequest{})
	if err != nil {
		return errors.Wrap(err, "get version")
	}
	a.log.Info("Connected to tetragon server", zap.String("version", version.Version))
	b, err := client.GetEvents(ctx, &tetragon.GetEventsRequest{
		AllowList: []*tetragon.Filter{
			{
				Labels: []string{
					"prio.go-faster.io/managed=true",
				},
				EventSet: []tetragon.EventType{
					tetragon.EventType_PROCESS_EXEC,
				},
			},
		},
	})
	if err != nil {
		return errors.Wrap(err, "get events")
	}
	for {
		resp, err := b.Recv()
		switch err {
		case io.EOF, context.Canceled:
			return nil
		case nil:
			switch resp.EventType() {
			case tetragon.EventType_PROCESS_EXEC:
				if err := a.Handle(ctx, resp.GetProcessExec()); err != nil {
					zctx.From(ctx).Error("Handle",
						zap.String("err", err.Error()),
					)
				}
			default:
				a.log.Warn("Unknown event type", zap.Stringer("type", resp.EventType()))
			}
		default:
			if status.Code(err) == codes.Canceled {
				return nil
			}
			return errors.Wrap(err, "recv")
		}
	}
}
