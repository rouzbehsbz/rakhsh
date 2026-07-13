package api

import (
	"fmt"
	"net/http"
	apiUtils "rakhsh/internal/api/utils"
	"rakhsh/internal/common"
	"rakhsh/internal/core/client"
	"rakhsh/internal/core/message"

	gzip "github.com/gin-contrib/gzip"

	"github.com/gin-gonic/gin"
)

type RootHandlers struct {
	ClientHandler  *client.ClientHandler
	MessageHandler *message.MessageHandler
}

type Server struct {
	engine *gin.Engine
	server *http.Server

	rootHandlers RootHandlers
}

func NewServer(host string, port uint16, rootHandlers RootHandlers) *Server {
	engine := gin.New()

	s := &Server{
		engine: engine,
		server: &http.Server{
			Addr:    fmt.Sprintf("%s:%d", host, port),
			Handler: engine,
		},
		rootHandlers: rootHandlers,
	}

	s.registerMiddlewares()
	s.registerRoutes()

	return s
}

func (s *Server) Run() error {
	return s.server.ListenAndServe()
}

func (s *Server) registerMiddlewares() {
	s.engine.NoRoute(func(c *gin.Context) {
		apiUtils.SendError(c, common.NotFoundError(""))
	})
	s.engine.NoMethod(func(c *gin.Context) {
		apiUtils.SendError(c, common.ForbiddenError("You're not allowed to access this method"))
	})

	s.engine.Use(
		gzip.Gzip(gzip.DefaultCompression),
		gin.Logger(),
		RecoveryMiddleware(),
		ErrorHandlerMiddleware(),
	)
}

func (s *Server) registerRoutes() {
	s.engine.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	api := s.engine.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			clients := v1.Group("/clients")
			{
				clients.GET("/self", AuthorizationMiddleware(), s.rootHandlers.ClientHandler.GetSelfClientInfoHandler)
			}
			messages := v1.Group("/messages")
			{
				messages.POST("", AuthorizationMiddleware(), s.rootHandlers.MessageHandler.PostMessage)
			}
		}
	}

	webhook := s.engine.Group("/webhook")
	{
		v1 := webhook.Group("/v1")
		{
			clients := v1.Group("/clients")
			{
				clients.POST("/balance", AuthorizationMiddleware(), s.rootHandlers.ClientHandler.ChargeBalanceWebhook)
			}
		}
	}
}
