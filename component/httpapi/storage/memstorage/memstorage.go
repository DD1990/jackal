/*
 * 内存持久化.
 * See the LICENSE file for more information.
 */

package memstorage

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/ortuman/jackal/model"
	"github.com/ortuman/jackal/model/rostermodel"
	"github.com/ortuman/jackal/xmpp"
)

// 当激活模拟错误时，任何存储方法都会返回ErrMockedError.
var ErrMockedError = errors.New("storage mocked error")

// 内存持久化数据结构.
type Storage struct {
	mockErr             uint32
	mu                  sync.RWMutex
	users               map[string]*model.User
	rosterItems         map[string][]rostermodel.Item
	rosterVersions      map[string]rostermodel.Version
	rosterNotifications map[string][]rostermodel.Notification
	vCards              map[string]xmpp.XElement
	privateXML          map[string][]xmpp.XElement
	offlineMessages     map[string][]*xmpp.Message
	blockListItems      map[string][]model.BlockListItem
}

// 返回内存持久化存储实例.
func New() *Storage {
	return &Storage{
		users:               make(map[string]*model.User),
		rosterItems:         make(map[string][]rostermodel.Item),
		rosterVersions:      make(map[string]rostermodel.Version),
		rosterNotifications: make(map[string][]rostermodel.Notification),
		vCards:              make(map[string]xmpp.XElement),
		privateXML:          make(map[string][]xmpp.XElement),
		offlineMessages:     make(map[string][]*xmpp.Message),
		blockListItems:      make(map[string][]model.BlockListItem),
	}
}

// 关闭持久化实例.
func (m *Storage) Shutdown() {
}

// 处理内存中的模拟错误.
func (m *Storage) ActivateMockedError() {
	atomic.StoreUint32(&m.mockErr, 1)
}

// 在内存中禁用模拟错误.
func (m *Storage) DeactivateMockedError() {
	atomic.StoreUint32(&m.mockErr, 0)
}

func (m *Storage) inWriteLock(f func() error) error {
	if atomic.LoadUint32(&m.mockErr) == 1 {
		return ErrMockedError
	}
	m.mu.Lock()
	err := f()
	m.mu.Unlock()
	return err
}

func (m *Storage) inReadLock(f func() error) error {
	if atomic.LoadUint32(&m.mockErr) == 1 {
		return ErrMockedError
	}
	m.mu.RLock()
	err := f()
	m.mu.RUnlock()
	return err
}
