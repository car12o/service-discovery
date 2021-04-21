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
package snapshot

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/protobuf/types/known/anypb"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
)

func makeCluster(clusterName string, upstreamHost string, upstreamPort uint32) *cluster.Cluster {
	return &cluster.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       ptypes.DurationProto(5 * time.Second),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_LOGICAL_DNS},
		LbPolicy:             cluster.Cluster_ROUND_ROBIN,
		LoadAssignment:       makeEndpoint(clusterName, upstreamHost, upstreamPort),
	}
}

func makeEndpoint(clusterName string, upstreamHost string, upstreamPort uint32) *endpoint.ClusterLoadAssignment {
	return &endpoint.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []*endpoint.LocalityLbEndpoints{{
			LbEndpoints: []*endpoint.LbEndpoint{{
				HostIdentifier: &endpoint.LbEndpoint_Endpoint{
					Endpoint: &endpoint.Endpoint{
						Address: &core.Address{
							Address: &core.Address_SocketAddress{
								SocketAddress: &core.SocketAddress{
									Protocol: core.SocketAddress_TCP,
									Address:  upstreamHost,
									PortSpecifier: &core.SocketAddress_PortValue{
										PortValue: upstreamPort,
									},
								},
							},
						},
					},
				},
			}},
		}},
	}
}

func makeHTTPListener(version uint32, listenerPort uint32, clusterName string) *listener.Listener {
	routeName := "local_route"
	listenerName := fmt.Sprintf("listener_%s_%d", clusterName, version)

	// HTTP filter configuration
	manager := &hcm.HttpConnectionManager{
		StatPrefix: "ingress_http",
		CodecType:  hcm.HttpConnectionManager_AUTO,
		RouteSpecifier: &hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: &route.RouteConfiguration{
				Name: routeName,
				VirtualHosts: []*route.VirtualHost{{
					Name:    routeName,
					Domains: []string{"*"},
					Routes: []*route.Route{{
						Match: &route.RouteMatch{
							PathSpecifier: &route.RouteMatch_Prefix{
								Prefix: "/",
							},
						},
						Action: &route.Route_Route{
							Route: &route.RouteAction{
								ClusterSpecifier: &route.RouteAction_Cluster{
									Cluster: clusterName,
								},
							},
						},
					}},
				}},
			},
		},
		HttpFilters: []*hcm.HttpFilter{{
			Name: wellknown.Router,
		}},
	}
	pbst, err := anypb.New(manager)
	if err != nil {
		panic(err)
	}

	return &listener.Listener{
		Name: listenerName,
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  "0.0.0.0",
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: listenerPort,
					},
				},
			},
		},
		FilterChains: []*listener.FilterChain{{
			Filters: []*listener.Filter{{
				Name: wellknown.HTTPConnectionManager,
				ConfigType: &listener.Filter_TypedConfig{
					TypedConfig: pbst,
				},
			}},
		}},
	}
}

func generateSnapshot(config Config) cache.Snapshot {
	var HTTPListeners []types.Resource
	var clusters []types.Resource
	for _, nodeConfig := range config.Nodes {
		clusterName := fmt.Sprintf("cluster_%s", nodeConfig.UpstreamHost)
		HTTPListener := makeHTTPListener(config.Version, uint32(nodeConfig.ListenerPort), clusterName)
		HTTPListeners = append(HTTPListeners, HTTPListener)
		cluster := makeCluster(clusterName, nodeConfig.UpstreamHost, nodeConfig.UpstreamPort)
		clusters = append(clusters, cluster)
	}

	return cache.NewSnapshot(
		fmt.Sprintf("%d", config.Version),
		nil,
		clusters,
		nil,
		HTTPListeners,
		nil,
		nil,
	)
}
