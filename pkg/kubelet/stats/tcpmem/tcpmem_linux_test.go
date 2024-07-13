package tcpmem

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	utilpointer "k8s.io/utils/ptr"
)

func TestParseTCPMemFile(t *testing.T) {
	testCases := []struct {
		name         string
		readFileFunc readFileFunc
		expected     *int64
		err          error
	}{
		{
			name: "valid ipv4/tcp_mem file",
			readFileFunc: func(_ string) ([]byte, error) {
				return []byte("24027	32036	48054"), nil
			},
			expected: utilpointer.To(int64(48054)),
		},
		{
			name: "invalid ipv4/tcp_mem file",
			readFileFunc: func(_ string) ([]byte, error) {
				return []byte(""), nil
			},
			expected: nil,
			err:      errors.New("failed to process file: strconv.ParseInt: parsing \"\": invalid syntax"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tcpMem, err := fetchTCPMax(tc.readFileFunc)

			if tc.err != nil {
				assert.Equal(t, tc.err.Error(), err.Error())
				assert.Nil(t, tcpMem)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, tcpMem)
			}
		})
	}
}

func TestMemoryUsedTCP(t *testing.T) {

	sockstat := `
	sockets: used 449
TCP: inuse 49 orphan 0 tw 13 alloc 144 mem 10
UDP: inuse 43 mem 64
UDPLITE: inuse 0
RAW: inuse 0
FRAG: inuse 0 memory 0
`
	testCases := []struct {
		name         string
		readFileFunc readFileFunc
		expected     *int64
		err          error
	}{
		{
			name: "valid sockstat file",
			readFileFunc: func(_ string) ([]byte, error) {
				return []byte(sockstat), nil
			},
			expected: utilpointer.To(int64(10)),
		},
		{
			name: "invalid sockstat file",
			readFileFunc: func(_ string) ([]byte, error) {
				return []byte(""), nil
			},
			expected: nil,
			err:      errors.New("not enough fields in /proc/net/sockstat on the TCP line"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runningTasks, err := memoryUsedTCP(tc.readFileFunc)

			if tc.err != nil {
				assert.Equal(t, tc.err.Error(), err.Error())
				assert.Equal(t, tc.expected, runningTasks)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, runningTasks)
			}
		})
	}
}
