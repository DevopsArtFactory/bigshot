/*
Copyright 2020 The bigshot Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"

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

// New creates a new Server struct
func New() Server {
	return Server{
		Router: http.NewServeMux(),
		Config: Config{
			Addr: constants.DefaultServerAddr,
			Port: constants.DefaultServerPort,
		},
	}
}

// Wrapper wraps handler
func Wrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Header
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Add("Content-Type", "application/json")
		h.ServeHTTP(w, r) // call original
	})
}

// SetDefaultSetting setup default setting values
func (s *Server) SetDefaultSetting(logFile string) error {
	logLevel := logrus.GetLevel()
	if err := setLogFormat(logLevel.String(), logFile); err != nil {
		return err
	}

	logrus.Infof("Setup Default Settings")
	logrus.Infof("Log Level : %s", logrus.GetLevel())
	return nil
}

// setLogFormat sets log format for server
func setLogFormat(logLevel, logFile string) error {
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

	if logLevel == logrus.DebugLevel.String() {
		logrus.SetReportCaller(true)
	}

	if len(logFile) > 0 {
		file, err := os.OpenFile("logrus.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			logrus.Info("Failed to log to file, using default stderr")
		} else {
			logrus.SetOutput(file)
			logrus.SetFormatter(&logrus.JSONFormatter{})
		}
	}

	return nil
}

// GetAddr returns the server address
func (s *Server) GetAddr() string {
	return fmt.Sprintf("%s:%d", s.Config.Addr, s.Config.Port)
}
