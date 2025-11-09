// src/http/handler/config.go
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/daniellavrushin/b4/config"
	"github.com/daniellavrushin/b4/log"
	"github.com/daniellavrushin/b4/metrics"
	"github.com/daniellavrushin/b4/utils"
)

func (api *API) RegisterConfigApi() {

	api.mux.HandleFunc("/api/config", api.handleConfig)
	api.mux.HandleFunc("/api/config/reset", api.resetConfig)
}

func (a *API) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.getConfig(w)
	case http.MethodPut:
		a.updateConfig(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *API) getConfig(w http.ResponseWriter) {
	setJsonHeader(w)

	totalDomains := 0
	categories := []string{}
	for _, set := range a.cfg.Sets {
		totalDomains += len(set.Targets.DomainsToMatch)

		categories = append(categories, set.Targets.GeoSiteCategories...)

	}
	categoryBreakdown, _ := a.geodataManager.GetCategoryCounts(utils.FilterUniqueStrings(categories))

	response := ConfigResponse{
		Config: a.cfg,
		DomainStats: DomainStatistics{
			TotalDomains:      totalDomains,
			GeositeAvailable:  a.geodataManager.IsGeositeConfigured(),
			GeoipAvailable:    a.geodataManager.IsGeoipConfigured(),
			CategoryBreakdown: categoryBreakdown,
		},
	}

	configCopy := *a.cfg
	response.Config = &configCopy

	enc := json.NewEncoder(w)
	_ = enc.Encode(response)
}

func (a *API) updateConfig(w http.ResponseWriter, r *http.Request) {
	var newConfig config.Config

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&newConfig); err != nil {
		log.Errorf("Failed to decode config update: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := newConfig.Validate(); err != nil {
		log.Errorf("Invalid configuration: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	newConfig.ConfigPath = a.cfg.ConfigPath

	a.geodataManager.UpdatePaths(newConfig.System.Geo.GeoSitePath, newConfig.System.Geo.GeoIpPath)

	allDomainsCount := 0
	allIpsCount := 0
	categories := []string{}

	if newConfig.System.Geo.GeoSitePath != "" {

		for _, set := range newConfig.Sets {
			_, _, err := newConfig.GetTargetsForSet(set)
			if err != nil {
				log.Errorf("Failed to load domains for set '%s': %v", set.Name, err)
			}

			allDomainsCount += len(set.Targets.DomainsToMatch)
			allIpsCount += len(set.Targets.IPs)
			categories = append(categories, set.Targets.GeoSiteCategories...)
		}

	}

	categoryBreakdown, _ := a.geodataManager.GetCategoryCounts(utils.FilterUniqueStrings(categories))

	if err := a.saveAndPushConfig(&newConfig); err != nil {
		log.Errorf("Failed to update config: %v", err)
		http.Error(w, "Failed to update config", http.StatusInternalServerError)
		return
	}

	m := metrics.GetMetricsCollector()
	m.RecordEvent("info", fmt.Sprintf("Loaded %d domains from geodata across %d sets", allDomainsCount, len(newConfig.Sets)))

	response := map[string]interface{}{
		"success": true,
		"message": "Configuration updated successfully",
		"domain_stats": DomainStatistics{
			TotalDomains:      allDomainsCount,
			TotalIPs:          allIpsCount,
			CategoryBreakdown: categoryBreakdown,
		},
	}

	setJsonHeader(w)
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	_ = enc.Encode(response)
}

func (a *API) resetConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	log.Infof("Config reset requested")

	defaultCfg := config.DefaultConfig
	defaultCfg.System.Checker = a.cfg.System.Checker
	defaultCfg.ConfigPath = a.cfg.ConfigPath
	defaultCfg.System.WebServer.IsEnabled = a.cfg.System.WebServer.IsEnabled

	for _, set := range a.cfg.Sets {
		set.ResetToDefaults()
		_, _, err := defaultCfg.GetTargetsForSet(set)
		if err != nil {
			log.Errorf("Failed to load domains for set '%s': %v", set.Name, err)
		}
		defaultCfg.Sets = append(defaultCfg.Sets, set)
	}

	defaultCfg.MainSet = defaultCfg.Sets[0]

	if err := a.saveAndPushConfig(&defaultCfg); err != nil {
		log.Errorf("Failed to reset config: %v", err)
		http.Error(w, "Failed to reset config", http.StatusInternalServerError)
		return
	}

	setJsonHeader(w)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Configuration reset to defaults (domains and checker preserved)",
	})
}

func (a *API) saveAndPushConfig(cfg *config.Config) error {
	if globalPool != nil {
		err := globalPool.UpdateConfig(cfg)
		if err != nil {
			return fmt.Errorf("failed to update global pool config: %v", err)
		}
	}

	err := cfg.SaveToFile(cfg.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to save config to file: %v", err)
	}

	*a.cfg = *cfg

	return nil
}

type domainStats struct {
	ManualDomains  int
	GeositeDomains int
	TotalDomains   int
}
