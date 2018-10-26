package handlerImpl

import (
	"github.com/ortuman/jackal/component/httpapi/handlerImpl/domainHandle"
	"github.com/ortuman/jackal/component/httpapi/handlerImpl/pubsubHandle"
	"github.com/ortuman/jackal/component/httpapi/handlerImpl/userHandle"
	"github.com/ortuman/jackal/component/httpapi/handlerImpl/xmlhandle"
)

//接口实例
type Handlerface struct {
	xmlhandle.XmlHandlerface
	domainHandle.DoaminHandlerface
	userHandle.UserHandlerface
	pubsubHandle.PubSubHandlerface
}

//初始化接口实例
func New() *Handlerface {
	s := &Handlerface{
		xmlhandle.XmlHandlerface{},
		domainHandle.DoaminHandlerface{},
		userHandle.UserHandlerface{},
		pubsubHandle.PubSubHandlerface{},
	}
	return s
}
