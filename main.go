package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/exp/slog"
)

type Config struct {
	FlapperVersion string // FLAPPER_VERSION
	VersionFile    string // VERSION_FILE
	VersionPrefix  string // VERSION_PREFIX
	EnvVarPrefix   string // ENV_VAR_PREFIX
	ServerPort     string // SERVER_PORT
}

var cfg Config

func init() {
	cfg = envConfig()
}

func main() {
	slog.Info(fmt.Sprintf("Starting Flapper v%s", cfg.FlapperVersion))
	slog.Info(fmt.Sprintf("Serving at 0.0.0.0:%s", cfg.ServerPort))
	slog.Info(fmt.Sprintf("Serving environment variables at %s", cfg.EnvVarPrefix))
	slog.Info(fmt.Sprintf("Serving version at %s", cfg.VersionPrefix))

	http.HandleFunc(cfg.EnvVarPrefix, publishEnvVars)
	http.HandleFunc(cfg.VersionPrefix, publishVersion)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, nil))
}

type Variable struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

func publishEnvVars(w http.ResponseWriter, r *http.Request) {
	envVars := make([]Variable, 0)
	variables := os.Environ()
	for _, variable := range variables {
		pair := strings.SplitN(variable, "=", 2)
		if strings.HasPrefix(pair[0], "X_") {
			envVars = append(envVars, Variable{Name: pair[1], Enabled: false})
		} else if strings.HasPrefix(variable, "O_") {
			envVars = append(envVars, Variable{Name: pair[1], Enabled: true})
		}
	}
	respJson, err := json.Marshal(envVars)
	if err != nil {
		slog.Error(err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write(respJson)
}

func publishVersion(w http.ResponseWriter, r *http.Request) {
	// add flapper version to object
	version := make(map[string]interface{})
	version["flapper_version"] = cfg.FlapperVersion

	// read and consume version file
	versionFile, err := os.Open(cfg.VersionFile)
	if err != nil {
		slog.Warn("Version file not found", "error", err)
	} else {
		byteValue, err := io.ReadAll(versionFile)
		if err != nil {
			slog.Error("Failed to read version file", "error", err)
			http.Error(w, fmt.Sprintf("Failed to read version file: %v", err), http.StatusInternalServerError)
		}

		var result map[string]interface{}
		err = json.Unmarshal([]byte(byteValue), &result)
		if err != nil {
			slog.Error("Failed to parse version file", "error", err)
			http.Error(w, fmt.Sprintf("Failed to parse version file: %v", err), http.StatusInternalServerError)
		}

		// combine
		for key, value := range result {
			version[key] = value
		}
	}
	defer versionFile.Close()

	respJson, err := json.Marshal(version)
	if err != nil {
		slog.Error("JSON encoding failed: ", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write(respJson)
}

// prepare environment variables
func envConfig() Config {
	cfg := Config{
		FlapperVersion: getEnv("FLAPPER_VERSION", "0.0.0-dev (not set)"),
		VersionFile:    getEnv("VERSION_FILE", "example.json"),
		VersionPrefix:  getEnv("VERSION_PREFIX", "/version"),
		EnvVarPrefix:   getEnv("ENV_VAR_PREFIX", "/env"),
		ServerPort:     getEnv("SERVER_PORT", "8080"),
	}

	if cfg.VersionFile == cfg.EnvVarPrefix {
		log.Fatal("ENV_VAR_PREFIX and VERSION_PREFIX cannot be the same.")
	}
	return cfg
}

// retreives environment variable if it exists, else default
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
