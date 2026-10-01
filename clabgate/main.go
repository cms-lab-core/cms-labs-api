// Package main clabgate проект для работы с c9s внутри k8s кластера
package main

import (
	"context"
	"os"
	"time"

	"github.com/cms-lab-core/cms-labs-api/clabgate/app/di"
	_ "github.com/cms-lab-core/cms-labs-api/clabgate/docs" // load API Docs files (Swagger)
	"github.com/cms-lab-core/cms-labs-api/clabgate/pkg/configs"
	"github.com/cms-lab-core/cms-labs-api/clabgate/pkg/kubeconfig"
	"github.com/cms-lab-core/cms-labs-api/clabgate/pkg/middleware"
	"github.com/cms-lab-core/cms-labs-api/clabgate/pkg/routes"
	"github.com/cms-lab-core/cms-labs-api/clabgate/pkg/utils"
	"github.com/cms-lab-core/cms-labs-api/shared/logs"
	fiber "github.com/gofiber/fiber/v2"
	_ "github.com/joho/godotenv/autoload" // load .env file automatically
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

// @title Clabernetes Gate API
// @version 1.0
// @description This is an auto-generated API Docs.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email your@mail.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath /clabgate/api
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	// Define Fiber config.
	config := configs.FiberConfig()
	logs.ZeroLogInit(configs.AppConfig.Debug)

	// Define a new Fiber app with config.
	app := fiber.New(config)

	// Middlewares.
	middleware.FiberMiddleware(app) // Register Fiber's middleware for app.
	routes.FiberRoutes(app)         // Register Fiber's routes for app.
	startSessionReconciler(context.Background())

	// Start server (with or without graceful shutdown).
	if configs.AppConfig.Server.Layer == "dev" {
		utils.StartServer(app)
	} else {
		utils.StartServerWithGracefulShutdown(app)
	}
}

func startSessionReconciler(ctx context.Context) {
	config := configs.AppConfig
	if config.Session.ReconcileSeconds <= 0 || config.CMS.BaseURL == "" || config.CMS.ClientID == "" || config.CMS.Token == "" {
		return
	}
	go func() {
		loggerConf := (&logs.ZeroLoggerConf{}).SetName("tasks.SessionReconciler")
		logger := logs.NewZeroLogger(loggerConf)
		clusterConfig, err := kubeconfig.Load()
		if err != nil {
			logger.Error().Err(err).Msg("initialize reconciler leader election")
			return
		}
		clientset, err := kubernetes.NewForConfig(clusterConfig)
		if err != nil {
			logger.Error().Err(err).Msg("initialize leader election client")
			return
		}
		identity := os.Getenv("POD_NAME")
		if identity == "" {
			identity, _ = os.Hostname()
		}
		namespace := os.Getenv("POD_NAMESPACE")
		if namespace == "" {
			namespace = "default"
		}
		lock := &resourcelock.LeaseLock{
			LeaseMeta:  metav1.ObjectMeta{Name: "clabgate-session-reconciler", Namespace: namespace},
			Client:     clientset.CoordinationV1(),
			LockConfig: resourcelock.ResourceLockConfig{Identity: identity},
		}
		leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
			Lock: lock, LeaseDuration: 30 * time.Second, RenewDeadline: 20 * time.Second, RetryPeriod: 5 * time.Second,
			ReleaseOnCancel: true,
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: func(leaderCtx context.Context) { runSessionReconciler(leaderCtx, loggerConf) },
				OnStoppedLeading: func() { logger.Warn().Msg("session reconciler leadership lost") },
				OnNewLeader: func(current string) {
					if current != identity {
						logger.Info().Str("leader", current).Msg("session reconciler leader elected")
					}
				},
			},
		})
	}()
}

func runSessionReconciler(ctx context.Context, loggerConf *logs.ZeroLoggerConf) {
	logger := logs.NewZeroLogger(loggerConf)
	minInterval := time.Duration(configs.AppConfig.Session.ReconcileSeconds) * time.Second
	maxInterval := time.Duration(configs.AppConfig.Session.ReconcileMaxSeconds) * time.Second
	if maxInterval < minInterval {
		maxInterval = minInterval
	}
	container, err := di.NewDIContainer(loggerConf)
	if err != nil {
		logger.Error().Err(err).Msg("initialize session reconciler")
		return
	}
	defer container.Close()
	uc, err := container.SessionsUC()
	if err != nil {
		logger.Error().Err(err).Msg("initialize session reconciler clients")
		return
	}

	reconcile := func() bool {
		changes, reconcileErr := uc.Reconcile(ctx)
		if reconcileErr != nil {
			logger.Error().Err(reconcileErr).Msg("reconcile sessions")
			return false
		}
		if changes > 0 {
			logger.Info().Int("changes", changes).Msg("sessions reconciled")
		}
		return changes > 0
	}

	interval := minInterval
	if !reconcile() {
		interval = nextReconcileInterval(interval, minInterval, maxInterval, false)
	}
	timer := time.NewTimer(interval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			interval = nextReconcileInterval(interval, minInterval, maxInterval, reconcile())
			timer.Reset(interval)
		}
	}
}

func nextReconcileInterval(current, minimum, maximum time.Duration, changed bool) time.Duration {
	if changed || current < minimum {
		return minimum
	}
	if current >= maximum || current > maximum/2 {
		return maximum
	}
	return current * 2
}
