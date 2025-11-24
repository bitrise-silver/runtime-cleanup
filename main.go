package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/log"
	"golang.org/x/mod/semver"
)

type Runtime struct {
	Identifier string `json:"identifier"`
	Version    string `json:"version"`
	Deletable  bool   `json:"deletable"`
}

type RuntimeInfo map[string]Runtime

// ConfigsModel ...
type ConfigsModel struct {
	RemoveVersionsLowerThan  string `env:"remove_versions_lower_than"`
	RemoveVersionsHigherThan string `env:"remove_versions_higher_than"`
}

func createConfigsModelFromEnvs() (ConfigsModel, error) {
	var c ConfigsModel
	if err := stepconf.Parse(&c); err != nil {
		return ConfigsModel{}, err
	}

	return c, nil
}

func failf(format string, v ...interface{}) {
	log.Errorf(format, v...)
	os.Exit(1)
}

func RemoveRuntime(version string, udid string) {
	_, err := exec.Command("xcrun", "simctl", "runtime", "delete", udid).Output()
	if err != nil {
		failf(err.Error())
	}
	fmt.Println("Removed runtime: " + version)
}

func main() {
	configs, err := createConfigsModelFromEnvs()
	if err != nil {
		failf(err.Error())
	}
	stepconf.Print(configs)

	lower := configs.RemoveVersionsLowerThan
	upper := configs.RemoveVersionsHigherThan
	vUpper := "v" + upper
	vLower := "v" + lower
	if !semver.IsValid(vUpper) {
		fmt.Fprintf(os.Stderr, "Invalid version:"+upper)
		os.Exit(1)
	}
	if !semver.IsValid(vLower) {
		fmt.Fprintf(os.Stderr, "Invalid version:"+lower)
		os.Exit(1)
	}
	// Read and store runtimes
	output, err := exec.Command("xcrun", "simctl", "runtime", "list", "--json").Output()
	if err != nil {
		failf(err.Error())
	}
	var info RuntimeInfo
	err = json.Unmarshal(output, &info)
	if err != nil {
		failf(err.Error())
	}
	for _, line := range info {
		udid := line.Identifier
		line.Version = "v" + line.Version
		if !line.Deletable {
			continue
		}
		result := semver.Compare(line.Version, vLower)
		if result <= 0 {
			RemoveRuntime(line.Version, udid)
		}
		result = semver.Compare(line.Version, vUpper)
		if result >= 0 {
			RemoveRuntime(line.Version, udid)
		}
	}
}
