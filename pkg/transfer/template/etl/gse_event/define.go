// Tencent is pleased to support the open source community by making
// 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

package gse_event

import (
	"encoding/json"
	"fmt"
	"github.com/cstockton/go-conv"
)

type EventRecord struct {
	EventName      string                 `json:"event_name"`
	Event          map[string]interface{} `json:"event"`
	EventDimension map[string]interface{} `json:"dimension"`
	Target         string                 `json:"target"`
	Timestamp      *float64               `json:"timestamp"`
}

// SystemEventData : 自定义字符串事件
type SystemEventData struct {
	Time   string `json:"utctime2"`
	Values []struct {
		EventTime string          `json:"event_time"`
		Extra     json.RawMessage `json:"extra"`
	} `json:"value"`
}

type EventTypeData struct {
	Type int `json:"type"`
}

type BaseEvent interface {
	Flat() []EventRecord
}

// AgentLostEvent : agent失联事件
type AgentLostEvent struct {
	Hosts []struct {
		IP      string `json:"ip"`
		CloudID int    `json:"cloudid"`
		AgentID string `json:"agent_id"`
	} `json:"host"`
}

func (e *AgentLostEvent) Flat() []EventRecord {
	var records []EventRecord
	var target string
	for _, host := range e.Hosts {

		if host.AgentID == "" {
			target = host.AgentID
		} else {
			target = fmt.Sprintf("%d:%s", host.CloudID, host.IP)
		}

		records = append(records, EventRecord{
			EventName: "agent_lost",
			Target:    target,
			Event: map[string]interface{}{
				"content": "agent_lost",
			},
			EventDimension: map[string]interface{}{
				"bk_target_cloud_id": conv.String(host.CloudID),
				"bk_target_ip":       host.IP,
				"ip":                 host.IP,
				"bk_cloud_id":        conv.String(host.CloudID),
				"bk_agent_id":        host.AgentID,
			},
		})
	}
	return records
}

// CorefileEvent : corefile事件
type CorefileEvent struct {
	Host           string `json:"host"`
	CloudID        int    `json:"cloudid"`
	Executable     string `json:"executable"`
	ExecutablePath string `json:"executable_path"`
	Signal         string `json:"signal"`
	Corefile       string `json:"corefile"`
	Filesize       string `json:"filesize"`
}

func (e *CorefileEvent) Flat() []EventRecord {
	return []EventRecord{
		{
			EventName: "corefile",
			Target:    fmt.Sprintf("%d:%s", e.CloudID, e.Host),
			Event: map[string]interface{}{
				"content": "corefile",
			},
			EventDimension: map[string]interface{}{
				"bk_target_cloud_id": conv.String(e.CloudID),
				"bk_target_ip":       e.Host,
				"ip":                 e.Host,
				"bk_cloud_id":        conv.String(e.CloudID),
				"executable":         e.Executable,
				"executable_path":    e.ExecutablePath,
				"signal":             e.Signal,
				"corefile":           e.Corefile,
				"filesize":           e.Filesize,
			},
		},
	}
}

// DiskFullEvent : 磁盘满事件
type DiskFullEvent struct {
	Host       string `json:"host"`
	CloudID    int    `json:"cloudid"`
	Disk       string `json:"disk"`
	FileSystem string `json:"file_system"`
	FsType     string `json:"fstype"`
}

func (e *DiskFullEvent) Flat() []EventRecord {
	return []EventRecord{
		{
			EventName: "disk_full",
			Target:    fmt.Sprintf("%d:%s", e.CloudID, e.Host),
			Event: map[string]interface{}{
				"content": "disk_full",
			},
			EventDimension: map[string]interface{}{
				"bk_target_cloud_id": conv.String(e.CloudID),
				"bk_target_ip":       e.Host,
				"ip":                 e.Host,
				"bk_cloud_id":        conv.String(e.CloudID),
				"disk":               e.Disk,
				"file_system":        e.FileSystem,
				"fstype":             e.FsType,
			},
		},
	}
}

// DiskReadonlyEvent : 磁盘只读事件
type DiskReadonlyEvent struct {
	Host    string `json:"host"`
	CloudID int    `json:"cloudid"`
	Ro      []struct {
		Position string `json:"position"`
		Fs       string `json:"fs"`
		Type     string `json:"type"`
	} `json:"ro"`
}

func (e *DiskReadonlyEvent) Flat() []EventRecord {
	events := make([]EventRecord, 0)
	for _, ro := range e.Ro {
		events = append(events, EventRecord{
			EventName: "disk_readonly",
			Target:    fmt.Sprintf("%d:%s", e.CloudID, e.Host),
			Event: map[string]interface{}{
				"content": "disk_readonly",
			},
			EventDimension: map[string]interface{}{
				"bk_target_cloud_id": conv.String(e.CloudID),
				"bk_target_ip":       e.Host,
				"ip":                 e.Host,
				"bk_cloud_id":        conv.String(e.CloudID),
				"position":           ro.Position,
				"fs":                 ro.Fs,
				"type":               ro.Type,
			},
		})
	}
	return events
}

// OOMEvent : OOM事件
type OOMEvent struct {
	Host       string `json:"host"`
	CloudID    int    `json:"cloudid"`
	Process    string `json:"process"`
	Message    string `json:"message"`
	OOMMemcg   string `json:"oom_memcg"`
	TaskMemcg  string `json:"task_memcg"`
	Task       string `json:"task"`
	Constraint string `json:"constraint"`
}

func (e *OOMEvent) Flat() []EventRecord {
	return []EventRecord{
		{
			EventName: "oom",
			Target:    fmt.Sprintf("%d:%s", e.CloudID, e.Host),
			Event: map[string]interface{}{
				"content": "oom",
			},
			EventDimension: map[string]interface{}{
				"bk_target_cloud_id": conv.String(e.CloudID),
				"bk_target_ip":       e.Host,
				"ip":                 e.Host,
				"bk_cloud_id":        conv.String(e.CloudID),
				"process":            e.Process,
				"message":            e.Message,
				"oom_memcg":          e.OOMMemcg,
				"task_memcg":         e.TaskMemcg,
				"task":               e.Task,
				"constraint":         e.Constraint,
			},
		},
	}
}

// PingLostEvent : ping不可达事件
type PingLostEvent struct {
	Hosts   []string `json:"iplist"`
	CloudID int      `json:"cloudid"`
}

func (e *PingLostEvent) Flat() []EventRecord {
	events := make([]EventRecord, 0)
	for _, host := range e.Hosts {
		events = append(events, EventRecord{
			EventName: "ping_lost",
			Target:    fmt.Sprintf("%d:%s", e.CloudID, host),
			Event: map[string]interface{}{
				"content": "ping_lost",
			},
			EventDimension: map[string]interface{}{
				"bk_target_cloud_id": conv.String(e.CloudID),
				"bk_target_ip":       host,
				"ip":                 host,
				"bk_cloud_id":        conv.String(e.CloudID),
			},
		})
	}
	return events
}

func parseSystemEvent(data json.RawMessage) []EventRecord {
	var event BaseEvent
	eventType := new(EventTypeData)
	err := json.Unmarshal(data, eventType)
	if err != nil {
		return nil
	}

	// 根据事件类型转换为不同的事件
	switch eventType.Type {
	case 2:
		// agent失联事件
		agentLostEvent := new(AgentLostEvent)
		err = json.Unmarshal(data, agentLostEvent)
		if err != nil {
			break
		}
		event = agentLostEvent
	case 3:
		// disk readonly
		diskReadonlyEvent := new(DiskReadonlyEvent)
		err = json.Unmarshal(data, diskReadonlyEvent)
		if err != nil {
			break
		}
		event = diskReadonlyEvent
	case 6:
		// disk full
		diskFullEvent := new(DiskFullEvent)
		err = json.Unmarshal(data, diskFullEvent)
		if err != nil {
			break
		}
		event = diskFullEvent
	case 7:
		// corefile
		corefileEvent := new(CorefileEvent)
		err = json.Unmarshal(data, corefileEvent)
		if err != nil {
			break
		}
		event = corefileEvent
	case 8:
		// ping
		pingEvent := new(PingLostEvent)
		err = json.Unmarshal(data, pingEvent)
		if err != nil {
			break
		}
		event = pingEvent
	case 9:
		// oom
		oomEvent := new(OOMEvent)
		err = json.Unmarshal(data, oomEvent)
		if err != nil {
			break
		}
		event = oomEvent
	}

	if event == nil {
		return nil
	}

	// 将数据转换为标准事件
	return event.Flat()
}