/*
 * MySql数据持久化.
 * See the LICENSE file for more information.
 */

package sql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	_ "github.com/go-sql-driver/mysql" // SQL driver
	"github.com/ortuman/jackal/log"
	"github.com/ortuman/jackal/pool"
	// 引入数据库持久化框架gorm（建表用）
	"github.com/jinzhu/gorm"
	//_ "github.com/jinzhu/gorm/dialects/mysql"      //重复引用，上面已经有mysql，实际上他们是同等的
)

var (
	nowExpr = sq.Expr("NOW()")
)

type rowScanner interface {
	Scan(...interface{}) error
}

type rowsScanner interface {
	rowScanner
	Next() bool
}

// 数据库配置.
type Config struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	PoolSize int    `yaml:"pool_size"`
}

// 持久化数据结构.
type Storage struct {
	db     *sql.DB
	gormdb *gorm.DB //建表使用
	pool   *pool.BufferPool
	doneCh chan chan bool
}

// 持久化存储实例.
func New(cfg *Config) *Storage {
	var err error
	s := &Storage{
		pool:   pool.NewBufferPool(),
		doneCh: make(chan chan bool),
	}
	host := cfg.Host
	user := cfg.User
	pass := cfg.Password
	db := cfg.Database
	poolSize := cfg.PoolSize

	//初始化数据库原生db
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", user, pass, host, db)
	s.db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("%v", err)
	}
	s.db.SetMaxOpenConns(poolSize) // set max opened connection count

	if err := s.db.Ping(); err != nil {
		log.Fatalf("%v", err)
	}

	s.gormdb, err = gorm.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("%v", err)
		panic(err)
	}

	//初始化http-api组件的数据库表 initBaseTable()
	//err = s.initBaseTable() //TODO(lxf)  暂时不用（增加接口请求记录表，记录请求，方便恶意访问分析及控制）
	if err != nil {
		log.Fatalf("%v", err)
	}
	go s.loop()

	return s
}

// 返回一个存储实例.
func NewMock() (*Storage, sqlmock.Sqlmock) {
	var err error
	var sqlMock sqlmock.Sqlmock
	s := &Storage{
		pool: pool.NewBufferPool(),
	}
	s.db, sqlMock, err = sqlmock.New()
	if err != nil {
		log.Fatalf("%v", err)
	}
	return s, sqlMock
}

// 关闭持久化.
func (s *Storage) Shutdown() {
	ch := make(chan bool)
	s.doneCh <- ch
	<-ch
}

func (s *Storage) loop() {
	tc := time.NewTicker(time.Second * 15)
	defer tc.Stop()
	for {
		select {
		case <-tc.C:
			err := s.db.Ping()
			if err != nil {
				log.Error(err)
			}
		case ch := <-s.doneCh:
			s.db.Close()
			close(ch)
			return
		}
	}
}

//事物控制管理
func (s *Storage) inTransaction(f func(tx *sql.Tx) error) error {
	tx, txErr := s.db.Begin()
	if txErr != nil {
		return txErr
	}
	if err := f(tx); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}
