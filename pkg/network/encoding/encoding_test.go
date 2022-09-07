// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package encoding

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"syscall"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	model "github.com/DataDog/agent-payload/v5/process"

	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/network"
	"github.com/DataDog/datadog-agent/pkg/network/dns"
	"github.com/DataDog/datadog-agent/pkg/network/http"
	"github.com/DataDog/datadog-agent/pkg/process/util"
)

type connTag = uint64

// ConnTag constant must be the same for all platform
const (
	tagGnuTLS  connTag = 1 // netebpf.GnuTLS
	tagOpenSSL connTag = 2 // netebpf.OpenSSL
)

var originalConfig = config.Datadog

func restoreGlobalConfig() {
	config.Datadog = originalConfig
}

func newConfig() {
	config.Datadog = config.NewConfig("datadog", "DD", strings.NewReplacer(".", "_"))
	config.InitConfig(config.Datadog)
}

func getExpectedConnections(encodedWithQueryType bool, httpOutBlob []byte) *model.Connections {
	var dnsByDomain map[int32]*model.DNSStats
	var dnsByDomainByQuerytype map[int32]*model.DNSStatsByQueryType

	if encodedWithQueryType {
		dnsByDomainByQuerytype = map[int32]*model.DNSStatsByQueryType{
			0: {
				DnsStatsByQueryType: map[int32]*model.DNSStats{
					int32(dns.TypeA): {
						DnsTimeouts:          0,
						DnsSuccessLatencySum: 0,
						DnsFailureLatencySum: 0,
						DnsCountByRcode:      map[uint32]uint32{0: 1},
					},
				},
			},
		}
	} else {
		dnsByDomain = map[int32]*model.DNSStats{
			0: {
				DnsTimeouts:          0,
				DnsSuccessLatencySum: 0,
				DnsFailureLatencySum: 0,
				DnsCountByRcode:      map[uint32]uint32{0: 1},
			},
		}
	}

	out := &model.Connections{
		Conns: []*model.Connection{
			{
				Laddr:               &model.Addr{Ip: "10.1.1.1", Port: int32(1000)},
				Raddr:               &model.Addr{Ip: "10.2.2.2", Port: int32(9000)},
				LastBytesSent:       2,
				LastBytesReceived:   101,
				LastRetransmits:     201,
				LastTcpEstablished:  1,
				LastTcpClosed:       1,
				LastPacketsSent:     1,
				LastPacketsReceived: 100,
				Rtt:                 uint32(999),
				RttVar:              uint32(88),
				Pid:                 int32(6000),
				NetNS:               7,
				IpTranslation: &model.IPTranslation{
					ReplSrcIP:   "20.1.1.1",
					ReplDstIP:   "20.1.1.1",
					ReplSrcPort: int32(40000),
					ReplDstPort: int32(80),
				},

				Type:                 model.ConnectionType_udp,
				Family:               model.ConnectionFamily_v4,
				Direction:            model.ConnectionDirection_local,
				IsLocalPortEphemeral: model.EphemeralPortState_ephemeralFalse,

				RouteIdx:         0,
				HttpAggregations: httpOutBlob,
			},
			{
				Laddr: &model.Addr{Ip: "10.1.1.1", Port: int32(1000)},
				Raddr: &model.Addr{Ip: "8.8.8.8", Port: int32(53)},

				Type:                 model.ConnectionType_udp,
				Family:               model.ConnectionFamily_v4,
				Direction:            model.ConnectionDirection_outgoing,
				IsLocalPortEphemeral: model.EphemeralPortState_ephemeralTrue,

				DnsCountByRcode:             map[uint32]uint32{0: 1},
				DnsStatsByDomain:            dnsByDomain,
				DnsStatsByDomainByQueryType: dnsByDomainByQuerytype,
				DnsSuccessfulResponses:      1, // TODO: verify why this was needed

				RouteIdx: -1,
			},
			{
				Laddr: &model.Addr{Ip: "::1", Port: int32(1100)},
				Raddr: &model.Addr{Ip: "::2", Port: int32(1200)},

				Type:                 model.ConnectionType_udp,
				Family:               model.ConnectionFamily_v6,
				Direction:            model.ConnectionDirection_incoming,
				IsLocalPortEphemeral: model.EphemeralPortState_ephemeralUnspecified,

				RouteIdx:  -1,
				IntraHost: true,
			},
		},
		Dns: map[string]*model.DNSEntry{
			"172.217.12.145": {Names: []string{"golang.org"}},
		},
		Domains: []string{"foo.com"},
		Routes: []*model.Route{
			{
				Subnet: &model.Subnet{
					Alias: "subnet-foo",
				},
			},
		},
		AgentConfiguration: &model.AgentConfiguration{
			NpmEnabled: false,
			TsmEnabled: false,
		},
		ConnTelemetry: nil,
		ConnTelemetryMap: map[string]int64{
			string(network.MonotonicKprobesTriggered): 456,
			string(network.ConnsBpfMapSize):           10000,
		},
		CompilationTelemetryByAsset: map[string]*model.RuntimeCompilationTelemetry{
			"tracer": {
				RuntimeCompilationEnabled:  true,
				RuntimeCompilationResult:   model.RuntimeCompilationResult_CompilationSuccess,
				RuntimeCompilationDuration: 215,
				KernelHeaderFetchResult:    model.KernelHeaderFetchResult_DownloadedHeadersFound,
			},
		},
		Tags: network.GetStaticTags(1),
	}
	if runtime.GOOS == "linux" {
		out.Conns[1].Tags = []uint32{0}
		out.Conns[1].TagsChecksum = uint32(3241915907)
	}
	return out
}

func getInputData() *network.Connections {
	return &network.Connections{
		BufferedData: network.BufferedData{
			Conns: []network.ConnectionStats{
				{
					Source: util.AddressFromString("10.1.1.1"),
					Dest:   util.AddressFromString("10.2.2.2"),
					Monotonic: network.StatCountersByCookie{
						{
							StatCounters: network.StatCounters{
								SentBytes:   1,
								RecvBytes:   100,
								Retransmits: 201,
							},
						},
					},
					Last: network.StatCounters{
						SentBytes:      2,
						RecvBytes:      101,
						TCPEstablished: 1,
						TCPClosed:      1,
						Retransmits:    201,
						SentPackets:    1,
						RecvPackets:    100,
					},
					LastUpdateEpoch: 50,
					RTT:             999,
					RTTVar:          88,
					Pid:             6000,
					NetNS:           7,
					SPort:           1000,
					DPort:           9000,
					IPTranslation: &network.IPTranslation{
						ReplSrcIP:   util.AddressFromString("20.1.1.1"),
						ReplDstIP:   util.AddressFromString("20.1.1.1"),
						ReplSrcPort: 40000,
						ReplDstPort: 80,
					},

					Type:             network.UDP,
					Family:           network.AFINET,
					Direction:        network.LOCAL,
					SPortIsEphemeral: network.EphemeralFalse,
					Via: &network.Via{
						Subnet: network.Subnet{
							Alias: "subnet-foo",
						},
					},
				},
				{
					Source:           util.AddressFromString("10.1.1.1"),
					Dest:             util.AddressFromString("8.8.8.8"),
					SPort:            1000,
					DPort:            53,
					Type:             network.UDP,
					Family:           network.AFINET,
					Direction:        network.OUTGOING,
					SPortIsEphemeral: network.EphemeralTrue,
					Tags:             uint64(1),
				},
				{
					Source:           util.AddressFromString("::1"),
					Dest:             util.AddressFromString("::2"),
					SPort:            1100,
					DPort:            1200,
					Type:             network.UDP,
					Family:           network.AFINET6,
					Direction:        network.INCOMING,
					SPortIsEphemeral: network.EphemeralUnknown,
					IntraHost:        true,
				},
			},
		},
		DNS: map[util.Address][]dns.Hostname{
			util.AddressFromString("172.217.12.145"): {dns.ToHostname("golang.org")},
		},
		DNSStats: dns.StatsByKeyByNameByType{
			dns.Key{
				ClientIP:   util.AddressFromString("10.1.1.1"),
				ServerIP:   util.AddressFromString("8.8.8.8"),
				ClientPort: uint16(1000),
				Protocol:   syscall.IPPROTO_UDP,
			}: map[dns.Hostname]map[dns.QueryType]dns.Stats{
				dns.ToHostname("foo.com"): {
					dns.TypeA: {
						Timeouts:          0,
						SuccessLatencySum: 0,
						FailureLatencySum: 0,
						CountByRcode:      map[uint32]uint32{0: 1},
					},
				},
			},
		},
		HTTP: map[http.Key]*http.RequestStats{
			http.NewKey(
				util.AddressFromString("20.1.1.1"),
				util.AddressFromString("20.1.1.1"),
				40000,
				80,
				"/testpath",
				true,
				http.MethodGet,
			): {},
		},
		ConnTelemetry: map[network.ConnTelemetryType]int64{
			network.MonotonicKprobesTriggered: 456,
			network.ConnsBpfMapSize:           10000,
		},
		CompilationTelemetryByAsset: map[string]network.RuntimeCompilationTelemetry{
			"tracer": {
				RuntimeCompilationEnabled:  true,
				RuntimeCompilationResult:   1,
				KernelHeaderFetchResult:    4,
				RuntimeCompilationDuration: 215,
			},
		},
	}
}

func TestSerialization(t *testing.T) {
	in := getInputData()

	httpOut := &model.HTTPAggregations{
		EndpointAggregations: []*model.HTTPStats{
			{
				Path:     "/testpath",
				Method:   model.HTTPMethod_Get,
				FullPath: true,
				StatsByResponseStatus: []*model.HTTPStats_Data{
					{
						Count:     0,
						Latencies: nil,
					},
					{
						Count:     0,
						Latencies: nil,
					},
					{
						Count:     0,
						Latencies: nil,
					},
					{
						Count:     0,
						Latencies: nil,
					},
					{
						Count:     0,
						Latencies: nil,
					},
				},
			},
		},
	}

	httpOutBlob, err := proto.Marshal(httpOut)
	require.NoError(t, err)

	roundtrip := func(t *testing.T, contentType string, marshaler Marshaler, unmarshaler Unmarshaler, out *model.Connections) *model.Connections {
		assert.Equal(t, contentType, marshaler.ContentType())

		blob, err := marshaler.Marshal(in)
		require.NoError(t, err)

		result, err := unmarshaler.Unmarshal(blob)
		require.NoError(t, err)

		return result
	}

	t.Run("requesting application/json serialization (no query types)", func(t *testing.T) {
		newConfig()
		defer restoreGlobalConfig()
		config.Datadog.Set("system_probe_config.collect_dns_domains", false)

		out := getExpectedConnections(false, httpOutBlob)
		ct := "application/json"
		result := roundtrip(t, ct, GetMarshaler(ct), GetUnmarshaler(ct), out)

		// fixup: json marshaler encode nil slice as empty
		result.Conns[0].Tags = nil
		result.Conns[2].Tags = nil
		if runtime.GOOS != "linux" {
			result.Conns[1].Tags = nil
			result.Tags = nil
		}
		assert.Equal(t, out, result)
	})

	t.Run("requesting application/json serialization (with query types)", func(t *testing.T) {
		newConfig()
		defer restoreGlobalConfig()
		config.Datadog.Set("system_probe_config.collect_dns_domains", false)
		config.Datadog.Set("network_config.enable_dns_by_querytype", true)

		out := getExpectedConnections(true, httpOutBlob)
		ct := "application/json"
		result := roundtrip(t, ct, GetMarshaler(ct), GetUnmarshaler(ct), out)

		// fixup: json marshaler encode nil slice as empty
		result.Conns[0].Tags = nil
		result.Conns[2].Tags = nil
		if runtime.GOOS != "linux" {
			result.Conns[1].Tags = nil
			result.Tags = nil
		}
		assert.Equal(t, out, result)
	})

	t.Run("requesting empty serialization", func(t *testing.T) {
		newConfig()
		defer restoreGlobalConfig()
		config.Datadog.Set("system_probe_config.collect_dns_domains", false)

		out := getExpectedConnections(false, httpOutBlob)
		// in case we request empty serialization type, default to application/json
		result := roundtrip(t, "application/json", GetMarshaler(""), GetUnmarshaler(""), out)

		// fixup: json marshaler encode nil slice as empty
		result.Conns[0].Tags = nil
		result.Conns[2].Tags = nil
		if runtime.GOOS != "linux" {
			result.Conns[1].Tags = nil
			result.Tags = nil
		}
		assert.Equal(t, out, result)
	})

	t.Run("requesting unsupported serialization format", func(t *testing.T) {
		newConfig()
		defer restoreGlobalConfig()
		config.Datadog.Set("system_probe_config.collect_dns_domains", false)

		out := getExpectedConnections(false, httpOutBlob)
		ct := "application/json"
		// In case we request an unsupported serialization type, we default to application/json
		result := roundtrip(t, ct, GetMarshaler("application/whatever"), GetUnmarshaler(ct), out)

		// fixup: json marshaler encode nil slice as empty
		result.Conns[0].Tags = nil
		result.Conns[2].Tags = nil
		if runtime.GOOS != "linux" {
			result.Conns[1].Tags = nil
			result.Tags = nil
		}
		assert.Equal(t, out, result)
	})

	t.Run("render default values with application/json", func(t *testing.T) {
		marshaler := GetMarshaler("application/json")
		assert.Equal(t, "application/json", marshaler.ContentType())

		// Empty connection batch
		blob, err := marshaler.Marshal(&network.Connections{
			BufferedData: network.BufferedData{
				Conns: []network.ConnectionStats{{}},
			},
		})
		require.NoError(t, err)

		res := struct {
			Conns []map[string]interface{} `json:"conns"`
		}{}
		require.NoError(t, json.Unmarshal(blob, &res))

		require.Len(t, res.Conns, 1)
		// Check that it contains fields even if they are zeroed
		for _, field := range []string{
			"type", "lastBytesSent", "lastBytesReceived", "lastRetransmits",
			"netNS", "family", "direction", "pid",
		} {
			assert.Contains(t, res.Conns[0], field)
		}
	})

	t.Run("requesting application/protobuf serialization (no query types)", func(t *testing.T) {
		newConfig()
		defer restoreGlobalConfig()
		config.Datadog.Set("system_probe_config.collect_dns_domains", false)

		out := getExpectedConnections(false, httpOutBlob)
		ct := "application/protobuf"
		result := roundtrip(t, ct, GetMarshaler(ct), pSerializer, out)
		assert.Equal(t, out, result)
	})

	t.Run("requesting application/protobuf serialization (with query types)", func(t *testing.T) {
		newConfig()
		defer restoreGlobalConfig()
		config.Datadog.Set("system_probe_config.collect_dns_domains", false)
		config.Datadog.Set("network_config.enable_dns_by_querytype", true)

		out := getExpectedConnections(true, httpOutBlob)
		ct := "application/protobuf"
		result := roundtrip(t, ct, GetMarshaler(ct), pSerializer, out)
		assert.Equal(t, out, result)
	})

	t.Run("molecule deserialization (no query types)", func(t *testing.T) {
		newConfig()
		defer restoreGlobalConfig()
		config.Datadog.Set("system_probe_config.collect_dns_domains", false)

		out := getExpectedConnections(false, httpOutBlob)
		ct := "application/protobuf"
		result := roundtrip(t, ct, GetMarshaler(ct), mDeserializer, out)
		assert.Equal(t, out, result)
	})

	t.Run("molecule deserialization (with query types)", func(t *testing.T) {
		newConfig()
		defer restoreGlobalConfig()
		config.Datadog.Set("system_probe_config.collect_dns_domains", false)
		config.Datadog.Set("network_config.enable_dns_by_querytype", true)

		out := getExpectedConnections(true, httpOutBlob)
		ct := "application/protobuf"
		result := roundtrip(t, ct, GetMarshaler(ct), mDeserializer, out)
		assert.Equal(t, out, result)
	})
}

func TestHTTPSerializationWithLocalhostTraffic(t *testing.T) {
	var (
		clientPort = uint16(52800)
		serverPort = uint16(8080)
		localhost  = util.AddressFromString("127.0.0.1")
	)

	var httpReqStats http.RequestStats
	in := &network.Connections{
		BufferedData: network.BufferedData{
			Conns: []network.ConnectionStats{
				{
					Source: localhost,
					Dest:   localhost,
					SPort:  clientPort,
					DPort:  serverPort,
				},
				{
					Source: localhost,
					Dest:   localhost,
					SPort:  serverPort,
					DPort:  clientPort,
				},
			},
		},
		HTTP: map[http.Key]*http.RequestStats{
			http.NewKey(
				localhost,
				localhost,
				clientPort,
				serverPort,
				"/testpath",
				true,
				http.MethodGet,
			): &httpReqStats,
		},
	}

	httpOut := &model.HTTPAggregations{
		EndpointAggregations: []*model.HTTPStats{
			{
				Path:     "/testpath",
				Method:   model.HTTPMethod_Get,
				FullPath: true,
				StatsByResponseStatus: []*model.HTTPStats_Data{
					{Count: 0, Latencies: nil},
					{Count: 0, Latencies: nil},
					{Count: 0, Latencies: nil},
					{Count: 0, Latencies: nil},
					{Count: 0, Latencies: nil},
				},
			},
		},
	}

	httpOutBlob, err := proto.Marshal(httpOut)
	require.NoError(t, err)

	out := &model.Connections{
		Conns: []*model.Connection{
			{
				Laddr:            &model.Addr{Ip: "127.0.0.1", Port: int32(clientPort)},
				Raddr:            &model.Addr{Ip: "127.0.0.1", Port: int32(serverPort)},
				HttpAggregations: httpOutBlob,
				RouteIdx:         -1,
			},
			{
				Laddr:            &model.Addr{Ip: "127.0.0.1", Port: int32(serverPort)},
				Raddr:            &model.Addr{Ip: "127.0.0.1", Port: int32(clientPort)},
				HttpAggregations: httpOutBlob,
				RouteIdx:         -1,
			},
		},
		AgentConfiguration: &model.AgentConfiguration{
			NpmEnabled: false,
			TsmEnabled: false,
		},
	}

	marshaler := GetMarshaler("application/protobuf")
	blob, err := marshaler.Marshal(in)
	require.NoError(t, err)

	unmarshaler := GetUnmarshaler("application/protobuf")
	result, err := unmarshaler.Unmarshal(blob)
	require.NoError(t, err)

	assert.Equal(t, out, result)
}

func TestPooledObjectGarbageRegression(t *testing.T) {
	// This test ensures that no garbage data is accidentally
	// left on pooled Connection objects used during serialization
	httpKey := http.NewKey(
		util.AddressFromString("10.0.15.1"),
		util.AddressFromString("172.217.10.45"),
		60000,
		8080,
		"",
		true,
		http.MethodGet,
	)

	in := &network.Connections{
		BufferedData: network.BufferedData{
			Conns: []network.ConnectionStats{
				{
					Source: util.AddressFromString("10.0.15.1"),
					SPort:  uint16(60000),
					Dest:   util.AddressFromString("172.217.10.45"),
					DPort:  uint16(8080),
				},
			},
		},
	}

	encodeAndDecodeHTTP := func(c *network.Connections) *model.HTTPAggregations {
		marshaler := GetMarshaler("application/protobuf")
		blob, err := marshaler.Marshal(c)
		require.NoError(t, err)

		unmarshaler := GetUnmarshaler("application/protobuf")
		result, err := unmarshaler.Unmarshal(blob)
		require.NoError(t, err)

		httpBlob := result.Conns[0].HttpAggregations
		if httpBlob == nil {
			return nil
		}

		httpOut := new(model.HTTPAggregations)
		err = proto.Unmarshal(httpBlob, httpOut)
		require.NoError(t, err)
		return httpOut
	}

	// Let's alternate between payloads with and without HTTP data
	for i := 0; i < 1000; i++ {
		if (i % 2) == 0 {
			httpKey.Path = http.Path{
				Content:  fmt.Sprintf("/path-%d", i),
				FullPath: true,
			}
			in.HTTP = map[http.Key]*http.RequestStats{httpKey: {}}
			out := encodeAndDecodeHTTP(in)

			require.NotNil(t, out)
			require.Len(t, out.EndpointAggregations, 1)
			require.Equal(t, httpKey.Path.Content, out.EndpointAggregations[0].Path)
		} else {
			// No HTTP data in this payload, so we should never get HTTP data back after the serialization
			in.HTTP = nil
			out := encodeAndDecodeHTTP(in)
			require.Nil(t, out, "expected a nil object, but got garbage")
		}
	}
}

func BenchmarkProtobufDeserialization(b *testing.B) {
	in := getInputData()
	marshaler := GetMarshaler("application/protobuf")
	blob, err := marshaler.Marshal(in)
	require.NoError(b, err)

	b.Run("molecule", func(b *testing.B) {
		ds := mDeserializer
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			c, err := ds.Unmarshal(blob)
			require.NoError(b, err)
			ResetConnections(c)
			connsPool.Put(c)
			runtime.KeepAlive(c)
		}
	})
	b.Run("protobuf", func(b *testing.B) {
		ds := pSerializer
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			c, err := ds.Unmarshal(blob)
			require.NoError(b, err)
			runtime.KeepAlive(c)
		}
	})
}
