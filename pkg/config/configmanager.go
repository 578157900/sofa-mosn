/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"github.com/alipay/sofa-mosn/pkg/api/v2"
	"github.com/alipay/sofa-mosn/pkg/log"
)

// TODO: The functions in this file is for service discovery, but the function implmentation is not general, should fix it

// dumper provides basic operation with mosn elements, like 'cluster', to write back the config file with dynamic changes
// biz logic operation, like 'clear all subscribe info', should be written in the bridge code, not in config module.
//
// changes dump flow :
//
// biz ops -> bridge module -> config module
//
//  dumped info load flow:
//
// 1. bridge module register key of interesting config(like 'cluster') into config module
// 2. config parser invoke callback functions (if exists) of config key
// 3. bridge module get biz info(like service subscribe/publish, application info) from callback invocations
// 4. biz module(like confreg) get biz info from bridge module directly

// ResetServiceRegistryInfo
// called when reset service registry info received
func ResetServiceRegistryInfo(appInfo v2.ApplicationInfo, subServiceList []string) {
	// reset service info
	config.ServiceRegistry.ServiceAppInfo = v2.ApplicationInfo{
		AntShareCloud: appInfo.AntShareCloud,
		DataCenter:    appInfo.DataCenter,
		AppName:       appInfo.AppName,
		DeployMode:    appInfo.DeployMode,
		MasterSystem:  appInfo.MasterSystem,
		CloudName:     appInfo.CloudName,
		HostMachine:   appInfo.HostMachine,
	}

	// reset servicePubInfo
	config.ServiceRegistry.ServicePubInfo = []v2.PublishInfo{}

	// delete subInfo / dynamic clusters
	RemoveClusterConfig(subServiceList)
}

// AddOrUpdateClusterConfig
// called when add cluster config info received
func AddOrUpdateClusterConfig(clusters []v2.Cluster) {
	addOrUpdateClusterConfig(clusters)
	go dump(true)
}

func addOrUpdateClusterConfig(clusters []v2.Cluster) {
	for _, clusterConfig := range clusters {
		exist := false

		for i := range config.ClusterManager.Clusters {
			// rewrite cluster's info if exist already
			if config.ClusterManager.Clusters[i].Name == clusterConfig.Name {
				config.ClusterManager.Clusters[i] = clusterConfig
				exist = true
				break
			}
		}

		//added cluster if not exist
		if !exist {
			config.ClusterManager.Clusters = append(config.ClusterManager.Clusters, clusterConfig)
		}
	}
}

func RemoveClusterConfig(clusterNames []string) {
	if removeClusterConfig(clusterNames) {
		go dump(true)
	}
}

func removeClusterConfig(clusterNames []string) bool {
	dirty := false
	for _, clusterName := range clusterNames {
		for i, cluster := range config.ClusterManager.Clusters {
			if cluster.Name == clusterName {
				//remove
				config.ClusterManager.Clusters = append(config.ClusterManager.Clusters[:i], config.ClusterManager.Clusters[i+1:]...)
				dirty = true
				break
			}
		}
	}
	return dirty
}

// AddPubInfo
// called when add pub info received
func AddPubInfo(pubInfoAdded map[string]string) {
	for srvName, srvData := range pubInfoAdded {
		exist := false
		srvPubInfo := v2.PublishInfo{
			Pub: v2.PublishContent{
				ServiceName: srvName,
				PubData:     srvData,
			},
		}
		for i := range config.ServiceRegistry.ServicePubInfo {
			// rewrite cluster's info
			if config.ServiceRegistry.ServicePubInfo[i].Pub.ServiceName == srvName {
				config.ServiceRegistry.ServicePubInfo[i] = srvPubInfo
				exist = true
				break
			}
		}

		if !exist {
			config.ServiceRegistry.ServicePubInfo = append(config.ServiceRegistry.ServicePubInfo, srvPubInfo)
		}
	}

	go dump(true)
}

// DelPubInfo
// called when delete publish info received
func DelPubInfo(serviceName string) {
	dirty := false

	for i, srvPubInfo := range config.ServiceRegistry.ServicePubInfo {
		if srvPubInfo.Pub.ServiceName == serviceName {
			//remove
			config.ServiceRegistry.ServicePubInfo = append(config.ServiceRegistry.ServicePubInfo[:i], config.ServiceRegistry.ServicePubInfo[i+1:]...)
			dirty = true
			break
		}
	}

	go dump(dirty)
}

// AddClusterWithRouter is a wrapper of AddOrUpdateCluster and AddOrUpdateRoutersConfig
// use this function to only dump config once
func AddClusterWithRouter(listenername string, clusters []v2.Cluster, routerConfig *v2.RouterConfiguration) {
	addOrUpdateClusterConfig(clusters)
	addOrUpdateRouterConfig(listenername, routerConfig)
	go dump(true)
}

func findListener(listenername string) (v2.Listener, int) {
	// support only one server
	listeners := config.Servers[0].Listeners
	for idx, ln := range listeners {
		if ln.Name == listenername {
			return ln, idx
		}
	}
	return v2.Listener{}, -1
}

func updateListener(idx int, ln v2.Listener) {
	listeners := config.Servers[0].Listeners
	if idx < len(listeners) {
		listeners[idx] = ln
	}
}

// AddOrUpdateRouterConfig update the connection_manager's config
func AddOrUpdateRouterConfig(listenername string, routerConfig *v2.RouterConfiguration) {
	if addOrUpdateRouterConfig(listenername, routerConfig) {
		go dump(true)
	}
}
func addOrUpdateRouterConfig(listenername string, routerConfig *v2.RouterConfiguration) bool {
	ln, idx := findListener(listenername)
	if idx == -1 {
		return false
	}
	// support only one filter chain
	nfs := ln.FilterChains[0].Filters
	filterIndex := -1
	for i, nf := range nfs {
		if nf.Type == v2.CONNECTION_MANAGER {
			filterIndex = i
			break
		}
	}

	if data, err := json.Marshal(routerConfig); err == nil {
		cfg := make(map[string]interface{})
		if err := json.Unmarshal(data, &cfg); err != nil {
			log.DefaultLogger.Errorf("invalid router config, update config failed")
			return false
		}
		filter := v2.Filter{
			Type:   v2.CONNECTION_MANAGER,
			Config: cfg,
		}
		if filterIndex == -1 {
			nfs = append(nfs, filter)
			ln.FilterChains[0].Filters = nfs
			updateListener(idx, ln)
		} else {
			nfs[filterIndex] = filter
		}
		return true
	}
	return false
}

// AddOrUpdateStreamFilters update the stream filters config
func AddOrUpdateStreamFilters(listenername string, typ string, cfg map[string]interface{}) {
	if addOrUpdateStreamFilters(listenername, typ, cfg) {
		go dump(true)
	}
}

func addOrUpdateStreamFilters(listenername string, typ string, cfg map[string]interface{}) bool {
	ln, idx := findListener(listenername)
	if idx == -1 {
		return false
	}
	filterIndex := -1
	for i, sf := range ln.StreamFilters {
		if sf.Type == typ {
			filterIndex = i
			break
		}
	}
	filter := v2.Filter{
		Type:   typ,
		Config: cfg,
	}
	if filterIndex == -1 {
		ln.StreamFilters = append(ln.StreamFilters, filter)
		updateListener(idx, ln)
	} else {
		ln.StreamFilters[filterIndex] = filter
	}
	return true
}
