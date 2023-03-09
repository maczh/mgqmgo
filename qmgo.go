package mgqmgo

import (
	"context"
	"errors"
	"fmt"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/levigross/grequests"
	"github.com/qiniu/qmgo"
	"github.com/sadlil/gologger"
	"strings"
)

var Mongo = &mongo{}

type mongo struct {
	multi   bool
	conns   map[string]connection
	tags    []string
	max     int
	conf    *koanf.Koanf
	confUrl string
	pool    *poolConfig
}

type connection struct {
	conn *qmgo.Client
	db   string
	url  string
}

type poolConfig struct {
	Min     uint64
	Max     uint64
	Idle    int64
	Timeout int64
}

var logger = gologger.GetLogger()

func (m *mongo) Init(mongodbConfigUrl string) {
	if mongodbConfigUrl != "" {
		m.confUrl = mongodbConfigUrl
	}
	if m.confUrl == "" {
		logger.Error("MongoDB配置Url为空")
		return
	}
	m.tags = make([]string, 0)
	if m.conns == nil {
		m.conns = make(map[string]connection)
	}
	if len(m.conns) == 0 {
		if m.conf == nil {
			resp, err := grequests.Get(m.confUrl, nil)
			if err != nil {
				logger.Error("MongoDB配置下载失败! " + err.Error())
				return
			}
			m.conf = koanf.New(".")
			err = m.conf.Load(rawbytes.Provider([]byte(resp.String())), yaml.Parser())
			if err != nil {
				logger.Error("MongoDB配置解析错误:" + err.Error())
				m.conf = nil
				return
			}
		}
		pool := poolConfig{
			Min:     uint64(m.conf.Int64("go.data.mongo_pool.min")),
			Max:     uint64(m.conf.Int64("go.data.mongo_pool.max")),
			Idle:    m.conf.Int64("go.data.mongo_pool.idle"),
			Timeout: m.conf.Int64("go.data.mongo_pool.timeout"),
		}
		if pool.Min == 0 {
			pool.Min = 1
		}
		if pool.Max == 0 {
			pool.Min = 10 * pool.Min
		}
		if pool.Idle == 0 {
			pool.Idle = 300
		}
		if pool.Timeout == 0 {
			pool.Timeout = 60
		}
		m.pool = &pool
		m.multi = m.conf.Bool("go.data.mongodb.multidb")
		if m.multi {
			dbNames := strings.Split(m.conf.String("go.data.mongodb.dbNames"), ",")
			for _, dbName := range dbNames {
				if dbName != "" && m.conf.Exists(fmt.Sprintf("go.data.mongodb.%s.uri", dbName)) {
					uri := m.conf.String(fmt.Sprintf("go.data.mongodb.%s.uri", dbName))
					mgoConf := qmgo.Config{
						Uri:              uri,
						ConnectTimeoutMS: &pool.Idle,
						MaxPoolSize:      &pool.Max,
						MinPoolSize:      &pool.Min,
						SocketTimeoutMS:  &pool.Timeout,
					}
					session, err := qmgo.NewClient(context.Background(), &mgoConf)
					if err != nil {
						logger.Error(dbName + " MongoDB连接错误:" + err.Error())
						continue
					}
					m.conns[dbName] = connection{
						conn: session,
						db:   m.conf.String(fmt.Sprintf("go.data.mongodb.%s.db", dbName)),
						url:  uri,
					}
					m.tags = append(m.tags, dbName)
				}
			}
		} else {
			mgoConf := qmgo.Config{
				Uri:              m.conf.String("go.data.mongodb.uri"),
				ConnectTimeoutMS: &pool.Idle,
				MaxPoolSize:      &pool.Max,
				MinPoolSize:      &pool.Min,
				SocketTimeoutMS:  &pool.Timeout,
			}
			conn, err := qmgo.NewClient(context.Background(), &mgoConf)
			if err != nil {
				logger.Error("MongoDB连接错误:" + err.Error())
				return
			}
			m.conns["0"] = connection{
				conn: conn,
				db:   m.conf.String("go.data.mongodb.db"),
				url:  m.conf.String("go.data.mongodb.uri"),
			}
		}
	}
}

func (m *mongo) Close() {
	if m.multi {
		for k, _ := range m.conns {
			m.conns[k].conn.Close(context.Background())
			delete(m.conns, k)
		}
	} else {
		m.conns["0"].conn.Close(context.Background())
		delete(m.conns, "0")
	}
}

func (m *mongo) mgoCheck(tag string) error {
	if len(m.conns) == 0 {
		m.Init("")
	}
	if m.conns[tag].conn.Ping(30) != nil {
		uri := m.conns[tag].url
		db := m.conns[tag].db
		m.conns[tag].conn.Close(context.Background())
		mgoConf := qmgo.Config{
			Uri:              uri,
			ConnectTimeoutMS: &m.pool.Idle,
			MaxPoolSize:      &m.pool.Max,
			MinPoolSize:      &m.pool.Min,
			SocketTimeoutMS:  &m.pool.Timeout,
		}
		session, err := qmgo.NewClient(context.Background(), &mgoConf)
		if err != nil {
			logger.Error(tag + " MongoDB连接错误:" + err.Error())
			return err
		}
		m.conns[tag] = connection{
			conn: session,
			db:   db,
			url:  uri,
		}
	}
	return nil
}

func (m *mongo) Check() error {
	var err error
	if len(m.conns) == 0 {
		m.Init("")
	}
	if m.multi {
		for dbName, _ := range m.conns {
			err = m.mgoCheck(dbName)
			if err != nil {
				logger.Error(dbName + "连接检查失败:" + err.Error())
				continue
			}
		}
	} else {
		err = m.mgoCheck("0")
	}
	return err
}

func (m *mongo) GetConnection(dbName ...string) (*Database, error) {
	if m.multi {
		if len(dbName) > 1 || len(dbName) == 0 {
			return nil, errors.New("Multidb mongodb get connection must be specified one dbName")
		}
		if dbName[0] == "" {
			dbName[0] = m.tags[0]
		}
		if _, ok := m.conns[dbName[0]]; !ok {
			return nil, errors.New("MongoDB multidb db name invalid")
		}
		err := m.mgoCheck(dbName[0])
		if err != nil {
			return nil, err
		}
		return &Database{
			db:   m.conns[dbName[0]].conn.Database(m.conns[dbName[0]].db),
			conn: m.conns[dbName[0]].conn,
			ctx:  context.Background(),
		}, nil
	} else {
		m.Check()
		if len(m.conns) == 0 {
			return nil, errors.New("mongodb connection failed")
		}
		return &Database{
			db:   m.conns["0"].conn.Database(m.conns["0"].db),
			conn: m.conns["0"].conn,
			ctx:  context.Background(),
		}, nil
	}
}

func (m *mongo) ReturnConnection(conn *Database) {
}

func (m *mongo) IsMultiDB() bool {
	return m.multi
}

func (m *mongo) ListConnNames() []string {
	return m.tags
}
