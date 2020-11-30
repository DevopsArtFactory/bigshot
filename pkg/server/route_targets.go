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
	"encoding/json"
	"errors"
	fmt "fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/DevopsArtFactory/bigshot/pkg/controller"
	"github.com/DevopsArtFactory/bigshot/pkg/logger"
	"github.com/DevopsArtFactory/bigshot/pkg/parameter"
	"github.com/DevopsArtFactory/bigshot/pkg/schema"
)

// HealthCheck does health check
func HealthCheck(w http.ResponseWriter, req *http.Request) {
	err := logger.WriteAndReturn(w, fmt.Sprintf("[%s]health check, %s %s %s %d", time.Now().Format(time.RFC3339), req.RemoteAddr, req.Method, req.Host, req.ContentLength))
	if err != nil {
		logger.WriteError(err)
	}
}

// ListItems retrieves list of bigshot items
func ListItems(w http.ResponseWriter, req *http.Request) {
	items, err := controller.ListItems()
	if err != nil {
		logger.WriteError(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	m := map[string]interface{}{
		"body": items,
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(m)
}

// Run runs a synthetic
func Run(w http.ResponseWriter, req *http.Request) {
	var p parameter.RunParam

	parameter.Decode(req.Body, &p)

	if err := controller.Run(p.Template); err != nil {
		logger.WriteError(err)
	}
}

// RetrieveItemDetails retrieves detailed information about template
func RetrieveItemDetails(w http.ResponseWriter, req *http.Request) {
	template, err := getTemplateID(req.URL.Path, "/detail/")
	if err != nil {
		logger.WriteError(err)
		return
	}

	item, err := controller.GetDetail(template)
	if err != nil {
		logger.WriteError(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	m := map[string]interface{}{
		"body": item,
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(m)
}

// SaveTemplate retrieves detailed information about template
func SaveTemplate(w http.ResponseWriter, req *http.Request) {
	template, err := getTemplateID(req.URL.Path, "/save/template/")
	if err != nil {
		logger.WriteError(err)
		return
	}
	logrus.Infof("save template: %s", template)

	config := schema.Config{}
	err = json.NewDecoder(req.Body).Decode(&config)
	if err != nil {
		logger.WriteError(err)
	}

	config.Name = template
	err = controller.ModifyTemplate(config)
	if err != nil {
		logger.WriteError(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	enableCors(&w)
	w.WriteHeader(http.StatusOK)
}

// VerifyTarget tries to run verification of target configuration
func VerifyTarget(w http.ResponseWriter, req *http.Request) {
	enableCors(&w)
	target := schema.Target{}
	err := json.NewDecoder(req.Body).Decode(&target)
	if err != nil {
		logger.WriteError(err)
		return
	}

	fmt.Println(target)
	fmt.Println(req.ContentLength)

	res, err := controller.RunTargetVerification(target)
	if err != nil {
		logger.WriteError(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Println(res)

	m := map[string]interface{}{
		"body": *res,
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(m)
}

// getTemplateID returns template ID from URL
func getTemplateID(path, prefix string) (string, error) {
	template := path[len(prefix):]

	if len(template) == 0 {
		return template, errors.New("/detail: template is empty")
	}

	return template, nil
}

// enableCors allows cross-origin-header
func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Headers", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "DELETE, POST, GET, OPTIONS")
}
