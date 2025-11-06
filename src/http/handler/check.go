package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/daniellavrushin/b4/checker"
	"github.com/daniellavrushin/b4/log"
)

func (api *API) RegisterCheckApi() {
	api.mux.HandleFunc("/api/check/start", api.handleStartCheck)
	api.mux.HandleFunc("/api/check/status", api.handleCheckStatus)
	api.mux.HandleFunc("/api/check/cancel", api.handleCancelCheck)
}

func (api *API) handleStartCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	chckCfg := &api.cfg.System.Checker

	var req StartCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Errorf("Failed to decode check request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set defaults
	if req.Timeout <= 0 {
		req.Timeout = time.Duration(chckCfg.TimeoutSeconds) * time.Second
	}
	if req.MaxConcurrent <= 0 {
		req.MaxConcurrent = chckCfg.MaxConcurrent
	}

	// Select domains
	seen := make(map[string]bool)
	domains := []string{}

	for _, d := range append(chckCfg.Domains, api.cfg.Domains.SNIDomains...) {
		if !seen[d] {
			seen[d] = true
			domains = append(domains, d)
		}
	}

	if len(domains) == 0 {

		http.Error(w, "No domains configured for checking", http.StatusBadRequest)
		return
	}

	// Create check suite
	config := checker.CheckConfig{
		CheckURL:      req.CheckURL,
		Timeout:       req.Timeout,
		MaxConcurrent: req.MaxConcurrent,
	}

	suite := checker.NewCheckSuite(config)

	// Run tests asynchronously
	go suite.Run(domains)

	log.Infof("Started check suite %s with %d domains", suite.Id, len(domains))

	response := StartCheckResponse{
		Id:          suite.Id,
		TotalChecks: len(domains),
		Message:     "Check suite started",
	}

	setJsonHeader(w)
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

func (api *API) handleCheckStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	testID := r.URL.Query().Get("id")
	if testID == "" {
		http.Error(w, "Check ID required", http.StatusBadRequest)
		return
	}

	suite, ok := checker.GetCheckSuite(testID)
	if !ok {
		http.Error(w, "Check suite not found", http.StatusNotFound)
		return
	}

	snapshot := suite.GetSnapshot()

	setJsonHeader(w)
	json.NewEncoder(w).Encode(snapshot)
}

func (api *API) handleCancelCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	testID := r.URL.Query().Get("id")
	if testID == "" {
		http.Error(w, "Check ID required", http.StatusBadRequest)
		return
	}

	if err := checker.CancelCheckSuite(testID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	log.Infof("Canceled test suite %s", testID)

	setJsonHeader(w)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Check suite canceled",
	})
}
