/*
 * Copyright (c) 2025, NVIDIA CORPORATION.  All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"encoding/json"
	"os"
)

// DevicePluginConfig holds global and per-node device plugin settings loaded
// from a ConfigMap-mounted JSON file.
type DevicePluginConfig struct {
	// PreConfiguredDeviceMemoryMB is the fallback total GPU memory in MiB used
	// when NVML GetMemoryInfo returns ERROR_NOT_SUPPORTED (e.g. unified memory
	// GPUs like NVIDIA GB10/DGX Spark). 0 disables the fallback.
	PreConfiguredDeviceMemoryMB int64        `json:"preConfiguredDeviceMemoryMB"`
	// NodeConfig allows per-node overrides that take priority over the global default.
	NodeConfig                  []NodeConfig `json:"nodeConfig"`
}

// NodeConfig overrides DevicePluginConfig fields for a specific node.
type NodeConfig struct {
	// Name is the Kubernetes node name this config applies to.
	Name                        string `json:"name"`
	// PreConfiguredDeviceMemoryMB overrides the global value for this node.
	// nil means inherit the global default.
	PreConfiguredDeviceMemoryMB *int64 `json:"preConfiguredDeviceMemoryMB"`
}

func loadDevicePluginConfig(path string) (*DevicePluginConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg DevicePluginConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// resolvePreConfiguredDeviceMemoryBytes returns the configured fallback GPU
// memory in bytes for nodeName. Per-node config takes priority over the global
// default. Returns 0 when no override is configured.
func resolvePreConfiguredDeviceMemoryBytes(cfg *DevicePluginConfig, nodeName string) uint64 {
	if cfg == nil {
		return 0
	}
	for _, nc := range cfg.NodeConfig {
		if nc.Name == nodeName && nc.PreConfiguredDeviceMemoryMB != nil {
			return uint64(*nc.PreConfiguredDeviceMemoryMB) * 1024 * 1024
		}
	}
	return uint64(cfg.PreConfiguredDeviceMemoryMB) * 1024 * 1024
}
