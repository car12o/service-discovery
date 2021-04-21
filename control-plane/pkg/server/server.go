// Copyright 2020 Envoyproxy Authors
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.
package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"

	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	endpointservice "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	runtimeservice "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	secretservice "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	testv3 "github.com/envoyproxy/go-control-plane/pkg/test/v3"
)

type CacheManager interface {
	Start()
	GetCache() cachev3.SnapshotCache
}

type Server struct {
	Port      uint
	Cm        CacheManager
	Callbacks serverv3.Callbacks
}

const (
	grpcMaxConcurrentStreams = 1000000
)

func New(nodeId string, port uint, etcd *clientv3.Client, debug bool) *Server {
	return &Server{
		Port: port,
		Cm:   NewServerCacheManager(nodeId, etcd, debug),
		// TODO: test different callbacks
		Callbacks: &testv3.Callbacks{Debug: debug},
	}
}

func (s *Server) Run() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		log.Fatal(err)
	}

	s.Cm.Start()

	grpcOptions := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams),
	}
	grpcServer := grpc.NewServer(grpcOptions...)

	srv3 := serverv3.NewServer(context.Background(), s.Cm.GetCache(), s.Callbacks)
	registerServer(grpcServer, srv3)

	log.Printf("Management server listening on %d\n", s.Port)
	if err = grpcServer.Serve(lis); err != nil {
		return err
	}

	return nil
}

func registerServer(grpcServer *grpc.Server, server serverv3.Server) {
	discoverygrpc.RegisterAggregatedDiscoveryServiceServer(grpcServer, server)
	endpointservice.RegisterEndpointDiscoveryServiceServer(grpcServer, server)
	clusterservice.RegisterClusterDiscoveryServiceServer(grpcServer, server)
	routeservice.RegisterRouteDiscoveryServiceServer(grpcServer, server)
	listenerservice.RegisterListenerDiscoveryServiceServer(grpcServer, server)
	secretservice.RegisterSecretDiscoveryServiceServer(grpcServer, server)
	runtimeservice.RegisterRuntimeDiscoveryServiceServer(grpcServer, server)
}
