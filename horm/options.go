// Copyright (c) 2024 The horm-database Authors. All rights reserved.
// This file Author:  CaoHao <18500482693@163.com> .
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package horm

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/horm-database/common/errs"
	"github.com/horm-database/common/log/logger"
	"github.com/horm-database/common/snowflake"
	"github.com/horm-database/common/util"

	"gopkg.in/yaml.v3"
)

// Options are client options.
type Options struct {
	WorkspaceID int    // workspace id
	Encryption  int8   // frame encryption
	Token       string // workspace token
	Timeout     uint32 // timeout Millisecond
	Name        string // call name it better to be workspace_name.app.server.service
	Caller      string // get from name
	Appid       uint64 // appid
	Secret      string // secret
	Target      string // server target address
	LocalIP     string // 本地 ip
	Location    struct {
		Region string // 区域
		Zone   string // 城市
		Compus string // 园区
	}
}

var options = make(map[string]*Options)

func getOptions(name string) *Options {
	opts, ok := options[name]
	if !ok {
		return &Options{
			LocalIP: util.GetLocalIP(),
			Timeout: defaultTimeout,
		}
	}

	return opts
}

// Option sets client options.
type Option func(*Options)

// WithTimeout returns an Option that sets timeout of server.
func WithTimeout(timeout uint32) Option {
	return func(o *Options) {
		o.Timeout = timeout
	}
}

// WithName returns an Option that sets call name of client.
func WithName(name string) Option {
	return func(o *Options) {
		o.Name = name
	}
}

// WithAppID returns an Option that sets appid.
func WithAppID(appid uint64) Option {
	return func(o *Options) {
		o.Appid = appid
	}
}

// WithSecret returns an Option that sets appid`s secret.
func WithSecret(secret string) Option {
	return func(o *Options) {
		o.Secret = secret
	}
}

// WithTarget returns an Option that sets target of server.
func WithTarget(target string) Option {
	return func(o *Options) {
		o.Target = target
	}
}

// WithLocalIP returns an Option that sets ip of client.
func WithLocalIP(ip string) Option {
	return func(o *Options) {
		o.LocalIP = ip
	}
}

// WithWorkspaceID returns an Option that sets workspace of server.
func WithWorkspaceID(workspaceID int) Option {
	return func(o *Options) {
		o.WorkspaceID = workspaceID
	}
}

// WithEncryption returns an Option that sets encryption of frame.
func WithEncryption(encryption int8) Option {
	return func(o *Options) {
		o.Encryption = encryption
	}
}

// WithToken returns an Option that sets token of workspace of server.
func WithToken(token string) Option {
	return func(o *Options) {
		o.Token = token
	}
}

// WithLocation returns an Option that sets location of client.
func WithLocation(region, zone, compus string) Option {
	return func(o *Options) {
		o.Location.Region = region
		o.Location.Zone = zone
		o.Location.Compus = compus
	}
}

const (
	confFile       = "./orm.yaml"
	defaultTimeout = 60000 // 单位 ms
)

func init() {
	LoadConfig()
}

// config 配置
type config struct {
	Machine   string `yaml:"machine"`    // 容器名称
	MachineID int    `yaml:"machine_id"` // 容器编号（主要用于 snowflake 生成全局唯一 id）
	LocalIP   string `yaml:"local_ip"`   // 本地IP，容器内为容器ip，物理机或虚拟机为本机ip
	Location  struct {
		Region string `yaml:"region"` // 区域
		Zone   string `yaml:"zone"`   // 城市
		Compus string `yaml:"compus"` // 园区
	} `yaml:"location"` // 接入端所属区域，主要用于就近路由

	Server []*serverConfig  `yaml:"server"`
	DB     []*dbConfig      `yaml:"db"`
	Log    []*logger.Config `yaml:"log"`
}

type serverConfig struct {
	WorkspaceID int             `yaml:"workspace_id"` // workspace
	Encryption  int8            `yaml:"encryption"`   // 帧签名方式 0-无（默认） 1-签名 2-加密
	Token       string          `yaml:"token"`        // token
	Target      string          `yaml:"target"`       // workspace 地址
	Timeout     uint32          `yaml:"timeout"`      // 接口调用超时时间（毫秒）
	Caller      []*callerConfig `yaml:"caller"`       // 调用方信息
}

type callerConfig struct {
	Name    string `yaml:"name"`    // 调用名（必须全局唯一）组成最好是 workspace_name.caller_app.caller_server.caller_service
	AppID   uint64 `yaml:"appid"`   // 调用方 appid
	Secret  string `yaml:"secret"`  // 调用方秘钥
	Timeout uint32 `yaml:"timeout"` // 接口调用超时时间（毫秒）
}

// DBConfig 数据库配置
type dbConfig struct {
	Name         string `yaml:"name"`          // 数据库名称
	Type         string `yaml:"type"`          // 数据库类型 elastic redis mysql postgresql clickhouse
	Version      string `yaml:"version"`       // 数据库版本，elastic v6，v7 版本调用接口不一样
	Network      string `yaml:"network"`       // network TCP/UDP，默认 TCP
	Address      string `yaml:"address"`       // 地址
	WriteTimeout int    `yaml:"write_timeout"` // 写超时（毫秒）
	ReadTimeout  int    `yaml:"read_timeout"`  // 读超时（毫秒）
	WarnTimeout  int    `yaml:"warn_timeout"`  // 告警超时（ms），如果请求耗时超过这个时间，就会打 warning 日志
	OmitError    int8   `yaml:"omit_error"`    // 是否忽略 error 日志，0-否 1-是
	Debug        int8   `yaml:"debug"`         // 是否开启 debug 日志，正常的数据库请求也会被打印到日志，0-否 1-是，会造成海量日志，慎重开启
}

var dbConfigs = make(map[string]*dbConfig)

// GetDBConfig 获取数据库配置
func GetDBConfig(name string) (*dbConfig, error) {
	dbCfg, ok := dbConfigs[name]
	if !ok || dbCfg == nil {
		return nil, errs.Newf(errs.ErrDBConfigNotFound, "not find db config: %s", name)
	}

	return dbCfg, nil
}

// LoadConfig 加载配置文件
func LoadConfig(confPath ...string) {
	fileName := confFile

	if len(confPath) > 0 {
		fileName = confPath[0]
	}

	cfg, err := parseConfFile(fileName)
	if err != nil {
		panic(fmt.Errorf("load horm db config error: %v", err))
	}

	if cfg.MachineID > 0 {
		snowflake.SetMachineID(cfg.MachineID)
	}

	for _, v := range cfg.DB {
		dbConfigs[v.Name] = v
	}

	for _, server := range cfg.Server {
		for _, caller := range server.Caller {
			opts := Options{
				WorkspaceID: server.WorkspaceID,
				Token:       server.Token,
				Encryption:  server.Encryption,
				Timeout:     caller.Timeout,
				Name:        caller.Name,
				Caller:      caller.Name,
				Appid:       caller.AppID,
				Secret:      caller.Secret,
				Target:      server.Target,
				LocalIP:     cfg.LocalIP,
			}

			i := strings.Index(caller.Name, ".")
			if i != -1 {
				opts.Caller = caller.Name[i+1:]
			}

			opts.Location.Region = cfg.Location.Region
			opts.Location.Zone = cfg.Location.Zone
			opts.Location.Compus = cfg.Location.Compus

			if opts.Timeout == 0 {
				opts.Timeout = server.Timeout
				if opts.Timeout == 0 {
					opts.Timeout = defaultTimeout
				}
			}

			options[caller.Name] = &opts
		}
	}

	return
}

func parseConfFile(confFile string) (*config, error) {
	buf, err := ioutil.ReadFile(confFile)
	if err != nil {
		return nil, err
	}

	cfg := config{}
	err = yaml.Unmarshal(buf, &cfg)
	return &cfg, err
}
