package mgqmgo

import (
	"errors"
	"github.com/qiniu/qmgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MgoDao 注意使用前必须先将CollectionName赋值
type MgoDao[E any] struct {
	CollectionName string
	Tag            func() string
}

// Insert mongo动态插入数据
func (m *MgoDao[E]) Insert(entity *E) error {
	if m.CollectionName == "" {
		return errors.New("CollectionName未定义")
	}
	conn, err := Mongo.GetConnection(m.Tag())
	if err != nil {
		logger.Error("数据库连接失败: " + err.Error())
		return errors.New("数据库连接失败")
	}
	err = conn.C(m.CollectionName).Insert(entity)
	if err != nil {
		logger.Error("数据库插入失败: " + err.Error())
		return errors.New("数据库插入失败")
	}
	return nil
}

// Delete mongo动态删除数据
func (m *MgoDao[E]) Delete(query bson.M) error {
	if m.CollectionName == "" {
		return errors.New("CollectionName未定义")
	}
	conn, err := Mongo.GetConnection(m.Tag())
	if err != nil {
		logger.Error("数据库连接失败: " + err.Error())
		return errors.New("数据库连接失败")
	}
	err = conn.C(m.CollectionName).Remove(query)
	if err != nil {
		logger.Error("数据库删除失败: " + err.Error())
		return errors.New("数据库删除失败")
	}
	return nil
}

// Updates mongo动态更新数据
func (m *MgoDao[E]) Updates(id primitive.ObjectID, fields bson.M) error {
	if m.CollectionName == "" {
		return errors.New("CollectionName未定义")
	}
	conn, err := Mongo.GetConnection(m.Tag())
	if err != nil {
		logger.Error("数据库连接失败: " + err.Error())
		return errors.New("数据库连接失败")
	}
	err = conn.C(m.CollectionName).UpdateId(id, fields)
	if err != nil {
		logger.Error("数据库更新失败: " + err.Error())
		return errors.New("数据库更新失败")
	}
	return nil
}

// All mongo动态查询数据
func (m *MgoDao[E]) All(query bson.M) ([]E, error) {
	if m.CollectionName == "" {
		return nil, errors.New("CollectionName未定义")
	}
	conn, err := Mongo.GetConnection(m.Tag())
	if err != nil {
		logger.Error("数据库连接失败: " + err.Error())
		return nil, errors.New("数据库连接失败")
	}

	var result = make([]E, 0)
	err = conn.C(m.CollectionName).Find(query).All(&result)
	if err != nil {
		logger.Error("数据库查询失败: " + err.Error())
		return nil, errors.New("数据库查询失败")
	}
	return result, nil
}

// One mongo动态查询一条数据
func (m *MgoDao[E]) One(query bson.M) (*E, error) {
	if m.CollectionName == "" {
		return nil, errors.New("CollectionName未定义")
	}
	conn, err := Mongo.GetConnection(m.Tag())
	if err != nil {
		logger.Error("数据库连接失败: " + err.Error())
		return nil, errors.New("数据库连接失败")
	}
	var result *E
	err = conn.C(m.CollectionName).Find(query).One(result)
	if err != nil {
		if err == qmgo.ErrNoSuchDocuments {
			return nil, nil
		}
		logger.Error("数据库查询失败: " + err.Error())
		return nil, errors.New("数据库查询失败")
	}
	return result, nil
}

type ResultPage struct {
	Count int `json:"count"` //总页数
	Index int `json:"index"` //页号
	Size  int `json:"size"`  //分页大小
	Total int `json:"total"` //总记录数
}

// Pager mongo简单分页查询数据
func (m *MgoDao[E]) Pager(query bson.M, sort []string, page, size int) ([]E, *ResultPage, error) {
	if m.CollectionName == "" {
		return nil, nil, errors.New("CollectionName未定义")
	}
	conn, err := Mongo.GetConnection(m.Tag())
	if err != nil {
		logger.Error("数据库连接失败: " + err.Error())
		return nil, nil, errors.New("数据库连接失败")
	}
	defer Mongo.ReturnConnection(conn)
	// 默认分页大小为20条
	if size == 0 {
		size = 20
	}
	var result = make([]E, 0)
	var count int
	var p = ResultPage{
		Index: page,
		Size:  size,
	}
	count, err = conn.C(m.CollectionName).Count()
	if err != nil {
		logger.Error("数据库查询失败: " + err.Error())
		return nil, nil, errors.New("数据库查询失败")
	}
	p.Total = count
	p.Count = (count / size) + 1
	if count == 0 || count < (page-1)*size {
		return result, &p, err
	}
	q := conn.C(m.CollectionName).Find(query)
	if sort != nil && len(sort) > 0 {
		q = q.Sort(sort...)
	}
	err = q.Skip(int64((page - 1) * size)).Limit(int64(size)).All(&result)
	if err != nil {
		logger.Error("数据库查询失败: " + err.Error())
		return nil, nil, errors.New("数据库查询失败")
	}
	return result, &p, nil
}
