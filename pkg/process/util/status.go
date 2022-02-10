package util

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"runtime"
	"time"

	apiutil "github.com/DataDog/datadog-agent/pkg/api/util"
	ddconfig "github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/metadata/host"
	"github.com/DataDog/datadog-agent/pkg/util"
	"github.com/DataDog/datadog-agent/pkg/version"
)

var httpClient = apiutil.GetClient(false)

// CoreStatus holds core info about the process-agent
type CoreStatus struct {
	AgentVersion string       `json:"version"`
	GoVersion    string       `json:"go_version"`
	Arch         string       `json:"build_arch"`
	Config       ConfigStatus `json:"config"`
	Metadata     host.Payload `json:"metadata"`
}

// ConfigStatus holds config settings from process-agent
type ConfigStatus struct {
	LogLevel string `json:"log_level"`
}

// InfoVersion holds information about process-agent version
type InfoVersion struct {
	Version   string
	GitCommit string
	GitBranch string
	BuildDate string
	GoVersion string
}

// ProcessExpvars holds values fetched from the exp var server
type ProcessExpvars struct {
	Pid        int     `json:"pid"`
	Uptime     int     `json:"uptime"`
	UptimeNano float64 `json:"uptime_nano"`
	MemStats   struct {
		Alloc uint64 `json:"alloc"`
	} `json:"memstats"`
	Version             InfoVersion         `json:"version"`
	DockerSocket        string              `json:"docker_socket"`
	LastCollectTime     string              `json:"last_collect_time"`
	ProcessCount        int                 `json:"process_count"`
	ContainerCount      int                 `json:"container_count"`
	ProcessQueueSize    int                 `json:"process_queue_size"`
	RTProcessQueueSize  int                 `json:"rtprocess_queue_size"`
	PodQueueSize        int                 `json:"pod_queue_size"`
	ProcessQueueBytes   int                 `json:"process_queue_bytes"`
	RTProcessQueueBytes int                 `json:"rtprocess_queue_bytes"`
	PodQueueBytes       int                 `json:"pod_queue_bytes"`
	ContainerID         string              `json:"container_id"`
	ProxyURL            string              `json:"proxy_url"`
	LogFile             string              `json:"log_file"`
	EnabledChecks       []string            `json:"enabled_checks"`
	Endpoints           map[string][]string `json:"endpoints"`
}

// Status holds status info from process-agent
type Status struct {
	Date    float64        `json:"date"`
	Core    CoreStatus     `json:"core"`    // Contains the status from the core agent
	Expvars ProcessExpvars `json:"expvars"` // Contains the expvars retrieved from the process agent
}

// StatusOption is a function that acts on a Status object
type StatusOption func(s *Status)

// ConnectionError represents an error to connect to a HTTP server
type ConnectionError struct {
	error
}

// NewConnectionError returns a new ConnectionError
func NewConnectionError(err error) ConnectionError {
	return ConnectionError{err}
}

// OverrideTime overrides the Date from a Status object
func OverrideTime(t time.Time) StatusOption {
	return func(s *Status) {
		s.Date = float64(t.UnixNano())
	}
}

func getCoreStatus() (s CoreStatus) {
	hostnameData, err := util.GetHostnameData(context.TODO())
	var metadata *host.Payload
	if err != nil {
		log.Errorf("Error grabbing hostname for status: %v", err)
		metadata = host.GetPayloadFromCache(context.TODO(), util.HostnameData{Hostname: "unknown", Provider: "unknown"})
	} else {
		metadata = host.GetPayloadFromCache(context.TODO(), hostnameData)
	}

	return CoreStatus{
		AgentVersion: version.AgentVersion,
		GoVersion:    runtime.Version(),
		Arch:         runtime.GOARCH,
		Config: ConfigStatus{
			LogLevel: ddconfig.Datadog.GetString("log_level"),
		},
		Metadata: *metadata,
	}
}

func getExpvars() (s ProcessExpvars, err error) {
	ipcAddr, err := ddconfig.GetIPCAddress()
	if err != nil {
		return ProcessExpvars{}, fmt.Errorf("config error: %s", err.Error())
	}

	port := ddconfig.Datadog.GetInt("process_config.expvar_port")
	if port <= 0 {
		_ = log.Warnf("Invalid process_config.expvar_port -- %d, using default port %d\n", port, ddconfig.DefaultProcessExpVarPort)
		port = ddconfig.DefaultProcessExpVarPort
	}
	expvarEndpoint := fmt.Sprintf("http://%s:%d/debug/vars", ipcAddr, port)
	b, err := apiutil.DoGet(httpClient, expvarEndpoint)
	if err != nil {
		return s, ConnectionError{err}
	}

	err = json.Unmarshal(b, &s)
	return
}

// GetStatus returns a Status object with runtime information about process-agent
func GetStatus() (*Status, error) {
	coreStatus := getCoreStatus()
	processExpVars, err := getExpvars()
	if err != nil {
		return nil, err
	}

	return &Status{
		Date:    float64(time.Now().UnixNano()),
		Core:    coreStatus,
		Expvars: processExpVars,
	}, nil
}
