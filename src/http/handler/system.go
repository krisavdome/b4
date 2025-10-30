package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/daniellavrushin/b4/log"
)

func (api *API) RegisterSystemApi() {
	api.mux.HandleFunc("/api/system/restart", api.handleRestart)
	api.mux.HandleFunc("/api/system/info", api.handleSystemInfo)
}

// detectServiceManager determines which service manager is managing B4
func detectServiceManager() string {
	// Check for systemd
	if _, err := os.Stat("/etc/systemd/system/b4.service"); err == nil {
		if _, err := exec.LookPath("systemctl"); err == nil {
			return "systemd"
		}
	}

	// Check for Entware/OpenWRT init script
	if _, err := os.Stat("/opt/etc/init.d/S99b4"); err == nil {
		return "entware"
	}

	// Check for standard init script
	if _, err := os.Stat("/etc/init.d/b4"); err == nil {
		return "init"
	}

	// Check if running as a standalone process (no service manager)
	return "standalone"
}

type RestartResponse struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	ServiceManager string `json:"service_manager"`
	RestartCommand string `json:"restart_command,omitempty"`
}

type SystemInfo struct {
	ServiceManager string `json:"service_manager"`
	OS             string `json:"os"`
	Arch           string `json:"arch"`
	CanRestart     bool   `json:"can_restart"`
}

func (api *API) handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	serviceManager := detectServiceManager()
	canRestart := serviceManager != "standalone"

	info := SystemInfo{
		ServiceManager: serviceManager,
		OS:             runtime.GOOS,
		Arch:           runtime.GOARCH,
		CanRestart:     canRestart,
	}

	setJsonHeader(w)
	json.NewEncoder(w).Encode(info)
}

func (api *API) handleRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	serviceManager := detectServiceManager()
	log.Infof("Restart requested via web UI (service manager: %s)", serviceManager)

	var response RestartResponse
	response.ServiceManager = serviceManager

	switch serviceManager {
	case "systemd":
		response.Success = true
		response.Message = "Restart initiated via systemd"
		response.RestartCommand = "systemctl restart b4"

	case "entware":
		response.Success = true
		response.Message = "Restart initiated via Entware init script"
		response.RestartCommand = "/opt/etc/init.d/S99b4 restart"

	case "init":
		response.Success = true
		response.Message = "Restart initiated via init script"
		response.RestartCommand = "/etc/init.d/b4 restart"

	case "standalone":
		response.Success = false
		response.Message = "Cannot restart: B4 is not running as a service. Please restart manually."
		setJsonHeader(w)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return

	default:
		response.Success = false
		response.Message = "Unknown service manager"
		setJsonHeader(w)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Send response immediately before triggering restart
	setJsonHeader(w)
	json.NewEncoder(w).Encode(response)

	// Flush the response to ensure it's sent
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// Trigger restart in a goroutine with a small delay
	// This allows the HTTP response to be sent before the service stops
	go func() {
		time.Sleep(500 * time.Millisecond)
		log.Infof("Executing restart command: %s", response.RestartCommand)

		var cmd *exec.Cmd
		switch serviceManager {
		case "systemd":
			cmd = exec.Command("systemctl", "restart", "b4")
		case "entware":
			cmd = exec.Command("/opt/etc/init.d/S99b4", "restart")
		case "init":
			cmd = exec.Command("/etc/init.d/b4", "restart")
		}

		if cmd != nil {
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Errorf("Restart command failed: %v\nOutput: %s", err, string(output))
			} else {
				log.Infof("Restart command executed successfully")
			}
		}
	}()
}
