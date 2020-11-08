package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/DevopsArtFactory/bigshot/pkg/controller"
	"github.com/DevopsArtFactory/bigshot/pkg/logger"
	"github.com/DevopsArtFactory/bigshot/pkg/parameter"
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
		fmt.Fprintf(w, err.Error())
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
		fmt.Fprintf(w, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	m := map[string]interface{}{
		"body": item,
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
