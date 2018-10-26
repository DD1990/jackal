/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package roommodel

// 房间数据模型.
type RoomItem struct {
	Roomname     string  //房间名称
	Roomserver   string  //房间所在服务器
	Password     string  //房间密码
	Ispravite    bool    //房间是否私有
	Roomtype     int8    //房间类型
}

// 房间成员数据模型
type RoommemberItem struct {
	Membername   string  //成员名称
	Roomname     string  //房间名称
	Jid 		 string  //成员jid
	Isblack 	 bool    //成员是否在黑名单
	Role 		 int8    //成员角色
}

// 房间数据模型.
type RoomcfgItem struct {
	Roomname     string  //房间名称
	Cfgcontent   string  //配置内容
}