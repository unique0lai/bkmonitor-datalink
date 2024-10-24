// Tencent is pleased to support the open source community by making
// 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

package target

import (
	"fmt"
	"hash/fnv"
	"math"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/elastic/beats/libbeat/common/transport/tlscommon"
	"github.com/prometheus/prometheus/model/labels"
	"gopkg.in/yaml.v2"

	"github.com/TencentBlueKing/bkmonitor-datalink/pkg/operator/common/define"
)

const (
	relabelRuleWorkload = "v1/workload"
	relabelRuleNode     = "v1/node"
)

func IsBuiltinLabels(k string) bool {
	for _, label := range ConfBuiltinLabels {
		if k == label {
			return true
		}
	}
	return false
}

func toMonitorIndex(s string) int {
	if s == "" {
		return -1
	}
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return -1
	}
	return int(i)
}

// MetricTarget 指标采集配置
type MetricTarget struct {
	Meta                   define.MonitorMeta
	RelabelRule            string
	RelabelIndex           string
	NormalizeMetricName    bool
	Address                string
	NodeName               string
	Scheme                 string
	DataID                 int
	Namespace              string
	MaxTimeout             string
	MinPeriod              string
	Period                 string
	Timeout                string
	Path                   string
	ProxyURL               string
	Username               string
	Password               string
	TLSConfig              *tlscommon.Config
	BearerTokenFile        string
	BearerToken            string
	Labels                 labels.Labels     // 自动发现的静态label
	ExtraLabels            map[string]string // 添加的额外label
	DimensionReplace       map[string]string
	MetricReplace          map[string]string
	Params                 url.Values
	MetricRelabelConfigs   []yaml.MapSlice
	Mask                   string
	TaskType               string
	DisableCustomTimestamp bool
}

func (t *MetricTarget) FileName() string {
	s := fmt.Sprintf("%s-%s-%s-%d-%s", t.NodeName, t.Address, t.Path, t.Hash(), t.Mask)
	regx := regexp.MustCompile("[^a-zA-Z0-9]")
	name := regx.ReplaceAllString(s, "-")
	name = strings.ReplaceAll(name, "--", "-")
	return name
}

// RemoteRelabelConfig 返回采集器 workload 工作负载信息
func (t *MetricTarget) RemoteRelabelConfig() *yaml.MapItem {
	switch t.RelabelRule {
	case relabelRuleWorkload:
		// index >= 0 表示 annotations 中指定了 index label
		if idx := toMonitorIndex(t.RelabelIndex); idx >= 0 && idx != t.Meta.Index {
			return nil
		}
		return &yaml.MapItem{
			Key:   "metric_relabel_remote",
			Value: fmt.Sprintf("http://%s:%d/workload/node/%s", ConfServiceName, ConfServicePort, t.NodeName),
		}
	}
	return nil
}

func (t *MetricTarget) Hash() uint64 {
	h := fnv.New64a()
	b, _ := t.YamlBytes()
	h.Write(b)
	return h.Sum64()
}

func (t *MetricTarget) YamlBytes() ([]byte, error) {
	cfg := make(yaml.MapSlice, 0)
	if t.Period == "" {
		t.Period = ConfDefaultPeriod
	}
	if t.Timeout == "" {
		t.Timeout = t.Period
	}

	cfg = append(cfg, yaml.MapItem{Key: "type", Value: "metricbeat"})
	cfg = append(cfg, yaml.MapItem{Key: "name", Value: t.Address + t.Path})
	cfg = append(cfg, yaml.MapItem{Key: "version", Value: "1"})
	cfg = append(cfg, yaml.MapItem{Key: "dataid", Value: t.DataID})
	cfg = append(cfg, yaml.MapItem{Key: "max_timeout", Value: ConfMaxTimeout})
	cfg = append(cfg, yaml.MapItem{Key: "min_period", Value: ConfMinPeriod})

	task := make(yaml.MapSlice, 0)
	task = append(task, yaml.MapItem{Key: "task_id", Value: t.generateTaskID()})
	task = append(task, yaml.MapItem{Key: "bk_biz_id", Value: 2})
	task = append(task, yaml.MapItem{Key: "period", Value: t.Period})
	task = append(task, yaml.MapItem{Key: "timeout", Value: t.Timeout})
	task = append(task, yaml.MapItem{Key: "custom_report", Value: true})

	module := make(yaml.MapSlice, 0)
	module = append(module, yaml.MapItem{Key: "module", Value: "prometheus"})
	module = append(module, yaml.MapItem{Key: "metricsets", Value: []string{"collector"}})
	module = append(module, yaml.MapItem{Key: "enabled", Value: true})
	module = append(module, yaml.MapItem{Key: "period", Value: t.Period})
	module = append(module, yaml.MapItem{Key: "proxy_url", Value: t.ProxyURL})
	module = append(module, yaml.MapItem{Key: "timeout", Value: t.Timeout})

	if remoteRelabel := t.RemoteRelabelConfig(); remoteRelabel != nil {
		module = append(module, *remoteRelabel)
	}
	module = append(module, yaml.MapItem{Key: "disable_custom_timestamp", Value: t.DisableCustomTimestamp})
	module = append(module, yaml.MapItem{Key: "normalize_metric_name", Value: t.NormalizeMetricName})

	address := t.Address
	if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
		address = fmt.Sprintf("%s://%s", t.Scheme, address)
	}
	module = append(module, yaml.MapItem{Key: "hosts", Value: []string{address}})
	if len(t.Params) != 0 {
		params := make(yaml.MapSlice, 0)
		keys := make([]string, 0)
		for key := range t.Params {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			params = append(params, yaml.MapItem{
				Key:   key,
				Value: t.Params[key],
			})
		}
		module = append(module, yaml.MapItem{Key: "query", Value: params})
	}
	module = append(module, yaml.MapItem{Key: "namespace", Value: t.Namespace})
	module = append(module, yaml.MapItem{Key: "metrics_path", Value: t.Path})

	if t.Username != "" && t.Password != "" {
		module = append(module, yaml.MapItem{Key: "username", Value: t.Username})
		module = append(module, yaml.MapItem{Key: "password", Value: t.Password})
	}
	if t.BearerTokenFile != "" {
		module = append(module, yaml.MapItem{Key: "bearer_file", Value: t.BearerTokenFile})
	}
	if t.BearerToken != "" {
		module = append(module, yaml.MapItem{Key: "bearer_token", Value: t.BearerToken})
	}

	if t.DimensionReplace != nil {
		module = append(module, yaml.MapItem{Key: "dimension_replace", Value: sortMap(t.DimensionReplace)})
	}
	if t.MetricReplace != nil {
		module = append(module, yaml.MapItem{Key: "metric_replace", Value: sortMap(t.MetricReplace)})
	}
	if len(t.MetricRelabelConfigs) != 0 {
		module = append(module, yaml.MapItem{Key: "metric_relabel_configs", Value: t.MetricRelabelConfigs})
	}

	if t.Scheme == "https" {
		ssl := make(yaml.MapSlice, 0)
		ssl = append(ssl, yaml.MapItem{Key: "verification_mode", Value: "none"})
		if t.TLSConfig != nil {
			ssl = append(ssl, yaml.MapItem{Key: "certificate_authorities", Value: t.TLSConfig.CAs})
			ssl = append(ssl, yaml.MapItem{Key: "certificate", Value: t.TLSConfig.Certificate.Certificate})
			ssl = append(ssl, yaml.MapItem{Key: "key", Value: t.TLSConfig.Certificate.Key})
		}
		module = append(module, yaml.MapItem{Key: "ssl", Value: ssl})
	}

	lbs := make(yaml.MapSlice, 0)
	for _, label := range t.Labels {
		// 跳过内置维度，这些维度均不上报
		if strings.HasPrefix(label.Name, "__") && strings.HasSuffix(label.Name, "__") {
			continue
		}
		// 如果有内置管理维度 则追加 label 并统一加上 bk_ 前缀
		if IsBuiltinLabels(label.Name) {
			lbs = append(lbs, yaml.MapItem{
				Key:   fmt.Sprintf("bk_%s", label.Name),
				Value: label.Value,
			})
		}
		lbs = append(lbs, yaml.MapItem{
			Key:   label.Name,
			Value: label.Value,
		})
	}
	lbs = append(lbs, yaml.MapItem{Key: "bk_endpoint_url", Value: address + t.Path})
	lbs = append(lbs, yaml.MapItem{Key: "bk_endpoint_index", Value: fmt.Sprintf("%d", t.Meta.Index)})
	lbs = append(lbs, yaml.MapItem{Key: "bk_monitor_name", Value: t.Meta.Name})
	lbs = append(lbs, yaml.MapItem{Key: "bk_monitor_namespace", Value: t.Meta.Namespace})

	if t.RelabelRule == relabelRuleNode {
		lbs = append(lbs, yaml.MapItem{Key: "node", Value: t.NodeName})
	}

	lbs = append(lbs, sortMap(t.ExtraLabels)...)
	task = append(task, yaml.MapItem{Key: "labels", Value: []yaml.MapSlice{lbs}})
	task = append(task, yaml.MapItem{Key: "module", Value: module})
	cfg = append(cfg, yaml.MapItem{Key: "tasks", Value: []yaml.MapSlice{task}})
	return yaml.Marshal(cfg)
}

func (t *MetricTarget) generateTaskID() uint64 {
	h := fnv.New64a()

	h.Write([]byte(fmt.Sprintf("%d", t.DataID)))
	h.Write([]byte(t.Address))
	h.Write([]byte(t.Path))
	for _, label := range t.Labels {
		h.Write([]byte(label.Name))
		h.Write([]byte(label.Value))
	}
	h.Write([]byte(fmt.Sprintf("%d", t.Meta.Index)))
	h.Write([]byte(t.Namespace))
	h.Write([]byte(t.Meta.Name))
	return avoidOverflow(h.Sum64())
}

func avoidOverflow(num uint64) uint64 {
	if num > math.MaxInt32 {
		return avoidOverflow(num / 50)
	}
	return num
}

func sortMap(origin map[string]string) []yaml.MapItem {
	result := make(yaml.MapSlice, 0, len(origin))
	keys := make([]string, 0)
	for key := range origin {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		result = append(result, yaml.MapItem{
			Key:   key,
			Value: origin[key],
		})
	}
	return result
}
