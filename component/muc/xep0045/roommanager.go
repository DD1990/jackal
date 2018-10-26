/*
 * 创建时间：2018 09 18.
 * 聊天室协议包.
 */

package xep0045

import (
	"errors"
	"github.com/ortuman/jackal/component/muc/model/roommodel"
	"github.com/ortuman/jackal/component/muc/storage"
	"github.com/ortuman/jackal/log"
	"github.com/ortuman/jackal/module/xep0004"
	"github.com/ortuman/jackal/router"
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
	"strconv"
	"sync"
	"time"
)

const mailboxSize = 2048
const defaultroom = "defaultroom"

//muc关键标签
const (
	chatRoomNamespace = "http://jabber.org/protocol/muc"
	mucOwner          = "http://jabber.org/protocol/muc#owner"
	mucAdmin          = "http://jabber.org/protocol/muc#admin"
	mucuser           = "http://jabber.org/protocol/muc#user"

	shaone = "sha-1"
	smack  = "http://www.igniterealtime.org/projects/smack"
	caps   = "http://jabber.org/protocol/caps"
	verpas = "TJuVIXqTCVfJSthaPu4MtTbaf9A="
)

//muc相关属性
const (
	passwordprotectedFeature = "muc_passwordprotected"
	hiddenFeature            = "muc_hidden"
	temporaryFeature         = "muc_temporary"
	openFeature              = "muc_open"
	unmoderatedFeature       = "muc_unmoderated"
	nonanonymousFeature      = "muc_nonanonymous"
)

// 功能实体。
type Feature = string

// 信息标识实体。
type Identity struct {
	Category string
	Type     string
	Name     string
}

// 信息项实体。
type Item struct {
	Jid  string
	Name string
	Node string
}

//定义房间成员结构
type roomuser struct {
	roomname     string
	roomusername string
}

// 通用的disco信息域提供者.
type Provider interface {
	// 返回与提供者关联的所有标识。
	Identities(toJID, fromJID *jid.JID, node string) []Identity

	// 返回与提供者关联的所有项,如果出现错误，应该返回正确的节错误。
	Items(toJID, fromJID *jid.JID, node string) ([]Item, *xmpp.StanzaError)

	// 返回与提供者关联的所有特性,如果出现错误，应该返回正确的节错误。
	Features(toJID, fromJID *jid.JID, node string) ([]Feature, *xmpp.StanzaError)
}

// 聊天室是最后一个活动流模块。
type ChatRoom struct {
	rmu        sync.RWMutex
	startTime  time.Time
	room       map[string][]roomuser
	actorCh    chan func()
	shutdownCh <-chan struct{}
}

// New返回最后一个活动IQ处理程序模块。
func New(shutdownCh <-chan struct{}) *ChatRoom {
	x := &ChatRoom{
		startTime:  time.Now(),
		room:       make(map[string][]roomuser),
		actorCh:    make(chan func(), mailboxSize),
		shutdownCh: shutdownCh,
	}
	go x.loop()
	//TODO(lxf) disco服务注册？还是 muc组件注册？
	//x.registerDiscoInfo(disco)
	return x
}

// MatchesIQ返回IQ是否应该由最后一个活动模块处理。
func (x *ChatRoom) MatchesIQ(iq *xmpp.IQ) bool {
	return iq.IsGet() && (iq.Elements().ChildNamespace("query", chatRoomNamespace) != nil || iq.Elements().ChildNamespace("query", mucOwner) != nil) //新增
}

// ProcessIQ处理的是IQ根据相关流的行为进行的最后一个活动。
func (x *ChatRoom) ProcessIQ(iq *xmpp.IQ, stm stream.C2S) {
	x.actorCh <- func() { x.processIQ(iq, stm) }
}

// 运行在自己的go程中
func (x *ChatRoom) loop() {
	for {
		select {
		case f := <-x.actorCh:
			f()
		case <-x.shutdownCh:
			return
		}
	}
}

// 处理iq类型的请求
func (x *ChatRoom) processIQ(iq *xmpp.IQ, stm stream.C2S) {
	toJID := iq.ToJID()
	if toJID.IsServer() {
		x.sendServerUptime(iq, stm)
	} else if toJID.IsBare() {
		q := iq.Elements().Child("query")
		if q == nil {
			return
		}
		if iq.IsGet() {
			switch q.Namespace() {
			case mucOwner:
				x.SendRoomConfig(iq, stm)
			case mucAdmin: //右键管理成员返回
				x.Membermanage(iq, stm)
			case "vcard-temp": //TODO(lxf) 发送即时消息，没处理之前会发送到所在群中
				/*
					<iq to='roomtest1@muc.test' id='CpbW5-265' type='get'><vCard xmlns='vcard-temp'/></iq>
					<iq type="error" id="CpbW5-265" xmlns="jabber:client" to="tig_test1@test/Spark" from="roomtest1@muc.test"><vCard xmlns="vcard-temp"/><error type="cancel" code="501"><feature-not-implemented xmlns="urn:ietf:params:xml:ns:xmpp-stanzas"/></error></iq>
				*/
			default:
				x.CreateChatRoomResult(iq, stm) //创建房间成功通知
			}
		} else if iq.IsSet() {
			switch q.Namespace() {
			case mucOwner:
				x.SendRoomConfigResult(iq, stm)
			case mucAdmin:
				x.MembermanageAll(iq, stm)
			}
		}
	} else {
		stm.SendElement(iq.BadRequestError())
	}
}

// 处理presence类型的请求
func (r *ChatRoom) ProcessPresence(presence *xmpp.Presence, stm stream.C2S) {
	r.actorCh <- func() {
		if err := r.processPresence(presence, stm); err != nil {
			log.Error(err)
		}
	}
}

// 处理message类型的请求
// TODO(lxf) message类型处理 移到groupchat.go中，单独启动go程处理
func (r *ChatRoom) ProcessMessage(message *xmpp.Message, stm stream.C2S) {
	r.actorCh <- func() {
		if err := r.processMessage(message); err != nil {
			log.Error(err)
		}
	}
}

func (r *ChatRoom) processMessage(message *xmpp.Message) error {
	log.Infof("进入 chat room processMessage（） 方法")
	//TODO(lxf)此处路由放在上层实现
	switch message.Attributes().Get("type") {
	case "groupchat":
	case "":
	default: //TODO(lxf)
	}

	//TODO(lxf) 判断房间是否存在(是否需要，假设不存在，直接发送失败也可以)
	/*if room is not exist{

	}esle{

	}*/

	//step1：TODO(lxf) 判断此客户是否被禁止发声， 如果是，则单独给此用户发送错误，xml示例如下
	//step1.1、查询用户在房间的角色是否为visitor（是则代表被设定为游客，禁止发声）
	rm, err := storage.Instance().FetchRoommemberItem(message.FromJID().Node(), message.ToJID().Node())
	if err != nil { // 遍历成员列表
		log.Error(err)
		return err
	}
	if rm.Role == none {
		//step1.2、单独发送错误反馈的xml给此用户
		err = sendNoPrivilegesErr(rm, message)
		return err
	}

	//step2:获取房间的用户成员
	//方式一：去数据库查询
	roomusersitms, err := storage.Instance().FetchRoommemberItems(message.ToJID().Node())
	if err != nil { // 遍历成员列表
		log.Error(err)
		return err
	}
	errm := sendGroupChatMessage(roomusersitms, message)
	if errm != nil {
		log.Error(errm)
		return errm
	}

	//TODO(lxf) 方式二：使用缓存数据库，如Redis（暂时不用此方法）
	//	roomusersitms, err := redis.Instance().FetchRoomUsersItems()

	//TODO(lxf) 方式三：使用服务器内存（暂时不用此方法）

	return nil
}

func (r *ChatRoom) processPresence(presence *xmpp.Presence, stm stream.C2S) error {
	log.Infof("进入 chat room 的processPresence（） 方法")
	//TODO 判断房间是否存在，如果不存在则创建r.CreateChatRoom(presence)，如果存在则进入房间r.SendChatRoom(presence)
	//step1:获取服务列表
	roomitms, err := storage.Instance().FetchRoomItems()
	if err == nil {
		for _, roomitem := range roomitms {
			if presence.ToJID().Node() == roomitem.Roomname {
				//step2:判断当前房间是否在服务列表中 - 在
				err := r.SendChatRoom(presence)
				return err
			}
		}
	}
	//step3:不在
	r.CreateChatRoom(presence)

	return nil
}

//创建房间或进入房间后，通知房间成员房间内成员列表信息
func (r *ChatRoom) SendChatRoom(presence *xmpp.Presence) error {
	log.Infof("into chat room... ")

	/* 房间不存在，无法加入房间，返回客户端错误
	<iq type="result" id="CGrF2-189" from="localhost" to="sofija@localhost/Spark" xml:lang="en">
		<query xmlns="jabber:iq:search">
		<instructions>Fill in one or more fields to search for any matching Jabber users.</instructions>
		<first/><last/><nick/><email/>
	</query></iq>
	*/
	//step1、去数据库查询，获取房间的用户成员
	roomusersitms, err := storage.Instance().FetchRoommemberItems(presence.ToJID().Node())
	if err != nil {
		//TODO(lxf) 发送加入房间失败
		log.Error(err)

		return err
	}

	userIsExist := false
	for _, user := range roomusersitms {
		//step2、告诉新成员，房间已经存在那些人，分别是什么角色
		pre := newPreOrIQElement("presence", presence.ID(), presence.ToJID().Node()+"@"+presence.ToJID().Domain()+"/"+user.Membername, presence.FromJID().String(), "", "jabber:client")

		c := xmpp.NewElementName("c")
		c.SetAttribute("hash", shaone)
		c.SetAttribute("node", smack)
		c.SetAttribute("xmlns", caps)
		c.SetAttribute("ver", verpas)

		x := xmpp.NewElementNamespace("x", mucuser)
		x.SetAttribute("type", "modify")
		item := xmpp.NewElementName("item")
		x.AppendElement(item)
		pre.AppendElement(c)
		pre.AppendElement(x)
		switch user.Role {
		case owner:
			if user.Membername == presence.FromJID().Node() {
				status1 := xmpp.NewElementName("status")
				status1.SetAttribute("code", "110")
				x.AppendElement(status1)
				status2 := xmpp.NewElementName("status")
				status2.SetAttribute("code", "170")
				x.AppendElement(status2)
			}
			item.SetAttribute("affiliation", "owner")
			item.SetAttribute("jid", presence.FromJID().String())
			item.SetAttribute("nick", presence.FromJID().Node())
			item.SetAttribute("role", "moderator")
		case admin:
			item.SetAttribute("affiliation", "admin")
			item.SetAttribute("role", "moderator")
		case member:
			item.SetAttribute("affiliation", "member")
			item.SetAttribute("role", "participant")
		case none:
			item.SetAttribute("affiliation", "none")
			item.SetAttribute("role", "visitor")
		case outcast: //不存在，因为排斥者不允许进入房间

		default:
			item.SetAttribute("affiliation", "none")
			item.SetAttribute("role", "visitor")
		}

		if router.UserStreams(presence.FromJID().Node()) == nil {
			return errors.New("用户信息不存在")
		} else {
			stm := router.UserStreams(presence.FromJID().Node())[0]
			stm.SendElement(pre)
		}

		if user.Membername == presence.FromJID().Node() {
			userIsExist = true
		} else {

			//step3、告诉房间已经存在那些人，新进成员信息及角色
			pre = newPreOrIQElement("presence", "", presence.ToJID().String(), user.Jid, "", "")
			x = xmpp.NewElementNamespace("x", mucuser)
			item = xmpp.NewElementName("item")
			x.SetAttribute("type", "modify")
			item.SetAttribute("affiliation", "member")
			item.SetAttribute("role", "participant")
			x.AppendElement(item)
			pre.AppendElement(x)

			//判断如果成员在线，则发送
			if router.UserStreams(user.Membername) != nil {
				stm := router.UserStreams(user.Membername)[0]
				stm.SendElement(pre)
			} else {
				//TODO(lxf) 如果成员不在线，则发送离线消息
			}
		}
	}

	if !userIsExist {
		//step4、发送新客户信息给自己，需增加<status code='110'/> 标识
		pre := newPreOrIQElement("presence", "", presence.ToJID().String(), presence.FromJID().String(), "", "")
		x := xmpp.NewElementNamespace("x", mucuser)
		item := xmpp.NewElementName("item")
		x.SetAttribute("type", "modify")
		item.SetAttribute("affiliation", "member")
		item.SetAttribute("role", "participant")
		status := xmpp.NewElementName("status")
		status.SetAttribute("code", "110")
		x.AppendElement(item)
		x.AppendElement(status)
		pre.AppendElement(x)

		stm := router.UserStreams(presence.FromJID().Node())[0]
		stm.SendElement(pre)

		//step5、将新成员加入到房间
		rmi := roommodel.RoommemberItem{presence.FromJID().Node(), presence.ToJID().Node(), presence.FromJID().String(), false, member}
		err = storage.Instance().InsertOrUpdateRoommemberItem(&rmi)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	return nil
}

//创建房间
func (r *ChatRoom) CreateChatRoom(presence *xmpp.Presence) error {

	rn := presence.ToJID().Node()   //要创建的房间名
	un := presence.FromJID().Node() //创建房间的用户名

	/*内存方式获取房间成员列表*/
	/*	r.rmu.Lock()
		defer r.rmu.Unlock()
		room := r.room[rn]
		log.Infof("start to create room...")
		if len(room) != 0 {//名为rn的房间已经存在于map中
			for _,ru := range room {
				if ru.roomusername == un {//TODO 如果是别的房间存在这个用户呢？（这种情况不存在，因为是拿房间名作为key值来找的）
					break
				}
			}
			room = append(room,roomuser{rn,un})
		}else{
			room = []roomuser{roomuser{rn,un}}
		}
		r.room[rn] = room//TODO 重要*/

	/**
	例子 144. 服务承认房间新建成功
	<presence from='darkcave@chat.shakespeare.lit/firstwitch' to='crone1@shakespeare.lit/desktop'>
	  <x xmlns='http://jabber.org/protocol/muc#user'>
		<item affiliation='owner' role='moderator'/>
		<status code='110'/>
		<status code='201'/>
	  </x>
	</presence>
	*/

	//step1、发送房间创建成功通知
	pre := newPreOrIQElement("presence", "", presence.ToJID().String(), presence.FromJID().String(), "", "")
	x := xmpp.NewElementNamespace("x", mucuser)
	x.SetAttribute("type", "modify")
	item := xmpp.NewElementName("item")
	item.SetAttribute("affiliation", "owner")
	item.SetAttribute("role", "moderator")
	status1 := xmpp.NewElementName("status")
	status1.SetAttribute("code", "110")
	status2 := xmpp.NewElementName("status")
	status2.SetAttribute("code", "201")
	x.AppendElement(item)
	x.AppendElement(status1)
	x.AppendElement(status2)
	pre.AppendElement(x)
	if router.UserStreams(un) != nil {
		stm := router.UserStreams(un)[0]
		stm.SendElement(pre)
	}

	//step2、将新房间持久化到数据库
	ri := roommodel.RoomItem{rn, presence.ToJID().Domain(), "", false, reservedroom}
	err := storage.Instance().InsertOrUpdateRoomItem(&ri)
	if err != nil {
		log.Infof("InsertOrUpdateRoomItem room... (%s)", err)
		return err
	}

	//step3、将创建此房间的用户持久化到房间成员表，设置为所有者角色
	rmi := roommodel.RoommemberItem{presence.FromJID().Node(), presence.ToJID().Node(), presence.FromJID().String(), false, owner}
	err = storage.Instance().InsertOrUpdateRoommemberItem(&rmi)
	if err != nil {
		log.Error(err)
		return err
	}
	//TODO(lxf) 发送<iq to="muc" id="d88eG-105" type="get"><query xmlns="http://jabber.org/protocol/disco#items"/></iq> 实时刷新客户端房间列表？
	//这样可能会导致没打开房间列表窗口的用户弹出房间列表窗口
	return nil
}

//右键快速管理成员功能
func (r *ChatRoom) Membermanage(iq *xmpp.IQ, stm stream.C2S) error {
	log.Infof("into Membermanage... ")
	/*
		<iq type="result" xmlns="jabber:client" id="On99z-148" to="tig_test1@test/Spark" from="room1@muc.test">
			<query xmlns="http://jabber.org/protocol/muc#admin"/>
		</iq>
	*/
	iqe := newPreOrIQElement("iq", iq.ID(), iq.ToJID().String(), iq.FromJID().String(), "result", "jabber:client")
	que := xmpp.NewElementNamespace("query", mucAdmin)
	iqe.AppendElement(que)
	stm.SendElement(iqe)
	return nil
}

//服务承认房间新建成功
func (r *ChatRoom) CreateChatRoomResult(iq *xmpp.IQ, stm stream.C2S) error {
	log.Infof("into CreateChatRoomResult... ")
	/**
	例子 144. 服务承认房间新建成功
	<iq from='darkcave@chat.shakespeare.lit' id='create2' to='crone1@shakespeare.lit/desktop' type='result'/>
	*/

	iqe := newPreOrIQElement("iq", iq.ID(), iq.ToJID().String(), iq.FromJID().String(), "result", "jabber:client")
	stm.SendElement(iqe)

	ri := roommodel.RoomItem{iq.ToJID().Node(), iq.ToJID().Domain(), "", false, 1}
	err := storage.Instance().InsertOrUpdateRoomItem(&ri)
	if err != nil {
		log.Infof("InsertOrUpdateRoomItem room... (%s)", err)
		return err
	}
	return nil
}

//发送配置请求给客户端
func (r *ChatRoom) SendRoomConfig(iq *xmpp.IQ, stm stream.C2S) error {
	log.Infof("into chatroom config... ")

	//方式一：
	//将房间设置的参数初始化到数据库，查询并进行xml解析后返回
	//TODO(lxf) 应该指存储field部分，即<x>标签内容，或者再加上<query>标签内容
	elem, err := storage.Instance().FetchRoomcfgItem(iq, iq.ToJID().Node())
	if err != nil && elem == nil {
		elem, err = storage.Instance().FetchRoomcfgItem(iq, defaultroom)
	}
	if err == nil && elem != nil {
		iqe := newPreOrIQElement("iq", iq.ID(), iq.ToJID().String(), iq.FromJID().String(), "result", "jabber:client")
		que := xmpp.NewElementName("query")
		que.SetAttribute("xmlns", mucOwner)
		que.AppendElement(*elem)
		iqe.AppendElement(que)
		stm.SendElement(iqe)
	}
	return nil
}

//发送配置成功给客户端
/*
<iq from='darkcave@chat.shakespeare.lit' id='create2' to='crone1@shakespeare.lit/desktop' type='result'/>
*/
func (r *ChatRoom) SendRoomConfigResult(iq *xmpp.IQ, stm stream.C2S) error {

	log.Infof("into chatroom config... ")
	/*
		思路：
			1、获取原有或缺省的配置信息（string格式）------------------已实现
			2、将配置转化为element格式-------------------------------已实现
			3、使用xep0004协议，将element转为field结构体--------------已实现
			4、将新设置的field替换原有配置field-----------------------已实现//用新配置的value全部覆盖原配置的value
			5、将处理后的field转为element----------------------------已实现
			6、将element转为string并持久化到数据库--------------------已实现
	*/

	//step1、解析iq，实现方法：思路（1、2、3、4、5）
	iqe, err := storage.Instance().FetchRoomcfgItem(iq, defaultroom)
	if err == nil {
		form, _ := xep0004.NewFormFromElement(*iqe)
		aa := iq.Elements().ChildNamespace("query", mucOwner).Elements().ChildNamespace("x", "jabber:x:data")
		formiq, _ := xep0004.NewFormFromElement(aa)
		fields := form.Fields
		fieldsiq := formiq.Fields

		/*formnew := &xep0004.DataForm{}
		formnew.Type = xep0004.Form
		elem := formnew.Element()*/

		elem := xmpp.NewElementNamespace("x", "jabber:x:data")
		elem.SetAttribute("type", "form")
		for _, field := range fields {
			for _, fieldiq := range fieldsiq {
				if field.Var == fieldiq.Var {
					//TODO(lxf) 如果原field没有value或者有多个，但是数量比新field多还是少不明确时的处理
					field.Values = fieldiq.Values //value的替换处理,option是选择项配置，不做处理
				}
			}
			elem.AppendElement(field.Element())
		}

		//step2、将iq的xml存放于roomcfg表，实现方法：思路（6）
		err = storage.Instance().InsertOrUpdateRoomcfgItem(elem, iq.ToJID().Node())

		//step3、发送<iq from='darkcave@chat.shakespeare.lit' id='create2' to='crone1@shakespeare.lit/desktop' type='result'/>通知客户端成功
		if err == nil {
			iqnew := newPreOrIQElement("iq", iq.ID(), iq.ToJID().String(), iq.FromJID().String(), "result")
			stm.SendElement(iqnew)
		}
	}
	return nil
}

//组织服务端处理客户端成员管理请求的响应
func newPresenceElement(iq *xmpp.IQ, optype string, passive bool) (*xmpp.Element, error) {

	nick := iq.Elements().Child("query").Elements().Child("item").Attributes().Get("nick")
	pre := newPreOrIQElement("presence", iq.ID(), iq.ToJID().String()+"/"+nick, iq.FromJID().String())
	x := xmpp.NewElementNamespace("x", mucuser)

	switch optype {
	case noneauth:
		item := newAuthItem(noneauth, nick, noneauth)
		res := xmpp.NewElementName("reason")
		res.SetText("你被踢出此聊天室")
		sta := xmpp.NewElementName("status")
		sta.SetAttribute("code", "307")
		item.AppendElement(res)
		x.AppendElement(item)
		x.AppendElement(sta)

		pre.SetAttribute("type", "unavailable")
		pre.AppendElement(x)

	case visitor, participant: //撤销声音、授予发言权
		if passive {
			pre = newPreOrIQElement("presence", iq.ID(), iq.ToJID().String()+"/"+nick, nick+"@"+iq.FromJID().Domain()+"/"+iq.FromJID().Resource())
		}
		item := newAuthItem(noneauth, nick, optype)
		c := xmpp.NewElementNamespace("c", caps)
		c.SetAttribute("hash", shaone)
		c.SetAttribute("node", smack)
		c.SetAttribute("ver", verpas)

		x.AppendElement(item)
		pre.AppendElement(c)
		pre.AppendElement(x)
	default:
	}
	return pre, nil
}
func SendNoneAuther(iq *xmpp.IQ) error {
	nick := iq.Elements().Child("query").Elements().Child("item").Attributes().Get("nick")
	pre := newPreOrIQElement("presence", iq.ID(), iq.ToJID().String()+"/"+nick, nick+"@"+iq.FromJID().Domain()+"/"+iq.FromJID().Resource(), "unavailable")
	x := xmpp.NewElementNamespace("x", mucuser)
	item := newAuthItem(noneauth, nick, noneauth)

	actor := xmpp.NewElementName("actor")
	actor.SetAttribute("jid", iq.FromJID().String())
	res := xmpp.NewElementName("reason")
	res.SetText("你被踢出此聊天室")
	sta := xmpp.NewElementName("status")
	sta.SetAttribute("code", "307")

	item.AppendElement(actor)
	item.AppendElement(res)
	x.AppendElement(item)
	x.AppendElement(sta)

	pre.AppendElement(x)
	if router.UserStreams(nick) == nil {
		return errors.New("用户信息不存在")
	} else {
		stm := router.UserStreams(nick)[0]
		stm.SendElement(pre)
		return nil
	}
}

//右键快速管理成员功能
func (r *ChatRoom) MembermanageAll(iq *xmpp.IQ, stm stream.C2S) error {
	log.Infof("into MembermanageAll... ")
	//TODO(lxf) nick后期改为被踢出者的node
	optype := iq.Elements().Child("query").Elements().Child("item").Attributes().Get("role")
	switch optype {
	case noneauth, visitor, participant: //踢出使用者、撤销声音、授予发言权
		elem, err := newPresenceElement(iq, optype, false)
		if err == nil {
			stm.SendElement(elem)

			iqnew := newPreOrIQElement("iq", iq.ID(), iq.ToJID().String(), iq.FromJID().String(), "result")
			stm.SendElement(iqnew)
			if optype == noneauth {
				//TODO(lxf) 踢出使用者,是否需要数据库变更成员状态为none，下次加入房间需要管理员审核？
				err = SendNoneAuther(iq)
				if err != nil {
					log.Error(err)
				}
			} else {
				//TODO(lxf) 被禁言或者被授予发言权，数据库变更成员状态为visitor,participant
				nick := iq.Elements().Child("query").Elements().Child("item").Attributes().Get("nick")
				var role int8
				if optype == visitor {
					role = none
				} else {
					role = member
				}
				rm := roommodel.RoommemberItem{nick, iq.ToJID().Node(), nick + "@" + iq.FromJID().Domain() + "/" + iq.FromJID().Resource(), false, role}
				storage.Instance().InsertOrUpdateRoommemberItem(&rm)

				if router.UserStreams(nick) == nil {
					return errors.New("用户信息不存在")
				} else {
					stm = router.UserStreams(nick)[0]
					elem, err = newPresenceElement(iq, optype, true)
					stm.SendElement(elem)
				}
			}
		}
	case "stop": //阻止使用者

	case "ban": //禁止使用者

	case "Affiliation": //授予角色
		switch iq.Attributes().Get("type") {
		case "membership": //授予成员

		case "emcee": //授予主持人

		case "admin": //授予管理者

		case "owner": //授予所有者

		default: //异常处理
			return nil
		}
	case "user": //邀请使用者

	default: //异常处理
		return nil
	}
	return nil
}

func (x *ChatRoom) sendServerUptime(iq *xmpp.IQ, stm stream.C2S) {
	secs := int(time.Duration(time.Now().UnixNano()-x.startTime.UnixNano()) / time.Second)
	x.sendReply(iq, secs, "", stm)
}

func (x *ChatRoom) sendReply(iq *xmpp.IQ, secs int, status string, stm stream.C2S) {
	q := xmpp.NewElementNamespace("query", chatRoomNamespace)
	q.SetText(status)
	q.SetAttribute("seconds", strconv.Itoa(secs))
	res := iq.ResultIQ()
	res.AppendElement(q)
	stm.SendElement(res)
}

/*func (x *ChatRoom) sendUserChatRoom(iq *xmpp.IQ, to *jid.JID, stm stream.C2S) {
	if len(router.UserStreams(to.Node())) > 0 { // user is online
		x.sendReply(iq, 0, "", stm)
		return
	}
	usr, err := storage.Instance().FetchUser(to.Node())
	if err != nil {
		log.Error(err)
		stm.SendElement(iq.InternalServerError())
		return
	}
	if usr == nil {
		stm.SendElement(iq.ItemNotFoundError())
		return
	}
	var secs int
	var status string
	if p := usr.LastPresence; p != nil {
		secs = int(time.Duration(time.Now().UnixNano()-usr.LastPresenceAt.UnixNano()) / time.Second)
		if st := p.Elements().Child("status"); st != nil {
			status = st.Text()
		}
	}
	x.sendReply(iq, secs, status, stm)
}*/

// TODO 条目服务的两种注册方式：1-组件来初始化；2-服务让disco来注册
/*func (c *ChatRoom) registerDiscoInfo(disco *xep0030.DiscoInfo) {
	disco.RegisterServerItem(xep0030.Item{Jid: c.Host(), Name: mucServiceName})
	disco.RegisterProvider(c.Host(), &mucInfoProvider{})
}

func (c *ChatRoom) unregisterDiscoInfo() {
	c.discoInfo.UnregisterServerItem(xep0030.Item{Jid: c.Host(), Name: mucServiceName})
	c.discoInfo.UnregisterProvider(c.Host())
}*/
