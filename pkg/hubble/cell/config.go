// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Hubble

package hubblecell

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/spf13/pflag"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	flowpb "github.com/cilium/cilium/api/v1/flow"
	ciliumDefaults "github.com/cilium/cilium/pkg/defaults"
	hubbleDefaults "github.com/cilium/cilium/pkg/hubble/defaults"
	"github.com/cilium/cilium/pkg/hubble/exporter/exporteroption"
	"github.com/cilium/cilium/pkg/hubble/observer/observeroption"
	monitorAPI "github.com/cilium/cilium/pkg/monitor/api"
)

type config struct {
	// EnableHubble specifies whether to enable the hubble server.
	EnableHubble bool `mapstructure:"enable-hubble"`

	// EventBufferCapacity specifies the capacity of Hubble events buffer.
	EventBufferCapacity int `mapstructure:"hubble-event-buffer-capacity"`
	// EventQueueSize specifies the buffer size of the channel to receive
	// monitor events.
	EventQueueSize int `mapstructure:"hubble-event-queue-size"`
	// SkipUnknownCGroupIDs specifies if events with unknown cgroup ids should
	// be skipped.
	SkipUnknownCGroupIDs bool `mapstructure:"hubble-skip-unknown-cgroup-ids"`
	// MonitorEvents specifies Cilium monitor events for Hubble to observe. By
	// default, Hubble observes all monitor events.
	MonitorEvents []string `mapstructure:"hubble-monitor-events"`

	// SocketPath specifies the UNIX domain socket for Hubble server to listen
	// to.
	SocketPath string `mapstructure:"hubble-socket-path"`

	// ListenAddress specifies address for Hubble to listen to.
	ListenAddress string `mapstructure:"hubble-listen-address"`
	// PreferIpv6 controls whether IPv6 or IPv4 addresses should be preferred
	// for communication to agents, if both are available.
	PreferIpv6 bool `mapstructure:"hubble-prefer-ipv6"`
	// DisableServerTLS allows the Hubble server to run on the given listen
	// address without TLS.
	DisableServerTLS bool `mapstructure:"hubble-disable-tls"`
	// ServerTLSCertFile specifies the path to the public key file for the
	// Hubble server. The file must contain PEM encoded data.
	ServerTLSCertFile string `mapstructure:"hubble-tls-cert-file"`
	// ServerTLSKeyFile specifies the path to the private key file for the
	// Hubble server. The file must contain PEM encoded data.
	ServerTLSKeyFile string `mapstructure:"hubble-tls-key-file"`
	// ServerTLSClientCAFiles specifies the path to one or more client CA
	// certificates to use for TLS with mutual authentication (mTLS). The files
	// must contain PEM encoded data.
	ServerTLSClientCAFiles []string `mapstructure:"hubble-tls-client-ca-files"`

	// Metrics specifies enabled metrics and their configuration options.
	Metrics []string `mapstructure:"hubble-metrics"`
	// EnableOpenMetrics enables exporting hubble metrics in OpenMetrics
	// format.
	EnableOpenMetrics bool `mapstructure:"enable-hubble-open-metrics"`

	// MetricsServer specifies the addresses to serve Hubble metrics on.
	MetricsServer string `mapstructure:"hubble-metrics-server"`
	// EnableMetricsServerTLS run the Hubble metrics server on the given listen
	// address with TLS.
	EnableMetricsServerTLS bool `mapstructure:"hubble-metrics-server-enable-tls"`
	// MetricsServerTLSCertFile specifies the path to the public key file for
	// the Hubble metrics server. The file must contain PEM encoded data.
	MetricsServerTLSCertFile string `mapstructure:"hubble-metrics-server-tls-cert-file"`
	// MetricsServerTLSKeyFile specifies the path to the private key file for
	// the Hubble metrics server. The file must contain PEM encoded data.
	MetricsServerTLSKeyFile string `mapstructure:"hubble-metrics-server-tls-key-file"`
	// MetricsServerTLSClientCAFiles specifies the path to one or more client
	// CA certificates to use for TLS with mutual authentication (mTLS) on the
	// Hubble metrics server. The files must contain PEM encoded data.
	MetricsServerTLSClientCAFiles []string `mapstructure:"hubble-metrics-server-tls-client-ca-files"`

	// FlowlogsConfigFilePath specifies the filepath with configuration of
	// hubble flowlogs. e.g. "/etc/cilium/flowlog.yaml".
	FlowlogsConfigFilePath string `mapstructure:"hubble-flowlogs-config-path"`
	// ExportFilePath specifies the filepath to write Hubble events to. e.g.
	// "/var/run/cilium/hubble/events.log".
	ExportFilePath string `mapstructure:"hubble-export-file-path"`
	// ExportFileMaxSizeMB specifies the file size in MB at which to rotate the
	// Hubble export file.
	ExportFileMaxSizeMB int `mapstructure:"hubble-export-file-max-size-mb"`
	// ExportFileMaxBackups specifies the number of rotated files to keep.
	ExportFileMaxBackups int `mapstructure:"hubble-export-file-max-backups"`
	// ExportFileCompress specifies whether rotated files are compressed.
	ExportFileCompress bool `mapstructure:"hubble-export-file-compress"`
	// ExportAllowlist specifies allow list filter use by exporter.
	ExportAllowlist []*flowpb.FlowFilter `mapstructure:"hubble-export-allowlist"`
	// ExportDenylist specifies deny list filter use by exporter.
	ExportDenylist []*flowpb.FlowFilter `mapstructure:"hubble-export-denylist"`
	// ExportFieldmask specifies list of fields to log in exporter.
	ExportFieldmask []string `mapstructure:"hubble-export-fieldmask"`

	// EnableRecorderAPI specifies if the Hubble Recorder API should be served.
	EnableRecorderAPI bool `mapstructure:"enable-hubble-recorder-api"`
	// RecorderStoragePath specifies the directory in which pcap files created
	// via the Hubble Recorder API are stored.
	RecorderStoragePath string `mapstructure:"hubble-recorder-storage-path"`
	// RecorderSinkQueueSize is the queue size for each recorder sink.
	RecorderSinkQueueSize int `mapstructure:"hubble-recorder-sink-queue-size"`
}

var defaultConfig = config{
	EnableHubble: true,
	// Hubble internals (parser, ringbuffer) configuration
	EventBufferCapacity:  observeroption.Default.MaxFlows.AsInt(),
	EventQueueSize:       0, // see getDefaultMonitorQueueSize()
	SkipUnknownCGroupIDs: true,
	MonitorEvents:        []string{},
	// Hubble local server configuration
	SocketPath: hubbleDefaults.SocketPath,
	// Hubble TCP server configuration
	ListenAddress:          "",
	PreferIpv6:             false,
	DisableServerTLS:       false,
	ServerTLSCertFile:      "",
	ServerTLSKeyFile:       "",
	ServerTLSClientCAFiles: []string{},
	// Hubble metrics configuration
	Metrics:           []string{},
	EnableOpenMetrics: false,
	// Hubble metrics server configuration
	MetricsServer:                 "",
	EnableMetricsServerTLS:        false,
	MetricsServerTLSCertFile:      "",
	MetricsServerTLSKeyFile:       "",
	MetricsServerTLSClientCAFiles: []string{},
	// Hubble log export configuration
	FlowlogsConfigFilePath: "",
	ExportFilePath:         exporteroption.Default.Path,
	ExportFileMaxSizeMB:    exporteroption.Default.MaxSizeMB,
	ExportFileMaxBackups:   exporteroption.Default.MaxBackups,
	ExportFileCompress:     exporteroption.Default.Compress,
	ExportAllowlist:        []*flowpb.FlowFilter{},
	ExportDenylist:         []*flowpb.FlowFilter{},
	ExportFieldmask:        []string{},
	// Hubble recorder configuration
	EnableRecorderAPI:     true,
	RecorderStoragePath:   hubbleDefaults.RecorderStoragePath,
	RecorderSinkQueueSize: 1024,
}

func (def config) Flags(flags *pflag.FlagSet) {
	flags.Bool("enable-hubble", def.EnableHubble, "Enable hubble server")
	// Hubble internals (parser, ringbuffer) configuration
	flags.Int("hubble-event-buffer-capacity", def.EventBufferCapacity, "Capacity of Hubble events buffer. The provided value must be one less than an integer power of two and no larger than 65535 (ie: 1, 3, ..., 2047, 4095, ..., 65535)")
	flags.Int("hubble-event-queue-size", def.EventQueueSize, "Buffer size of the channel to receive monitor events.")
	flags.Bool("hubble-skip-unknown-cgroup-ids", def.SkipUnknownCGroupIDs, "Skip Hubble events with unknown cgroup ids")
	flags.StringSlice("hubble-monitor-events", def.MonitorEvents,
		fmt.Sprintf(
			"Cilium monitor events for Hubble to observe: [%s]. By default, Hubble observes all monitor events.",
			strings.Join(monitorAPI.AllMessageTypeNames(), " "),
		),
	)
	// Hubble local server configuration
	flags.String("hubble-socket-path", def.SocketPath, "Set hubble's socket path to listen for connections")
	// Hubble TCP server configuration
	flags.String("hubble-listen-address", def.ListenAddress, `An additional address for Hubble server to listen to, e.g. ":4244"`)
	flags.Bool("hubble-prefer-ipv6", def.PreferIpv6, "Prefer IPv6 addresses for announcing nodes when both address types are available.")
	flags.Bool("hubble-disable-tls", def.DisableServerTLS, "Allow Hubble server to run on the given listen address without TLS.")
	flags.String("hubble-tls-cert-file", def.ServerTLSCertFile, "Path to the public key file for the Hubble server. The file must contain PEM encoded data.")
	flags.String("hubble-tls-key-file", def.ServerTLSKeyFile, "Path to the private key file for the Hubble server. The file must contain PEM encoded data.")
	flags.StringSlice("hubble-tls-client-ca-files", def.ServerTLSClientCAFiles, "Paths to one or more public key files of client CA certificates to use for TLS with mutual authentication (mTLS). The files must contain PEM encoded data. When provided, this option effectively enables mTLS.")
	flags.StringSlice("hubble-metrics", def.Metrics, "List of Hubble metrics to enable.")
	flags.Bool("enable-hubble-open-metrics", def.EnableOpenMetrics, "Enable exporting hubble metrics in OpenMetrics format")
	// Hubble metrics server configuration
	flags.String("hubble-metrics-server", def.MetricsServer, "Address to serve Hubble metrics on.")
	flags.Bool("hubble-metrics-server-enable-tls", def.EnableMetricsServerTLS, "Run the Hubble metrics server on the given listen address with TLS.")
	flags.String("hubble-metrics-server-tls-cert-file", def.MetricsServerTLSCertFile, "Path to the public key file for the Hubble metrics server. The file must contain PEM encoded data.")
	flags.String("hubble-metrics-server-tls-key-file", def.MetricsServerTLSKeyFile, "Path to the private key file for the Hubble metrics server. The file must contain PEM encoded data.")
	flags.StringSlice("hubble-metrics-server-tls-client-ca-files", def.MetricsServerTLSClientCAFiles, "Paths to one or more public key files of client CA certificates to use for TLS with mutual authentication (mTLS). The files must contain PEM encoded data. When provided, this option effectively enables mTLS.")
	// Hubble log export configuration
	flags.String("hubble-flowlogs-config-path", def.FlowlogsConfigFilePath, "Filepath with configuration of hubble flowlogs")
	flags.String("hubble-export-file-path", def.ExportFilePath, "Filepath to write Hubble events to. By specifying `stdout` the flows are logged instead of written to a rotated file.")
	flags.Int("hubble-export-file-max-size-mb", def.ExportFileMaxSizeMB, "Size in MB at which to rotate Hubble export file.")
	flags.Int("hubble-export-file-max-backups", def.ExportFileMaxBackups, "Number of rotated Hubble export files to keep.")
	flags.Bool("hubble-export-file-compress", def.ExportFileCompress, "Compress rotated Hubble export files.")
	flags.StringSlice("hubble-export-allowlist", []string{}, "Specify allowlist as JSON encoded FlowFilters to Hubble exporter.")
	flags.StringSlice("hubble-export-denylist", []string{}, "Specify denylist as JSON encoded FlowFilters to Hubble exporter.")
	flags.StringSlice("hubble-export-fieldmask", def.ExportFieldmask, "Specify list of fields to use for field mask in Hubble exporter.")
	// Hubble recorder configuration
	flags.Bool("enable-hubble-recorder-api", def.EnableRecorderAPI, "Enable the Hubble recorder API")
	flags.String("hubble-recorder-storage-path", def.RecorderStoragePath, "Directory in which pcap files created via the Hubble Recorder API are stored")
	flags.Int("hubble-recorder-sink-queue-size", def.RecorderSinkQueueSize, "Queue size of each Hubble recorder sink")
}

func (cfg *config) normalize() {
	// Dynamically set the event queue size.
	if cfg.EventQueueSize == 0 {
		cfg.EventQueueSize = getDefaultMonitorQueueSize(runtime.NumCPU())
	}
}

func (cfg config) validate() error {
	if fm := cfg.ExportFieldmask; len(fm) > 0 {
		_, err := fieldmaskpb.New(&flowpb.Flow{}, fm...)
		if err != nil {
			return fmt.Errorf("hubble-export-fieldmask contains invalid fieldmask '%v': %w", fm, err)
		}
	}
	return nil
}

func getDefaultMonitorQueueSize(numCPU int) int {
	monitorQueueSize := numCPU * ciliumDefaults.MonitorQueueSizePerCPU
	if monitorQueueSize > ciliumDefaults.MonitorQueueSizePerCPUMaximum {
		monitorQueueSize = ciliumDefaults.MonitorQueueSizePerCPUMaximum
	}
	return monitorQueueSize
}
