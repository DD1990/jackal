package pubsubHandle

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"github.com/ortuman/jackal/component/httpapi/checkAuth"
	aer "github.com/ortuman/jackal/component/httpapi/errors"
	"github.com/ortuman/jackal/component/httpapi/handlerImpl/model"
	"github.com/ortuman/jackal/component/httpapi/storage"
	"github.com/ortuman/jackal/log"
	"github.com/ortuman/jackal/module"
	"github.com/ortuman/jackal/pool"
	"github.com/ortuman/jackal/router"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
	"io/ioutil"
	"net/http"
	"strings"
)

type PubSubHandlerface struct {
	aer.ApiErr
	Node   string `xml:"node"`
	Owner  string `xml:"owner"`
	Pubsub Pub    `xml:"pubsub"`
	Jids   Jid    `xml:"jids"`

	ItemId   string `xml:"item-id"`
	ExpireAt string `xml:"expire-at"`
	Entry    SEntry `xml:"entry"`
}
type Pub struct {
	Prefix   bool   `xml:"prefix,attr"`
	NodeType string `xml:"node_type"`
}
type Jid struct {
	Value []string `xml:"value"`
}
type SEntry struct {
	ItemEntry SItemEntry `xml:"item-entry"`
}

type SItemEntry struct {
	Title   string `xml:"title"`
	Author  string `xml:"author"`
	Content string `xml:"content"`
}

/*
<data>
	<node>node1_node</node>
	<owner>test7@localhost</owner>
	<pubsub prefix=\"true\">
		<node_type>leaf</node_type>
	</pubsub>
</data>
*/
func (ps *PubSubHandlerface) PubSubManage(w http.ResponseWriter, r *http.Request) {

	log.Infof("进入 PubSubManage()")
	//认证失败，返回
	if !checkAuth.CheckBasicAuth(w, r) {
		ps.HandleErr(&w, aer.AUTH_ERR_CODE, nil)
		return
	}

	defer r.Body.Close()
	con, _ := ioutil.ReadAll(r.Body) //获取body体数据
	bodyContent := string(con)

	log.Infof("请求包体：", strings.TrimSpace(bodyContent))
	//解析xml
	/*原生态方式*/
	pstemp := PubSubHandlerface{}
	xml.Unmarshal(con, &pstemp)
	ps = &pstemp

	/*jackal封装xep0004,缺点：无法解析换行的xml*/
	//parser := xmpp.NewParser(strings.NewReader(bodyContent), xmpp.DefaultMode, 0)
	//elem, elemerr := parser.ParseElement()
	//if elemerr != nil {
	//	log.Error(elemerr)
	//	return
	//}
	//domain.Name = elem.Elements().Child("fields").Elements().Children("item")[0].Attributes().Get("var")
	//user = &user1
	var err error
	if &r.RequestURI != nil && r.RequestURI != "" {

		if uris := strings.Split(r.RequestURI, "/"); len(uris) < 2 {
			ps.HandleErr(&w, aer.URI_ERR_CODE, nil)
		} else {
			switch uris[len(uris)-1] {
			case model.CREATENODE:
				err = ps.addNode()
				if err != nil {
					ps.HandleErr(&w, aer.ADD_NODE_ERR_CODE, err)
					return
				} else {
					goto RESPONSE
				}
			case model.SUBSCRIBENODE: //用户订阅
				//user.StringsToUser(uris)
				err = ps.subscribeNode()
				if err != nil {
					ps.HandleErr(&w, aer.SUBSCRIBE_NODE_ERR_CODE, err)
					return
				} else {
					goto RESPONSE
				}
			case model.PUBLISHITEM: //消息群发
				//user.StringsToUser(uris)
				err = ps.sendMessageMass(bodyContent)
				if err != nil {
					ps.HandleErr(&w, aer.PUBLISH_ITEM_ERR_CODE, err)
					return
				} else {
					goto RESPONSE
				}
			case model.DELETENODE:
				err = ps.deleteNode()

			}
		}

	}

RESPONSE:
	w.Write([]byte("处理成功\n"))
	w.Write([]byte("处理结果："))
}

func (ps *PubSubHandlerface) addNode() error {
	log.Infof("进入 addNode()")
	r := sha1.Sum([]byte(ps.Node))
	jr := sha1.Sum([]byte(ps.Owner))
	var nodeType int8 = 0
	if ps.Pubsub.NodeType == "leaf" {
		nodeType = 1
	}

	jid := &model.Jid{
		Jid:    ps.Owner,
		JidSha: hex.EncodeToString(jr[:]),
	}
	err := storage.Instance().InsertOrUpdateJid(jid)

	jid, err = storage.Instance().FetchJid(jid.Jid)
	if err != nil {
		log.Error(err)
		return err
	}
	var dom *model.Domain
	dom, err = storage.Instance().FetchDomain(strings.Split(ps.Owner, "@")[1])

	node := model.Node{
		ServiceId:   dom.ServiceId,            //虚拟域Id
		NodeName:    ps.Node,                  //节点名称
		NodeNameSha: hex.EncodeToString(r[:]), //节点名称加密
		Type:        nodeType,                 //节点类型
		Title:       "",                       //标题
		Description: "",                       //描述
		CreatorId:   jid.JidId,                //创建者Id
		//CreationDate:  time.Time,            //创建时间
		Configuration: `"<x xmlns="jabber:x:data" type="form"><field type="hidden" var="FORM_TYPE"><value>http://jabber.org/protocol/pubsub#node_config</value></field><field type="list-single" var="pubsub#node_type"><value>leaf</value><option><value>leaf</value></option><option><value>collection</value></option></field><field type="text-single" var="pubsub#title" label="A friendly name for the node"><value></value></field><field type="boolean" var="pubsub#deliver_payloads" label="Whether to deliver payloads with event notifications"><value>1</value></field><field type="boolean" var="pubsub#notify_config" label="Notify subscribers when the node configuration changes"><value>0</value></field><field type="boolean" var="pubsub#persist_items" label="Persist items to storage"><value>1</value></field><field type="text-single" var="pubsub#max_items" label="Max # of items to persist"><value>10</value></field><field type="text-single" var="pubsub#collection" label="The collection with which a node is affiliated"><value></value></field><field type="list-single" var="pubsub#access_model" label="Specify the subscriber model"><value>open</value><option><value>authorize</value></option><option><value>open</value></option><option><value>presence</value></option><option><value>roster</value></option><option><value>whitelist</value></option></field><field type="list-single" var="pubsub#publish_model" label="Specify the publisher model"><value>publishers</value><option><value>open</value></option><option><value>publishers</value></option><option><value>subscribers</value></option></field><field type="list-single" var="pubsub#send_last_published_item" label="When to send the last published item"><value>on_sub</value><option><value>never</value></option><option><value>on_sub</value></option><option><value>on_sub_and_presence</value></option></field><field type="text-multi" var="pubsub#domains" label="The domains allowed to access this node (blank for any)"/><field type="boolean" var="pubsub#presence_based_delivery" label="Whether to deliver notifications to available users only"><value>0</value></field><field type="boolean" var="tigase#presence_expired" label="Whether to subscription expired when subscriber going offline."><value>0</value></field><field type="text-multi" var="pubsub#embedded_body_xslt" label="The XSL transformation which can be applied to payloads in order to generate an appropriate message body element."/><field type="text-single" var="pubsub#body_xslt" label="The URL of an XSL transformation which can be applied to payloads in order to generate an appropriate message body element."><value></value></field><field type="text-multi" var="pubsub#roster_groups_allowed" label="Roster groups allowed to subscribe"/><field type="boolean" var="pubsub#notify_sub_aff_state" label="Notify subscribers when owner change their subscription or affiliation state"><value>1</value></field><field type="boolean" var="tigase#allow_view_subscribers" label="Allows get list of subscribers for each sybscriber"><value>0</value></field><field type="list-single" var="tigase#collection_items_odering" label="Whether to sort collection items by creation date or update time"><value>byUpdateDate</value><option><value>byCreationDate</value></option><option><value>byUpdateDate</value></option></field></x>"`, //配置
		//CollectionId:  0,                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         //
	}
	err = storage.Instance().InsertOrUpdateNode(&node)
	if err != nil {
		log.Error(err)
	}
	return nil
}

func (ps *PubSubHandlerface) getNode() error {
	return nil
}

func (ps *PubSubHandlerface) deleteNode() error {
	return nil
}

//用户订阅
func (ps *PubSubHandlerface) subscribeNode() error {
	log.Infof("进入 subscribeNode()")
	var jid *model.Jid
	var node *model.Node
	var err error

	node, err = storage.Instance().FetchNode(ps.Node)
	for _, jidValue := range ps.Jids.Value {
		//新增jids
		jr := sha1.Sum([]byte(jidValue))
		jidItem := &model.Jid{
			Jid:    jidValue,
			JidSha: hex.EncodeToString(jr[:]),
		}
		err = storage.Instance().InsertOrUpdateJid(jidItem)
		if err != nil {
			log.Error(err)
			return err
		}

		//查询jidId
		jid, err = storage.Instance().FetchJid(jidValue)
		if err != nil {
			log.Error(err)
			return err
		}

		//新增订阅Subscribe
		sub := model.Subscribe{
			NodeId:       node.NodeId,
			JidId:        jid.JidId,
			Subscription: "subscribed",
		}
		err = storage.Instance().InsertOrUpdateSubscribe(&sub)
		if err != nil {
			log.Error(err)
			return err
		}

	}
	return nil
}

//用户订阅
func (ps *PubSubHandlerface) sendMessageMass(data string) error {
	ir := sha1.Sum([]byte(ps.ItemId))
	//查询Node节点数据
	node, err := storage.Instance().FetchNode(ps.Node)
	if err != nil {
		log.Error(err)
		return err
	}

	//新增或更新Item
	item := &model.Item{
		NodeId: node.NodeId,
		Id:     ps.ItemId,
		IdSha1: hex.EncodeToString(ir[:]),
		//CreationDate time.Time //创建时间
		PublisherId: node.CreatorId, //发布者Id
		//UpdateDate   time.Time //更新时间
		Data: data,
	}
	err = storage.Instance().InsertOrUpdateItem(item)
	if err != nil {
		log.Error(err)
		return err
	}
	//查询节点所有者，暂时作为发布者
	jidOwner, err := storage.Instance().FetchJidById(node.CreatorId)

	//查询Node节点数据
	//var subscribe []model.Subscribe
	subscribes, err := storage.Instance().FetchSubscribe(node.NodeId)
	pool := pool.NewBufferPool()
	buf := pool.Get()
	defer pool.Put(buf) //释放缓冲到缓冲池
	for _, subscribe := range *subscribes {
		jidItem, err := storage.Instance().FetchJidById(subscribe.JidId)
		if err != nil {
			log.Error(err)
			return err
		}
		//封装消息
		/*
			<message to="user1@example.com" from="user2@example.com"><body>Example message</body></message>
		*/
		/*方式一*/
		messageContent := `<message to="` + jidItem.Jid + `" from="` + jidOwner.Jid + `"><body>` + ps.Entry.ItemEntry.Content + `</body></message>`
		buf.WriteString(messageContent)
		parser := xmpp.NewParser(buf, xmpp.DefaultMode, 0)
		elem, err := parser.ParseElement()
		fromJid, _ := jid.NewWithString(jidOwner.Jid, false)
		toJid, _ := jid.NewWithString(jidItem.Jid, false)
		message, err := xmpp.NewMessageFromElement(elem, fromJid, toJid)
		log.Infof("方式一封装的消息：", message.String())

		/*方式二*/
		/*		message := xmpp.NewElementName("message")
				message.SetAttribute("from", jidOwner.Jid)
				message.SetAttribute("to", jidItem.Jid)
				body := xmpp.NewElementName("body")
				body.SetText(ps.Entry.ItemEntry.Content)
				message.AppendElement(body)*/
		//判断如果成员在线，则发送
		if stms := router.UserStreams(message.ToJID().Node()); stms != nil {
			stm := stms[0]
			stm.SendElement(message)
		} else { //如果成员不在线，则发送离线消息
			//TODO(lxf) 一对一聊天可以，广播类型的离线消息未处理
			offlineServer := module.Modules().Offline
			offlineServer.ArchiveMessage(message)
		}
	}

	return nil
}
