package model

import "time"

//接口请求方法
const (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	DELETE = "DELETE"
)

//接口请求地址
const (
	RESTUSER   = "/rest/user/"   //用户管理
	RESTUSERS  = "/rest/users/"  //批量用户管理
	RESTAVATAR = "/rest/avatar/" //头像设置
	RESTPUBSUB = "/rest/pubsub/" //节点管理、消息订阅及发布
	RESTADHOC  = "/rest/adhoc/"  //虚拟域管理
	RESTSTATS  = "/rest/stats/"  //统计管理
	RESTSTREAM = "/rest/stream/" //单一消息发送
)

//接口请求方法
const (
	CREATENODE    = "create-node"           //创建节点
	SUBSCRIBENODE = "subscribe-node"        //用户订阅
	PUBLISHITEM   = "publish-item"          //消息群发
	DELETENODE    = "delete-node"           //删除节点
	ITEMUPDAT     = "comp-repo-item-update" //消息更新
	ITEMADD       = "comp-repo-item-add"    //消息发布
	ITEMREMOVE    = "comp-repo-item-remove" //消息删除
)

//虚拟域实例
type Domain struct {
	ServiceId     uint64 //虚拟域Id
	ServiceJid    string //虚拟域Jid
	ServiceJidSha string //虚拟域Jid加密
}

//节点实例
type Node struct {
	NodeId        uint64    //节点Id
	ServiceId     uint64    //虚拟域Id
	NodeName      string    //节点名称
	NodeNameSha   string    //节点名称加密
	Type          int8      //节点类型
	Title         string    //标题
	Description   string    //描述
	CreatorId     uint64    //创建者Id
	CreationDate  time.Time //创建时间
	Configuration string    //配置
	CollectionId  uint64    //（TODO(lxf) 作用未知）
}

//JID实例
type Jid struct {
	JidId  uint64 //JIDId
	Jid    string //Jid
	JidSha string //Jid加密
}

//用户订阅实例
type Subscribe struct {
	NodeId         uint64 //节点Id
	JidId          uint64 //JIDId
	Subscription   string // 订阅描述
	SubscriptionId string // 订阅id（TODO(lxf) 作用未知）
}

//消息群发实例
type Item struct {
	NodeId       uint64    //节点Id
	Id           string    //Item Id
	IdSha1       string    //Item Id 加密
	CreationDate time.Time //创建时间
	PublisherId  uint64    //发布者Id
	UpdateDate   time.Time //更新时间
	Data         string    //群发内容
}
