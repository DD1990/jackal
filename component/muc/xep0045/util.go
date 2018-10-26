/*工具包*/
package xep0045

import "github.com/ortuman/jackal/xmpp"

const (
	elementType = iota
	id
	from
	to
	eletype
	xmlns
)

//快速创建<item>标签
func newAuthItem(affiliation, nick, role string) *xmpp.Element{
	item := xmpp.NewElementName("item")
	item.SetAttribute("affiliation",affiliation)
	item.SetAttribute("nick",nick)
	item.SetAttribute("role",role)
	return item
}

//xmpp.Element带属性的快速创建，要求参数顺序必须为（elementType, ID, from, to, type, namespace）
//TODO(lxf) 继续优化，判断每个参数是否为空，为空则不处理，根据i值判断;        并增加错误处理及错误返回
func newPreOrIQElement(args ...string) *xmpp.Element{
	preOrIQ := xmpp.NewElementName(args[elementType])
	for i,value := range args {
		if value == ""{continue}
		switch i {
		case id:
			preOrIQ.SetAttribute("id",value)
		case from:
			preOrIQ.SetAttribute("from",value)
		case to:
			preOrIQ.SetAttribute("to",value)
		case eletype:
			preOrIQ.SetAttribute("type",value)
		case xmlns:
			preOrIQ.SetAttribute("xmlns","jabber:client")//TODO(lxf) 无法加载问题，但不影响功能
		}
	}
	return  preOrIQ
}

/*func (r *ChatRoom) parseVer(ver string) int {
	if len(ver) > 0 && ver[0] == 'v' {
		v, _ := strconv.Atoi(ver[1:])
		return v
	}
	return 0
}*/