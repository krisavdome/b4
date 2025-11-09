package handler

import (
	"encoding/json"
	"net/http"

	"github.com/daniellavrushin/b4/log"
)

func (api *API) RegisterGeoipApi() {
	api.mux.HandleFunc("/api/geoip", api.handleGeoIp)
}

func (a *API) handleGeoIp(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.getGeoIpTags(w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *API) getGeoIpTags(w http.ResponseWriter) {

	setJsonHeader(w)
	enc := json.NewEncoder(w)

	if !a.geodataManager.IsGeoipConfigured() {
		log.Tracef("Geoip path is not configured")
		_ = enc.Encode(GeositeResponse{Tags: []string{}})
		return
	}

	tags, err := a.geodataManager.ListCategories(a.geodataManager.GetGeoipPath())
	if err != nil {
		http.Error(w, "Failed to load geoip tags: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := GeositeResponse{
		Tags: tags,
	}

	_ = enc.Encode(response)
}
