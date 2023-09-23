package profiler

import (
	"fmt"
	"os"

	"github.com/grafana/pyroscope-go"
)

func New(server, service string) (*pyroscope.Profiler, error) {
	if server == "" {
		return nil, nil
	}
	return pyroscope.Start(pyroscope.Config{
		ApplicationName: fmt.Sprintf("bcdhub.%s", service),
		ServerAddress:   server,
		Tags: map[string]string{
			"hostname": os.Getenv("BCDHUB_SERVICE"),
			"project":  "bcdhub",
			"service":  service,
		},

		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	})
}
