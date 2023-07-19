// Code generated by cmd/cgo -godefs; DO NOT EDIT.
// cgo -godefs -- -I ../../ebpf/c -I ../../../ebpf/c -fsigned-char types.go

package http2

const (
	maxHTTP2Path = 0xa0
)

type connTuple = struct {
	Saddr_h  uint64
	Saddr_l  uint64
	Daddr_h  uint64
	Daddr_l  uint64
	Sport    uint16
	Dport    uint16
	Netns    uint32
	Pid      uint32
	Metadata uint32
}
type EbpfTx struct {
	Tup                   connTuple
	Response_last_seen    uint64
	Request_started       uint64
	Response_status_code  uint16
	Request_method        uint32
	Path_size             uint8
	Request_end_of_stream bool
	Pad_cgo_0             [6]byte
	Request_path          [160]uint8
}

type StaticTableEnumKey = uint32

const (
	MethodKey StaticTableEnumKey = 0x2
	PathKey   StaticTableEnumKey = 0x4
	StatusKey StaticTableEnumKey = 0x9
)

type StaticTableEnumValue = uint32

const (
	GetValue       StaticTableEnumValue = 0x2
	PostValue      StaticTableEnumValue = 0x3
	EmptyPathValue StaticTableEnumValue = 0x4
	IndexPathValue StaticTableEnumValue = 0x5
	K200Value      StaticTableEnumValue = 0x8
	K204Value      StaticTableEnumValue = 0x9
	K206Value      StaticTableEnumValue = 0xa
	K304Value      StaticTableEnumValue = 0xb
	K400Value      StaticTableEnumValue = 0xc
	K404Value      StaticTableEnumValue = 0xd
	K500Value      StaticTableEnumValue = 0xe
)

type StaticTableValue = struct {
	Key   uint32
	Value uint32
}
