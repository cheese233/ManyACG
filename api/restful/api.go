package restful

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/krau/ManyACG/api/restful/middleware"
	"github.com/krau/ManyACG/api/restful/routers"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/telegram"

	"github.com/penglongli/gin-metrics/ginmetrics"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Run(ctx context.Context) {
	if config.Cfg.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	if config.Cfg.API.Metrics {
		metrics := ginmetrics.GetMonitor()
		metrics.SetMetricPath("/metrics")
		metrics.Use(r)
	}

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
	if config.Cfg.Debug {
		fmt.Println("Allowing all origins in debug mode")
		corsConfig.AllowAllOrigins = true
	} else {
		corsConfig.AllowOrigins = config.Cfg.API.AllowedOrigins
	}

	r.Use(cors.New(corsConfig))

	middleware.Init()
	v1 := r.Group("/api/v1")
	routers.Init()
	routers.RegisterAllRouters(v1, middleware.JWTAuthMiddleware)
	if config.Cfg.API.ServeStatic {
		r.GET("/", func(c *gin.Context) {
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(fmt.Sprintf(`
				<!DOCTYPE html>
				<html lang="en">
				<head>
					<meta charset="UTF-8">
					<title>API</title>
					<style>
						.vert-center {
							display: flex;
							align-items: center;
							justify-content: center;
							height: 100vh;
						}
					</style>
					<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/semantic-ui@2.4.2/dist/semantic.min.css">
				</head>
				<body>
					<div class="vert-center">
						<div class="ui raised very padded text container segment">
							<h1 class="ui header">Backend Service</h1>
							<p class="ui grey text">A simple api to provide <a href="%s">%s</a></p>
						</div>
					</div>
					<script src="https://code.jquery.com/jquery-3.1.1.min.js"></script>
					<script src="https://cdn.jsdelivr.net/npm/semantic-ui@2.4.2/dist/semantic.min.js"></script>
				</body>
				</html>
			`, config.Cfg.API.AllowedOrigins[0], config.Cfg.API.AllowedOrigins[0])))
		})
	}

	if config.Cfg.Telegram.Token != "" {
		telegram.SetWebHook(ctx, r)
	}

	server := &http.Server{
		Addr:    config.Cfg.API.Address,
		Handler: r,
	}

	go func() {
		<-ctx.Done()
		common.Logger.Info("Shutting down api server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			common.Logger.Fatalf("Failed to shutdown server: %v", err)
		}
		common.Logger.Info("API server stopped")
	}()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			common.Logger.Panicf("Failed to start server: %v", err)
		}
	}()
}
