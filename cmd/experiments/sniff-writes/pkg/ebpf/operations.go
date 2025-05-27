package ebpf

import (
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/sniff-writes/pkg/models"
)

func AttachTracepoints(coll *ebpf.Collection, config *models.Config) ([]link.Link, error) {
	links := make([]link.Link, 0)

	operationMap := map[string]bool{}
	for _, op := range config.Operations {
		operationMap[op] = true
	}

	if operationMap["open"] {
		l, err := link.Tracepoint("syscalls", "sys_enter_openat", coll.Programs["trace_openat_enter"], nil)
		if err != nil {
			return nil, err
		}
		links = append(links, l)

		l, err = link.Tracepoint("syscalls", "sys_exit_openat", coll.Programs["trace_openat_exit"], nil)
		if err != nil {
			return nil, err
		}
		links = append(links, l)
	}

	if operationMap["read"] {
		l, err := link.Tracepoint("syscalls", "sys_enter_read", coll.Programs["trace_read_enter"], nil)
		if err != nil {
			return nil, err
		}
		links = append(links, l)

		l, err = link.Tracepoint("syscalls", "sys_exit_read", coll.Programs["trace_read_exit"], nil)
		if err != nil {
			return nil, err
		}
		links = append(links, l)
	}

	if operationMap["write"] {
		l, err := link.Tracepoint("syscalls", "sys_enter_write", coll.Programs["trace_write_enter"], nil)
		if err != nil {
			return nil, err
		}
		links = append(links, l)
	}

	if operationMap["close"] {
		l, err := link.Tracepoint("syscalls", "sys_enter_close", coll.Programs["trace_close_enter"], nil)
		if err != nil {
			return nil, err
		}
		links = append(links, l)
	}

	if operationMap["lseek"] {
		l, err := link.Tracepoint("syscalls", "sys_enter_lseek", coll.Programs["trace_lseek_enter"], nil)
		if err != nil {
			return nil, err
		}
		links = append(links, l)
	}

	return links, nil
}

func ConfigureContentCapture(coll *ebpf.Collection, config *models.Config) error {
	// Set content capture flag in eBPF map
	key := uint32(0)
	value := uint32(0)
	if config.CaptureContent {
		value = uint32(1)
	}

	return coll.Maps["content_capture_enabled"].Put(key, value)
}
