/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package storage

import (
	"github.com/ortuman/jackal/component/muc/model/roommodel"
	"github.com/ortuman/jackal/xmpp"
	"sync"

	"github.com/ortuman/jackal/log"
	//"github.com/ortuman/jackal/component/muc/storage/badgerdb"
	//"github.com/ortuman/jackal/component/muc/storage/memstorage"
	"github.com/ortuman/jackal/component/muc/storage/sql"
)

//房间持久化接口
type roomStorage interface {
	// 插入或修改一个房间.
	InsertOrUpdateRoomItem(ri *roommodel.RoomItem) error

	// 删除某房间.
	DeleteRoomItem(username, jid string) error

	// 查询所有房间列表.
	FetchRoomItems(domainname string) ([]roommodel.RoomItem, error)
}

//房间成员持久化接口
type roommemberStorage interface {
	//房间成员插入或更新
	InsertOrUpdateRoommemberItem(rm *roommodel.RoommemberItem) error

	//房间成员批量插入或更新
	InsertOrUpdateRoommemberItems(rm *[]roommodel.RoommemberItem) error

	//查询所有房间成员
	FetchRoommemberItems(arg interface{}) ([]roommodel.RoommemberItem, error)

	//查询单个房间成员
	FetchRoommemberItem(membername, roomname string) (*roommodel.RoommemberItem, error)

	//删除符合条件(未知类型参数)的房间成员
	DeleteRoommemberItem(arg interface{}) error
}

//房间配置持久化接口
type roomcfgStorage interface {
	// 查询单个房间配置.
	FetchRoomcfgItem(iq *xmpp.IQ, roomname string) (*xmpp.XElement, error)

	//插入房间配置
	InsertOrUpdateRoomcfgItem(iq xmpp.XElement, roomname string) error
}

// 实体持久化接口.
type Storage interface {
	//房间持久化实体
	roomStorage

	//房间成员持久化实体
	roommemberStorage

	//房间配置持久化实体
	roomcfgStorage

	// 关闭存储子系统.
	Shutdown()
}

var (
	instMu      sync.RWMutex
	inst        Storage
	initialized bool
)

// 初始化存储子系统.
func Initialize(cfg *Config) {
	instMu.Lock()
	defer instMu.Unlock()
	if initialized {
		return
	}
	switch cfg.Type {
	case BadgerDB:
		// TODO(lxf)暂时屏蔽
		//inst = badgerdb.New(cfg.BadgerDB)
	case MySQL:
		inst = sql.New(cfg.MySQL)
	case Memory:
		// TODO(lxf)暂时屏蔽
		//inst = memstorage.New()
	default:
		// should not be reached
		break
	}
	initialized = true
}

// Instance returns global storage sub system.
func Instance() Storage {
	instMu.RLock()
	defer instMu.RUnlock()
	if inst == nil {
		log.Fatalf("muc storage subsystem not initialized")
	}
	return inst
}

// Shutdown shuts down storage sub system.
// This method should be used only for testing purposes.
func Shutdown() {
	instMu.Lock()
	defer instMu.Unlock()
	inst.Shutdown()
	inst = nil
	initialized = false
}

// ActivateMockedError forces the return of ErrMockedError from current storage manager.
// This method should only be used for testing purposes.
func ActivateMockedError() {
	instMu.Lock()
	defer instMu.Unlock()

	// TODO（lxf）暂时不用
	//switch inst := inst.(type) {
	//case *memstorage.Storage:
	//	inst.ActivateMockedError()
	//}
}

// DeactivateMockedError disables mocked storage error from a previous activation.
// This method should only be used for testing purposes.
func DeactivateMockedError() {
	instMu.Lock()
	defer instMu.Unlock()

	// TODO（lxf）暂时不用
	//switch inst := inst.(type) {
	//case *memstorage.Storage:
	//	inst.DeactivateMockedError()
	//}
}
