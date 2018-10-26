package xep0045

//成员角色
const (
	owner = iota                      //所有者
	admin                      		  //管理员
	member                     		  //普通成员
	none                      		  //游客
	outcast                     	  //被排斥者
)

//成员权限
const (
	moderator = "moderator"           //所有者、管理员权限
	participant = "participant"       //普通成员权限
	visitor = "visitor"               //游客权限
	noneauth = "none"                 //被排斥者权限
)

//房间类型
const (
	instantroom = iota                //临时房间
	reservedroom                      //固定房间
)

//角色变更接口
//TODO(lxf) 根据实际应用补充参数句返回值
type roleiter interface {
	roleSet()
	roleUpdate()
}