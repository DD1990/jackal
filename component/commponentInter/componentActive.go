package commponentInter

import (
	"fmt"
	"sync"

	"github.com/ortuman/jackal/stream"

	"github.com/ortuman/jackal/component/muc"
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
func Initialize(domainName string) {
	instMu.Lock()
	defer instMu.Unlock()
	if !initialized {
		shutdownCh = make(chan struct{})
		comps = make(map[string]Component)
	}
	// 根据配置文件中的组件参数，加载组件

	discoInfo := module.Modules().DiscoInfo
	c := muc.New(muc.Config{"muc." + domainName}, discoInfo, shutdownCh)
	// 获取组件端口
	host := c.Host()
	if _, ok := comps[host]; ok {
		log.Fatalf("%v", fmt.Errorf("component host name conflict: %s", host))
	}
	//将组件放入内存中，以端口为key
	comps[host] = c
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
