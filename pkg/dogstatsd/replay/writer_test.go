// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

package replay

import (
	"io"
	"math/rand"
	"path"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/DataDog/zstd"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"

	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/dogstatsd/packets"
	pb "github.com/DataDog/datadog-agent/pkg/proto/pbgo"
)

func writerTest(t *testing.T, z bool) {
	captureFs.Lock()
	originalFs := captureFs.fs
	captureFs.fs = afero.NewMemMapFs()
	captureFs.Unlock()

	// setup directory
	captureFs.fs.MkdirAll("foo/bar", 0777)

	defer func() {
		captureFs.Lock()
		defer captureFs.Unlock()

		captureFs.fs = originalFs
	}()

	writer := NewTrafficCaptureWriter(1)

	// register pools
	manager := packets.NewPoolManager(packets.NewPool(config.Datadog.GetInt("dogstatsd_buffer_size")))
	oobManager := packets.NewPoolManager(packets.NewPool(32))

	writer.RegisterSharedPoolManager(manager)
	writer.RegisterOOBPoolManager(oobManager)

	var wg sync.WaitGroup
	const (
		iterations   = 100
		testDuration = 5 * time.Second
	)
	sleepDuration := testDuration / iterations
	// For test to fail consistently we need to run with more threads than available CPU
	threads := runtime.NumCPU()
	start := make(chan struct{})
	enqueued := atomic.NewInt32(0)

	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup, threadNo int) {
			defer wg.Done()
			r := rand.New(rand.NewSource(int64(9223372036854 / (threadNo + 1))))
			<-start
			// Add a little bit of controlled jitter so tests fail if enqueuing not correct
			duration := time.Duration(r.Int63n(int64(sleepDuration)))
			time.Sleep(duration)

			for i := 0; i < iterations; i++ {
				buff := CapPool.Get().(*CaptureBuffer)
				pkt := manager.Get().(*packets.Packet)
				pkt.Buffer = []byte("foo.bar|5|#some:tag")
				pkt.Source = packets.UDS
				pkt.Contents = pkt.Buffer

				buff.Pb.Timestamp = time.Now().Unix()
				buff.Buff = pkt
				buff.Pb.Pid = 0
				buff.Pb.AncillarySize = int32(0)
				buff.Pb.PayloadSize = int32(len(pkt.Buffer))
				buff.Pb.Payload = pkt.Buffer // or packet.Buffer[:n] ?

				if writer.Enqueue(buff) {
					enqueued.Inc()
				}
			}

			writer.StopCapture()
		}(&wg, i)
	}

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		close(start)
		writer.Capture("foo/bar", testDuration, z)
	}(&wg)

	wgc := make(chan struct{})
	go func(wg *sync.WaitGroup) {
		defer close(wgc)
		wg.Wait()
	}(&wg)

	<-start
	select {
	case <-wgc:
		break
	case <-time.After(testDuration * 2):
		assert.FailNow(t, "Timed out waiting for capture to finish", "Timeout was: %v", testDuration*2)
	}

	// assert file
	writer.RLock()
	assert.NotNil(t, writer.File)
	assert.False(t, writer.ongoing)

	stats, _ := writer.File.Stat()
	assert.Greater(t, stats.Size(), int64(0))

	var (
		err    error
		buf    []byte
		reader *TrafficCaptureReader
	)

	info, err := writer.File.Stat()
	assert.Nil(t, err)
	fp, err := captureFs.fs.Open(path.Join(writer.Location, info.Name()))
	assert.Nil(t, err)
	buf, err = afero.ReadAll(fp)
	assert.Nil(t, err)
	writer.RUnlock()

	if z {
		buf, err = zstd.Decompress(nil, buf)
		assert.Nil(t, err)
	}

	reader = &TrafficCaptureReader{
		Contents: buf,
		Version:  int(datadogFileVersion),
		Traffic:  make(chan *pb.UnixDogstatsdMsg, 1),
	}

	// file should contain no state as traffic had no ancillary data
	pidMap, entityMap, err := reader.ReadState()
	assert.Nil(t, pidMap)
	assert.Nil(t, entityMap)
	assert.Nil(t, err)

	reader.Lock()
	reader.offset = uint32(len(datadogHeader))
	reader.Unlock()

	var cnt int32
	for _, err = reader.ReadNext(); err != io.EOF; _, err = reader.ReadNext() {
		cnt++
	}
	assert.Equal(t, cnt, enqueued.Load())
	// Expect at least every thread to have enqueued a message
	assert.Greater(t, enqueued.Load(), int32(threads))
}

func TestWriterUncompressed(t *testing.T) {
	writerTest(t, false)
}

func TestWriterCompressed(t *testing.T) {
	writerTest(t, true)
}

func TestValidateLocation(t *testing.T) {
	captureFs.Lock()
	originalFs := captureFs.fs
	captureFs.fs = afero.NewMemMapFs()
	captureFs.Unlock()

	locationBad := "foo/bar"
	locationGood := "bar/quz"

	// setup directory
	captureFs.fs.MkdirAll(locationBad, 0770)
	captureFs.fs.MkdirAll(locationGood, 0776)

	defer func() {
		captureFs.Lock()
		defer captureFs.Unlock()

		captureFs.fs = originalFs
	}()

	_, err := ValidateLocation(locationBad)
	assert.NotNil(t, err)
	l, err := ValidateLocation(locationGood)
	assert.Nil(t, err)
	assert.Equal(t, locationGood, l)
}
