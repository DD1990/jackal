/*
 *接口抽象
 */
package handler

import (
	"github.com/ortuman/jackal/component/httpapi/handlerImpl"
	"log"
	"net/http"
	"sync"
)

var (
	instMu sync.RWMutex
	inst   Handlerface

	//	initialized bool
)

//HTTP请求处理接口
type userHandlerface interface {
	//请求URI：/rest/user/
	UserManage(w http.ResponseWriter, r *http.Request)
}

type XmlHandlerface interface {

	//请求URI：/rest/stream/
	StreamHandle(w http.ResponseWriter, r *http.Request)
}

type DoaminHandlerface interface {

	//请求URI：/rest/adhoc/
	DomainManage(w http.ResponseWriter, r *http.Request)
}

type PubSubHandlerface interface {

	//请求URI：/rest/pubsub/
	PubSubManage(w http.ResponseWriter, r *http.Request)
}

type Handlerface interface {
	//用户处理接口
	userHandlerface

	//xml处理接口实例
	XmlHandlerface

	//多域处理接口实例
	DoaminHandlerface

	//节点处理接口实例
	PubSubHandlerface
}

//初始化接口实例
func Intli() {
	inst = handlerImpl.New()
}

//接口锁，锁定当前接口，接口处理完成后解锁
func Instance() Handlerface {
	instMu.RLock()
	defer instMu.RUnlock()
	if inst == nil {
		log.Fatalf("Handlerface initialized")
	}
	return inst
}
