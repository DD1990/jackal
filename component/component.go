/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package component

import (
	"fmt"
	"github.com/ortuman/jackal/component/httpapi"
	"sync"

	"github.com/ortuman/jackal/stream"

	apiStorage "github.com/ortuman/jackal/component/httpapi/storage"
	"github.com/ortuman/jackal/component/httpupload"
	"github.com/ortuman/jackal/component/muc"
	mucStorage "github.com/ortuman/jackal/component/muc/storage"
	"github.com/ortuman/jackal/log"
	"github.com/ortuman/jackal/module"
	"github.com/ortuman/jackal/xmpp"
)

type Component interface {
	Host() string
	ProcessStanza(stanza xmpp.Stanza, stm stream.C2S)
}

// singleton interface
var (
	instMu      sync.RWMutex
	comps       map[string]Component
	shutdownCh  chan struct{}
	initialized bool
)

// Initialize initializes the components manager.
func Initialize(cfg *Config) {
	instMu.Lock()
	defer instMu.Unlock()
	if initialized {
		return
	}
	shutdownCh = make(chan struct{})
	// 根据配置文件中的组件参数，加载组件
	cs := loadComponents(cfg)

	comps = make(map[string]Component)
	for _, c := range cs {
		// 获取组件端口
		host := c.Host()
		if _, ok := comps[host]; ok {
			log.Fatalf("%v", fmt.Errorf("component host name conflict: %s", host))
		}
		//将组件放入内存中，以端口为key
		comps[host] = c
	}

	//start - add - lxf - 20181026
	//如果服务器重启，加载数据库中的虚拟域
	discoInfo := module.Modules().DiscoInfo
	doms, err := apiStorage.Instance().FetchDomainAll()
	if err != nil {
		log.Fatalf("%v", fmt.Errorf("Initialize domain in store error: %s", err))
		return
	}
	if len(*doms) > 0 {
		for _, dom := range *doms {
			c := muc.New(muc.Config{"muc." + dom.ServiceJid}, discoInfo, shutdownCh)
			// 获取组件端口
			host := c.Host()
			if _, ok := comps[host]; ok {
				log.Fatalf("%v", fmt.Errorf("component host name conflict: %s", host))
			}
			//将组件放入内存中，以端口为key
			comps[host] = c
		}
	}
	//end - add - lxf - 20181026

	initialized = true
}

// Shutdown shuts down components manager system.
// This method should be used only for testing purposes.
func Shutdown() {
	instMu.Lock()
	defer instMu.Unlock()
	if !initialized {
		return
	}
	close(shutdownCh)
	comps = nil
	initialized = false
}

func Get(host string) Component {
	instMu.Lock()
	defer instMu.Unlock()
	if !initialized {
		return nil
	}
	return comps[host]
}

func GetAll() []Component {
	instMu.Lock()
	defer instMu.Unlock()
	if !initialized {
		return nil
	}
	var ret []Component
	for _, comp := range comps {
		ret = append(ret, comp)
	}
	return ret
}

func loadComponents(cfg *Config) []Component {
	discoInfo := module.Modules().DiscoInfo

	var ret []Component
	// 加载HTTP-API组件
	if cfg.HttpApi != nil {
		ret = append(ret, httpapi.New(cfg.HttpApi, shutdownCh))
		// muc持久化配置加载
		if cfg.Storage != nil {
			apiStorage.Initialize(cfg.ApiStorage)
		}
	}

	// 加载文件上传组件
	if cfg.HttpUpload != nil {
		ret = append(ret, httpupload.New(cfg.HttpUpload, discoInfo, shutdownCh))
	}
	// 加载muc服务组件
	if cfg.Mucs != nil {
		for _, mucret := range cfg.Mucs {
			ret = append(ret, muc.New(mucret, discoInfo, shutdownCh))
		}
	}
	// muc持久化配置加载
	if cfg.Storage != nil {
		mucStorage.Initialize(cfg.Storage)
	}
	return ret
}
