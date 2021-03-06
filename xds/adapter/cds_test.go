/*
Copyright 2018 Turbine Labs, Inc.

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

package adapter

import (
	"testing"
	"time"

	envoyapi "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoyauth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	envoycluster "github.com/envoyproxy/go-control-plane/envoy/api/v2/cluster"
	envoycore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/gogo/protobuf/types"

	tbnapi "github.com/turbinelabs/api"
	"github.com/turbinelabs/nonstdlib/ptr"
	"github.com/turbinelabs/rotor/xds/poller"
	"github.com/turbinelabs/test/assert"
)

func checkClusters(t *testing.T, actual cache.Resources, expected []envoyapi.Cluster) {
	assert.Equal(t, len(actual.Items), len(expected))

	actualEnvoyClusters := []envoyapi.Cluster{}
	for _, resource := range actual.Items {
		if cluster, ok := resource.(*envoyapi.Cluster); !ok {
			assert.Failed(t, "could not cast resource to cluster")
		} else {
			actualEnvoyClusters = append(actualEnvoyClusters, *cluster)
		}
	}

	assert.HasSameElements(t, actualEnvoyClusters, expected)
}

func TestEmptyClusters(t *testing.T) {
	objects := poller.MkFixtureObjects()
	objects.Clusters = nil
	s := cds{}

	resources, err := s.resourceAdapter(objects)

	assert.Equal(t, resources.Version, objects.TerribleHash())
	assert.Nil(t, err)
	assert.Equal(t, len(resources.Items), 0)
}

func TestManyClusters(t *testing.T) {
	objects := poller.MkFixtureObjects()
	s := cds{"/etc/tls/ca.pem"}

	resources, err := s.resourceAdapter(objects)

	assert.Equal(t, resources.Version, objects.TerribleHash())
	assert.Nil(t, err)
	assert.NonNil(t, resources.Items)

	expectedEnvoyClusters := []envoyapi.Cluster{
		{
			Name: "foo",
			Type: envoyapi.Cluster_EDS,
			EdsClusterConfig: &envoyapi.Cluster_EdsClusterConfig{
				EdsConfig: &envoycore.ConfigSource{
					ConfigSourceSpecifier: &envoycore.ConfigSource_ApiConfigSource{
						ApiConfigSource: &envoycore.ApiConfigSource{
							ApiType:      envoycore.ApiConfigSource_GRPC,
							ClusterNames: []string{xdsClusterName},
							RefreshDelay: ptr.Duration(xdsRefreshDelaySecs * time.Second),
						},
					},
				},
				ServiceName: "foo",
			},
			ConnectTimeout: clusterConnectTimeoutSecs * time.Second,
			LbPolicy:       envoyapi.Cluster_LEAST_REQUEST,
			TlsContext: &envoyauth.UpstreamTlsContext{
				CommonTlsContext: &envoyauth.CommonTlsContext{
					TlsParams: &envoyauth.TlsParameters{
						TlsMinimumProtocolVersion: envoyauth.TlsParameters_TLS_AUTO,
						TlsMaximumProtocolVersion: envoyauth.TlsParameters_TLS_AUTO,
					},
					ValidationContext: &envoyauth.CertificateValidationContext{
						TrustedCa: &envoycore.DataSource{
							Specifier: &envoycore.DataSource_Filename{
								Filename: "/etc/tls/ca.pem",
							},
						},
					},
				},
				Sni: "foo",
			},
			LbSubsetConfig: &envoyapi.Cluster_LbSubsetConfig{
				FallbackPolicy: envoyapi.Cluster_LbSubsetConfig_ANY_ENDPOINT,
				SubsetSelectors: []*envoyapi.Cluster_LbSubsetConfig_LbSubsetSelector{
					{Keys: []string{"build", "stage", "sw_version"}},
					{Keys: []string{"stage"}},
					{Keys: []string{"stage", "sw_version"}},
				},
			},
			CircuitBreakers: &envoycluster.CircuitBreakers{
				Thresholds: []*envoycluster.CircuitBreakers_Thresholds{
					{
						Priority:   envoycore.RoutingPriority_DEFAULT,
						MaxRetries: &types.UInt32Value{Value: 10},
					},
				},
			},
			OutlierDetection: &envoycluster.OutlierDetection{
				Consecutive_5Xx:                    &types.UInt32Value{Value: 100},
				Interval:                           &types.Duration{Nanos: int32(100 * time.Millisecond)},
				EnforcingConsecutive_5Xx:           &types.UInt32Value{Value: 100},
				EnforcingConsecutiveGatewayFailure: &types.UInt32Value{Value: 0},
				EnforcingSuccessRate:               &types.UInt32Value{Value: 0},
			},
		},
		{
			Name: "baz",
			Type: envoyapi.Cluster_EDS,
			EdsClusterConfig: &envoyapi.Cluster_EdsClusterConfig{
				EdsConfig: &envoycore.ConfigSource{
					ConfigSourceSpecifier: &envoycore.ConfigSource_ApiConfigSource{
						ApiConfigSource: &envoycore.ApiConfigSource{
							ApiType:      envoycore.ApiConfigSource_GRPC,
							ClusterNames: []string{xdsClusterName},
							RefreshDelay: ptr.Duration(xdsRefreshDelaySecs * time.Second),
						},
					},
				},
				ServiceName: "baz",
			},
			ConnectTimeout: clusterConnectTimeoutSecs * time.Second,
			LbPolicy:       envoyapi.Cluster_LEAST_REQUEST,
		},
		{
			Name: "bar",
			Type: envoyapi.Cluster_EDS,
			EdsClusterConfig: &envoyapi.Cluster_EdsClusterConfig{
				EdsConfig: &envoycore.ConfigSource{
					ConfigSourceSpecifier: &envoycore.ConfigSource_ApiConfigSource{
						ApiConfigSource: &envoycore.ApiConfigSource{
							ApiType:      envoycore.ApiConfigSource_GRPC,
							ClusterNames: []string{xdsClusterName},
							RefreshDelay: ptr.Duration(xdsRefreshDelaySecs * time.Second),
						},
					},
				},
				ServiceName: "bar",
			},
			ConnectTimeout: clusterConnectTimeoutSecs * time.Second,
			LbPolicy:       envoyapi.Cluster_LEAST_REQUEST,
			LbSubsetConfig: &envoyapi.Cluster_LbSubsetConfig{
				FallbackPolicy: envoyapi.Cluster_LbSubsetConfig_ANY_ENDPOINT,
				SubsetSelectors: []*envoyapi.Cluster_LbSubsetConfig_LbSubsetSelector{
					{Keys: []string{"build", "stage", "sw_version"}},
					{Keys: []string{"flavor", "stage"}},
					{Keys: []string{"stage"}},
				},
			},
			CircuitBreakers: &envoycluster.CircuitBreakers{
				Thresholds: []*envoycluster.CircuitBreakers_Thresholds{
					{
						Priority:       envoycore.RoutingPriority_DEFAULT,
						MaxConnections: &types.UInt32Value{Value: uint32(2048)},
						MaxRequests:    &types.UInt32Value{Value: uint32(2048)},
					},
				},
			},
			OutlierDetection: &envoycluster.OutlierDetection{
				Interval:                           &types.Duration{Seconds: 30},
				EnforcingConsecutive_5Xx:           &types.UInt32Value{Value: 0},
				EnforcingConsecutiveGatewayFailure: &types.UInt32Value{Value: 0},
				EnforcingSuccessRate:               &types.UInt32Value{Value: 100},
			},
		},
	}

	checkClusters(t, resources, expectedEnvoyClusters)
}

func TestEnvoyToTbnCircuitBreakersReturnsNilForNil(t *testing.T) {
	assert.Nil(t, envoyToTbnCircuitBreakers(nil))
}

func TestEnvoyToTbnCircuitBreakersReturnsNilForHighPriorityThresholds(t *testing.T) {
	input := &envoycluster.CircuitBreakers{
		Thresholds: []*envoycluster.CircuitBreakers_Thresholds{
			{
				Priority:    envoycore.RoutingPriority_HIGH,
				MaxRequests: &types.UInt32Value{Value: uint32(40)},
			},
		},
	}

	assert.Nil(t, envoyToTbnCircuitBreakers(input))
}

func TestEnvoyToTbnCircuitBreakersReturnsFirstDefaultPriority(t *testing.T) {
	input := &envoycluster.CircuitBreakers{
		Thresholds: []*envoycluster.CircuitBreakers_Thresholds{
			{
				Priority:    envoycore.RoutingPriority_HIGH,
				MaxRequests: &types.UInt32Value{Value: uint32(10)},
			},
			{
				Priority:           envoycore.RoutingPriority_DEFAULT,
				MaxPendingRequests: &types.UInt32Value{Value: uint32(20)},
			},
			{
				Priority:   envoycore.RoutingPriority_HIGH,
				MaxRetries: &types.UInt32Value{Value: uint32(30)},
			},
			{
				Priority:    envoycore.RoutingPriority_DEFAULT,
				MaxRequests: &types.UInt32Value{Value: uint32(40)},
			},
		},
	}

	res := envoyToTbnCircuitBreakers(input)
	assert.NonNil(t, res)
	assert.Equal(t, *res.MaxPendingRequests, 20)
}

func TestEnvoyToTbnCircuitBreakers(t *testing.T) {
	input := &envoycluster.CircuitBreakers{
		Thresholds: []*envoycluster.CircuitBreakers_Thresholds{
			{
				Priority:           envoycore.RoutingPriority_DEFAULT,
				MaxConnections:     &types.UInt32Value{Value: uint32(10)},
				MaxPendingRequests: &types.UInt32Value{Value: uint32(20)},
				MaxRetries:         &types.UInt32Value{Value: uint32(30)},
				MaxRequests:        &types.UInt32Value{Value: uint32(40)},
			},
		},
	}

	expected := &tbnapi.CircuitBreakers{
		MaxConnections:     ptr.Int(10),
		MaxPendingRequests: ptr.Int(20),
		MaxRetries:         ptr.Int(30),
		MaxRequests:        ptr.Int(40),
	}
	res := envoyToTbnCircuitBreakers(input)
	assert.NonNil(t, res)
	assert.True(t, tbnapi.CircuitBreakersPtrEquals(res, expected))
}
