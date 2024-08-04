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

type readFileFunc func(string) ([]byte, error)

func fetchTCPMax(readFile readFileFunc) (*int64, error) {
	// https://www.kernel.org/doc/Documentation/networking/ip-sysctl.txt
	tcpMemFile := "/proc/sys/net/ipv4/tcp_rmem"
	fileContent, err := readFile(tcpMemFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %s", err.Error())
	}

	splitTcpMem := strings.Split(string(fileContent), "\t")
	tcpMemMax := splitTcpMem[len(splitTcpMem)-1]
	tcpMemMax = strings.TrimRight(tcpMemMax, "\n")

	limit, err := strconv.ParseInt(tcpMemMax, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to process file: %s", err.Error())
	}

	return &limit, nil
}

func memoryUsedTCP(readFile readFileFunc) (*int64, error) {
	// https://github.com/torvalds/linux/blob/v6.9/net/ipv4/proc.c#L60-L63
	sockStatFile := "/proc/net/sockstat"

	bytes, err := readFile(sockStatFile)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(bytes), "\n")
	tcpLine := ""

	for _, line := range lines {
		if strings.HasPrefix(line, "TCP") {
			tcpLine = line
			break
		}
	}

	fields := strings.Fields(tcpLine)
	if len(fields) < 11 {
		return nil, fmt.Errorf("not enough fields in /proc/net/sockstat on the TCP line")
	}
	usedMem, err := strconv.ParseInt(fields[len(fields)-1], 10, 64)

	return &usedMem, err
}
func Stats() (*statsapi.TCPMemStats, error) {

	tcpmem := &statsapi.TCPMemStats{}

	if memMax, err := fetchTCPMax(os.ReadFile); err == nil {
		tcpmem.MaxMem = memMax
	}

	if mem, err := memoryUsedTCP(os.ReadFile); err == nil {
		tcpmem.CurrentMem = mem
	}

	tcpmem.Time = v1.NewTime(time.Now())

	return tcpmem, nil
}
