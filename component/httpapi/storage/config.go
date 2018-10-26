/*
 * 持久化相关配置.
 * See the LICENSE file for more information.
 */

package storage

import (
	"errors"
	"fmt"

	"github.com/ortuman/jackal/component/httpapi/storage/badgerdb"
	"github.com/ortuman/jackal/component/httpapi/storage/sql"
)

const defaultMySQLPoolSize = 16

// 持久化变量类型.
type StorageType int

const (
	// MySQL
	MySQL StorageType = iota

	// BadgerDB
	BadgerDB

	// Memory
	Memory
)

// 持久化配置的数据结构.
type Config struct {
	Type     StorageType
	MySQL    *sql.Config
	BadgerDB *badgerdb.Config
}

type storageProxyType struct {
	Type     string           `yaml:"type"`
	MySQL    *sql.Config      `yaml:"mysql"`
	BadgerDB *badgerdb.Config `yaml:"badgerdb"`
}

// 实现 unmarshal 接口？.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	p := storageProxyType{}
	if err := unmarshal(&p); err != nil {
		return err
	}
	switch p.Type {
	case "mysql":
		if p.MySQL == nil {
			return errors.New("storage.Config: couldn't read MySQL configuration")
		}
		c.Type = MySQL

		// assign storage defaults
		c.MySQL = p.MySQL
		if c.MySQL != nil && c.MySQL.PoolSize == 0 {
			c.MySQL.PoolSize = defaultMySQLPoolSize
		}

	case "badgerdb":
		if p.BadgerDB == nil {
			return errors.New("storage.Config: couldn't read BadgerDB configuration")
		}
		c.Type = BadgerDB

		c.BadgerDB = p.BadgerDB
		if len(c.BadgerDB.DataDir) == 0 {
			c.BadgerDB.DataDir = "./data"
		}

	case "memory":
		c.Type = Memory

	case "":
		return errors.New("storage.Config: unspecified storage type")

	default:
		return fmt.Errorf("storage.Config: unrecognized storage type: %s", p.Type)
	}
	return nil
}
