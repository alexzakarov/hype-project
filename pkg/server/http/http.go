package http

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	mw_logger "github.com/gofiber/fiber/v2/middleware/logger"
	fiberSwagger "github.com/swaggo/fiber-swagger"
	"main/config"
	"main/docs"
	"main/pkg/logger"
	"main/pkg/utils/common"
	"os"
	"time"
)

const (
	maxHeaderBytes = 1 << 20
	stackSize      = 1 << 10 // 1 KB
	bodyLimit      = "2M"
	readTimeout    = 15 * time.Second
	writeTimeout   = 15 * time.Second
	gzipLevel      = 5
)

type ServerHttp struct {
	Server *fiber.App
	cfg    *config.Config
	logger logger.Logger
}

func NewHttpServer(cfg *config.Config, logger logger.Logger) *ServerHttp {
	server := fiber.New(fiber.Config{
		CaseSensitive: true,
		StrictRouting: false,
		JSONEncoder:   json.Marshal,
		JSONDecoder:   json.Unmarshal,
		ServerHeader:  os.Getenv("SERVER_HEADER"),
		AppName:       os.Getenv("APP_TITLE") + " " + os.Getenv("APP_VERSION"),
		Immutable:     true,
	})

	server.Use(cors.New())
	server.Use(mw_logger.New())

	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Title = "EventSourcing Microservice"
	docs.SwaggerInfo.Description = "EventSourcing CQRS Microservice."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.BasePath = "/api/v1"
	server.Get("/doc/*", fiberSwagger.WrapHandler)

	server.Get("/", func(c *fiber.Ctx) error {
		logger.Infof("Health check RequestID: %d", common.GenNum())
		return c.SendString(cfg.ServiceName)
	})

	/*
		server.Use(jwtware.New(jwtware.Config{
			SigningKey: []byte(s.cfg.Server.APP_SECRET),
		}))

	*/

	return &ServerHttp{
		Server: server,
		cfg:    cfg,
		logger: logger,
	}
}

func (s *ServerHttp) Listen(serve *fiber.App) (err error) {
	port := os.Getenv("PORT")

	if port == "" {
		port = s.cfg.Http.Port
		s.logger.Warnf("defaulting to ports %s", port)
	}
	URI := fmt.Sprintf("%s:%s", "", port)
	if s.cfg.AppEnv == "prod" {
		if err = serve.ListenTLS(URI, s.cfg.Http.SslCertPath, s.cfg.Http.SslKeyPath); err != nil {
			s.logger.Fatalf("Error starting Server with SSL : ", err)
		}
	} else {
		if err = serve.Listen(URI); err != nil {
			s.logger.Fatalf("Error starting Server : ", err)
		}
	}

	s.logger.Infof("Server is listening on PORT: %s", s.cfg.Http.Port)
	return
}
