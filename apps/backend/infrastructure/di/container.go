package di

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/dig"

	"github.com/Haya372/ai-trial/backend/domain/event"
	"github.com/Haya372/ai-trial/backend/domain/eventshare"
	"github.com/Haya372/ai-trial/backend/domain/eventsubscription"
	"github.com/Haya372/ai-trial/backend/domain/session"
	"github.com/Haya372/ai-trial/backend/domain/user"
	"github.com/Haya372/ai-trial/backend/infrastructure/db"
	"github.com/Haya372/ai-trial/backend/infrastructure/repository"
	"github.com/Haya372/ai-trial/backend/infrastructure/telemetry"
	"github.com/Haya372/ai-trial/backend/interface/handler"
	mw "github.com/Haya372/ai-trial/backend/interface/middleware"
	"github.com/Haya372/ai-trial/backend/usecase"
	authuc "github.com/Haya372/ai-trial/backend/usecase/auth"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
	eventshareuc "github.com/Haya372/ai-trial/backend/usecase/eventshare"
	eventsubscriptionuc "github.com/Haya372/ai-trial/backend/usecase/eventsubscription"
)

const serviceName = "ai-trial-backend"

func NewContainer(ctx context.Context) (*dig.Container, error) {
	c := dig.New()
	for _, p := range []any{
		func() context.Context { return ctx },
		newTelemetry,
		tracerProviderOf,
		meterProviderOf,
		metricsReaderOf,
		newLogger,
		newPool,
		newTxManager,
		newClock,
		repository.NewUserRepository,
		repository.NewSessionRepository,
		repository.NewEventQueryRepository,
		repository.NewEventRepository,
		repository.NewEventShareRepository,
		repository.NewEventSubscriptionRepository,
		newShareTokenLoader,
		newGetShareByTokenExecutor,
		newSubscribeToShareExecutor,
		newDeleteSubscriptionExecutor,
		newSignupExecutor,
		newLoginExecutor,
		newLogoutExecutor,
		newListEventsExecutor,
		newCreateEventExecutor,
		newUpdateEventExecutor,
		newDeleteEventExecutor,
		newCreateShareExecutor,
		handler.NewHealthHandler,
		handler.NewAuthHandler,
		handler.NewEventHandler,
		handler.NewEventShareHandler,
		handler.NewShareHandler,
		handler.NewEventSubscriptionHandler,
		newRouter,
	} {
		if err := c.Provide(p); err != nil {
			return nil, fmt.Errorf("provide: %w", err)
		}
	}
	return c, nil
}

func newTelemetry(ctx context.Context) (telemetry.Providers, error) {
	p, err := telemetry.Init(ctx, telemetry.Config{
		ServiceName:  serviceName,
		OTLPEndpoint: os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
	})
	if err != nil {
		return telemetry.Providers{}, fmt.Errorf("init telemetry: %w", err)
	}
	return p, nil
}

func tracerProviderOf(p telemetry.Providers) trace.TracerProvider { return p.Tracer }

func meterProviderOf(p telemetry.Providers) metric.MeterProvider { return p.Meter }

func metricsReaderOf(p telemetry.Providers) *sdkmetric.ManualReader { return p.Reader }

func newLogger() *slog.Logger {
	return slog.New(telemetry.NewTraceLogHandler(slog.NewJSONHandler(os.Stdout, nil)))
}

func newPool(ctx context.Context, logger *slog.Logger) (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	pool, err := db.NewPool(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	logger.Info("connected to database")
	return pool, nil
}

func newTxManager(pool *pgxpool.Pool) usecase.TransactionManager {
	return db.NewPgxTxManager(pool)
}

func newClock() usecase.Clock {
	return usecase.RealClock{}
}

func newSignupExecutor(
	ur user.Repository,
	sr session.Repository,
	tx usecase.TransactionManager,
	clock usecase.Clock,
) handler.SignupExecutor {
	return authuc.NewSignupCommand(ur, sr, tx, clock)
}

func newLoginExecutor(ur user.Repository, sr session.Repository, clock usecase.Clock) handler.LoginExecutor {
	return authuc.NewLoginCommand(ur, sr, clock)
}

func newLogoutExecutor(sr session.Repository) handler.LogoutExecutor {
	return authuc.NewLogoutCommand(sr)
}

func newListEventsExecutor(s eventuc.QueryService) handler.ListEventsExecutor {
	return eventuc.NewListEventsQuery(s)
}

func newCreateEventExecutor(r event.Repository) handler.CreateEventExecutor {
	return eventuc.NewCreateEventCommand(r)
}

func newUpdateEventExecutor(r event.Repository, logger *slog.Logger) handler.UpdateEventExecutor {
	return eventuc.NewUpdateEventCommand(r, logger)
}

func newDeleteEventExecutor(r event.Repository, logger *slog.Logger) handler.DeleteEventExecutor {
	return eventuc.NewDeleteEventCommand(r, logger)
}

func newCreateShareExecutor(
	er event.Repository, sr eventshare.Repository, tx usecase.TransactionManager, logger *slog.Logger,
) handler.CreateShareExecutor {
	return eventshareuc.NewCreateShareCommand(er, sr, tx, logger)
}

func newShareTokenLoader(sr eventshare.Repository, qs eventuc.QueryService) *eventshareuc.ShareTokenLoader {
	return eventshareuc.NewShareTokenLoader(sr, qs)
}

func newGetShareByTokenExecutor(
	l *eventshareuc.ShareTokenLoader,
	s eventsubscription.Repository,
) handler.GetShareByTokenExecutor {
	return eventshareuc.NewGetShareByTokenQuery(l, s)
}

func newSubscribeToShareExecutor(
	l *eventshareuc.ShareTokenLoader,
	s eventsubscription.Repository,
) handler.SubscribeToShareExecutor {
	return eventsubscriptionuc.NewSubscribeToShareCommand(l, s)
}

func newDeleteSubscriptionExecutor(
	s eventsubscription.Repository, logger *slog.Logger,
) handler.DeleteSubscriptionExecutor {
	return eventsubscriptionuc.NewDeleteSubscriptionCommand(s, logger)
}

func newRouter(
	health *handler.HealthHandler,
	auth *handler.AuthHandler,
	ev *handler.EventHandler,
	es *handler.EventShareHandler,
	share *handler.ShareHandler,
	sub *handler.EventSubscriptionHandler,
	sessRepo session.Repository,
	userRepo user.Repository,
	logger *slog.Logger,
	tp trace.TracerProvider,
	mp metric.MeterProvider,
	metricsReader *sdkmetric.ManualReader,
) *chi.Mux {
	r := chi.NewRouter()
	r.Use(mw.Tracing(tp))
	r.Use(mw.Metrics(mp))
	r.Use(mw.AccessLog(logger))
	r.Get("/health", health.ServeHTTP)
	r.With(mw.RequireMetricsToken(os.Getenv("METRICS_BEARER_TOKEN"))).
		Handle("/metrics", telemetry.NewMetricsHandler(metricsReader))
	r.Post("/auth/signup", auth.Signup)
	r.Post("/auth/login", auth.Login)
	r.With(mw.RequireAuth(sessRepo, userRepo, logger)).Post("/auth/logout", auth.Logout)
	r.With(mw.RequireAuth(sessRepo, userRepo, logger)).Get("/auth/me", auth.GetMe)
	r.With(mw.RequireAuth(sessRepo, userRepo, logger)).Get("/events", ev.ServeHTTP)
	r.With(mw.RequireAuth(sessRepo, userRepo, logger)).Post("/events", ev.CreateEvent)
	r.With(mw.RequireAuth(sessRepo, userRepo, logger)).Put("/events/{id}", ev.UpdateEvent)
	r.With(mw.RequireAuth(sessRepo, userRepo, logger)).Delete("/events/{id}", ev.DeleteEvent)
	r.With(mw.RequireAuth(sessRepo, userRepo, logger)).Post("/events/{id}/shares", es.CreateShare)
	r.With(mw.OptionalAuth(sessRepo, userRepo, logger)).Get("/shares/{token}", share.GetShareByToken)
	r.With(mw.RequireAuth(sessRepo, userRepo, logger)).Post("/shares/{token}/subscriptions", sub.SubscribeToShare)
	r.With(mw.RequireAuth(sessRepo, userRepo, logger)).Delete("/subscriptions/{id}", sub.DeleteSubscription)
	return r
}
