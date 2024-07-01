//go:build linux
// +build linux

/*
Copyright 2024 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tcpmem

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	statsapi "k8s.io/kubelet/pkg/apis/stats/v1alpha1"
)

// Stats provides basic information about max and current process count
func Stats() (*statsapi.TCPMemStats, error) {
	tcpmem := &statsapi.TCPMemStats{}

	TCPMemFile := "/proc/sys/net/ipv4"

	memMax := int64(-1)
	if content, err := os.ReadFile(TCPMemFile); err == nil {
		if limit, err := strconv.ParseInt(string(content[:len(content)-1]), 10, 64); err == nil {
			memMax = limit
		}
	}
	// Both reads did not fail.
	if memMax >= 0 {
		tcpmem.MaxTCP = &memMax
	}

	if mem, err := runningTaskCount(); err == nil {
		tcpmem.CurrentMem = &mem
	}
	/*
		else {
			var info syscall.Sysinfo_t
			syscall.Sysinfo(&info)
			procs := int64(info.Procs)
			tcpmem.CurrentMem = &procs
		}
	*/

	tcpmem.Time = v1.NewTime(time.Now())

	return tcpmem, nil
}

func runningTaskCount() (int64, error) {
	/*
		sockets: used 555
		TCP: inuse 54 orphan 0 tw 6 alloc 131 mem 282
		UDP: inuse 94 mem 108
		UDPLITE: inuse 0
		RAW: inuse 0
		FRAG: inuse 0 memory 0
	*/
	SockStatFile := "/proc/net/sockstat"

	bytes, err := os.ReadFile(SockStatFile)
	if err != nil {
		return 0, err
	}
	lines := strings.Split(string(bytes), "\n")
	tcp_line := ""

	for _, line := range lines {
		if strings.HasPrefix(line, "TCP") {
			tcp_line = line
			break
		}
	}

	fields := strings.Fields(tcp_line)
	if len(fields) < 11 {
		return 0, fmt.Errorf("not enough fields in /proc/net/sockstat on the TCP line")
	}
	return strconv.ParseInt(fields[len(fields)-1], 10, 64)
}
