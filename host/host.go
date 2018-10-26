/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package host

import (
	"crypto/tls"
	"fmt"
	"github.com/ortuman/jackal/component/httpapi/storage"
	"log"
	"sync"

	"github.com/ortuman/jackal/util"
)

const defaultDomain = "localhost"
const defaultDomain2 = "localhost2"
const defaultDomain3 = "localhost3"
const defaultDomain4 = "localhost4"

var (
	instMu      sync.RWMutex
	hosts       = make(map[string]tls.Certificate)
	initialized bool
)

// Initialize initializes host manager sub system.
func Initialize(configurations []Config) {
	instMu.Lock()
	defer instMu.Unlock()
	if initialized {
		return
	}
	//start - add - lxf - 20181026
	//如果服务器重启，加载数据库中的虚拟域
	doms, err := storage.Instance().FetchDomainAll()
	if err != nil {
		log.Fatalf("%v", fmt.Errorf("Initialize domain in store error: %s", err))
		return
	}
	if len(*doms) > 0 {
		for _, dom := range *doms {
			LoadCertificateForVirtualDomain(dom.ServiceJid)
		}
	}
	//end - add - lxf - 20181026

	if len(configurations) > 0 {
		for _, h := range configurations {
			hosts[h.Name] = h.Certificate
		}
	} else {
		cer, err := util.LoadCertificate("", "", defaultDomain)
		if err != nil {
			log.Fatalf("%v", err)
		}
		hosts[defaultDomain] = cer
		//start - update - lxf - 20181024
		LoadCertificateForVirtualDomain(defaultDomain2)
		LoadCertificateForVirtualDomain(defaultDomain3)
		LoadCertificateForVirtualDomain(defaultDomain4)
		//end - update - lxf - 20181024
	}
	initialized = true
}

//start - update - lxf - 20181024
//为新增的虚拟域创建证书，方法暴露给HTTP-API组件使用
func LoadCertificateForVirtualDomain(domainName string) (tls.Certificate, error) {
	cer, err := util.LoadCertificateForActiveDomain("", "", domainName)
	if err != nil {
		log.Fatalf("%v", err)
	}
	hosts[domainName] = cer
	return cer, err
}

//end - update - lxf - 20181024

// Shutdown shuts down host sub system.
func Shutdown() {
	instMu.Lock()
	defer instMu.Unlock()
	if initialized {
		hosts = make(map[string]tls.Certificate)
		initialized = false
	}
}

// HostNames returns current registered domain names.
func HostNames() []string {
	instMu.RLock()
	defer instMu.RUnlock()
	var ret []string
	for n, _ := range hosts {
		ret = append(ret, n)
	}
	return ret
}

// IsLocalHost returns true if domain is a local server domain.
func IsLocalHost(domain string) bool {
	instMu.RLock()
	defer instMu.RUnlock()
	_, ok := hosts[domain]
	return ok
}

// Certificates returns an array of all configured domain certificates.
func Certificates() []tls.Certificate {
	instMu.RLock()
	defer instMu.RUnlock()
	var certs []tls.Certificate
	for _, cer := range hosts {
		certs = append(certs, cer)
	}
	return certs
}

//start - add - lxf - 20181025
// 返回hosts
func Hosts() map[string]tls.Certificate {
	instMu.RLock()
	defer instMu.RUnlock()
	return hosts
}

//end - add - lxf - 20181025
