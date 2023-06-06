// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package workloadmeta

type mockableGrpcListener interface {
	writeEvents(procsToDelete, procsToAdd []*ProcessEntity)
}

var _ mockableGrpcListener = (*grpcListener)(nil)

type grpcListener struct {
	evts chan *ProcessEntity
}

func newGrpcListener() *grpcListener {
	return &grpcListener{
		evts: make(chan *ProcessEntity, 0),
	}
}

func (l *grpcListener) writeEvents(procsToDelete, procsToAdd []*ProcessEntity) {
	// TODO
}
