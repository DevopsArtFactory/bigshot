package server

import (
	"fmt"
	"github.com/DevopsArtFactory/bigshot/pkg/controller"
	"github.com/DevopsArtFactory/bigshot/pkg/logger"
	"github.com/DevopsArtFactory/bigshot/pkg/parameter"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/DevopsArtFactory/bigshot/pkg/constants"
	"github.com/DevopsArtFactory/bigshot/pkg/schema"
)

type Server struct {
	Config Config
	Router *http.ServeMux
}

type Config struct {
	Addr string
	Port int64
}

type RequestBody struct {
	Config schema.Config `json:"config"`
}

func New() Server {
	return Server{
		Router: http.NewServeMux(),
		Config: Config{
			Addr: constants.DefaultServerAddr,
			Port: constants.DefaultServerPort,
		},
	}
}

// SetRouter sets the router for server
func (s *Server) SetRouter() {
	s.Router.HandleFunc("/health", HealthCheck)
	s.Router.HandleFunc("/run", Run)
	s.Router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		io.WriteString(w, "Welcome to bigshot")
	})
}

// SetDefaultSetting setup default setting values
func (s *Server) SetDefaultSetting() {
	logrus.Infof("Setup Default Settings")
	logrus.Infof("Log Level : %s", constants.DefaultLogLevel)
}

// HealthCheck does health check
func HealthCheck(w http.ResponseWriter, req *http.Request) {
	err := logger.WriteAndReturn(w, fmt.Sprintf("[%s]health check, %s %s %s %d", time.Now().Format(time.RFC3339), req.RemoteAddr, req.Method, req.Host, req.ContentLength))
	if err != nil {
		logger.WriteError(err)
	}
}

// HealthCheck does health check
func Run(w http.ResponseWriter, req *http.Request) {
	var p parameter.RunParam

	parameter.Decode(req.Body, &p)

	if err := controller.Run(p.Template); err != nil {
		logger.WriteError(err)
	}
}

func (s *Server) GetAddr() string {
	return fmt.Sprintf("%s:%d", s.Config.Addr, s.Config.Port)
}
