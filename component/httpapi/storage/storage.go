/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package storage

import (
	md "github.com/ortuman/jackal/component/httpapi/handlerImpl/model"
	"github.com/ortuman/jackal/model"
	"sync"

	"github.com/ortuman/jackal/log"
	//"github.com/ortuman/jackal/component/httpapi/storage/badgerdb"
	//"github.com/ortuman/jackal/component/httpapi/storage/memstorage"
	"github.com/ortuman/jackal/component/httpapi/storage/sql"
)

type userStorage interface {
	// 增加或更新用户
	InsertOrUpdateUser(user *model.User) error

	// 删除用户
	DeleteUser(username string) error

	// 查询用户
	FetchUser(username string) (*model.User, error)

	// 查询用户是否存在
	UserExists(username string) (bool, error)
}

//虚拟化
type domainStorage interface {
	// 增加或更新虚拟域
	InsertOrUpdateDomain(user *md.Domain) error

	// 删除虚拟域
	DeleteDomain(username string) error

	// 查询虚拟域
	FetchDomainAll() (*[]md.Domain, error)

	// 查询虚拟域
	FetchDomain(name string) (*md.Domain, error)

	// 查询虚拟域是否存在
	DomainExists(username string) (bool, error)
}

type nodeStorage interface {
	// 增加或更新节点
	InsertOrUpdateNode(user *md.Node) error

	// 删除节点
	DeleteNode(username string) error

	// 查询节点
	FetchNode(nodename string) (*md.Node, error)

	// 查询节点是否存在
	NodeExists(username string) (bool, error)

	/********************************Jid操作***************************************/
	// 增加或更新JID
	InsertOrUpdateJid(user *md.Jid) error

	// 查询Ijd
	FetchJid(jid string) (*md.Jid, error)

	// 查询Ijd
	FetchJidById(id uint64) (*md.Jid, error)

	/********************************Subscribe操作***************************************/
	// 增加或更新订阅
	InsertOrUpdateSubscribe(user *md.Subscribe) error

	// 查询订阅
	FetchSubscribe(nodeId uint64) (*[]md.Subscribe, error)

	/********************************Item操作***************************************/
	// 增加或更新消息发布
	InsertOrUpdateItem(user *md.Item) error
}

// 实体持久化接口.
type Storage interface {
	//用户持久化实体
	userStorage

	//虚拟域持久化实体
	domainStorage

	//节点持久化实体
	nodeStorage

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
		log.Fatalf("http-api storage subsystem not initialized")
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
