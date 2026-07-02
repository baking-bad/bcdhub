package profiler

import (
	"fmt"
	"runtime"
	"time"

	"github.com/grafana/pyroscope-go"
)

func New(server, service string) (*pyroscope.Profiler, error) {
	if server == "" {
		return nil, nil
	}

	runtime.SetMutexProfileFraction(0)
	runtime.SetBlockProfileRate(0)

	return pyroscope.Start(pyroscope.Config{
		ApplicationName: fmt.Sprintf("bcd-%s", service),
		ServerAddress:   server,
		Tags: map[string]string{
			"project": "bcdhub",
			"service": service,
		},
		UploadRate: time.Minute,
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
		},
	})
}
