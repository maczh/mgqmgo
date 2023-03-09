package mgqmgo

import (
	"context"
	"github.com/qiniu/qmgo"
	"go.mongodb.org/mongo-driver/bson"
)

type Database struct {
	db   *qmgo.Database
	conn *qmgo.Client
	ctx  context.Context
}

type Collection struct {
	col *qmgo.Collection
	db  *Database
	ctx context.Context
}

func (d *Database) Context(ctx context.Context) *Database {
	d.ctx = ctx
	return d
}

func (d *Database) Qmgo() *qmgo.Database {
	return d.db
}

func (d *Database) Client() *qmgo.Client {
	return d.conn
}

func (d *Database) C(name string) *Collection {
	return &Collection{col: d.db.Collection(name), db: d, ctx: d.ctx}
}

func (d *Database) DropDatabase() error {
	return d.db.DropDatabase(d.ctx)
}

func (d *Database) Run(cmd interface{}, result interface{}) error {
	return d.db.RunCommand(d.ctx, cmd).Decode(&result)
}

func (d *Database) Session() *qmgo.Session {
	s, _ := d.conn.Session()
	return s
}

func (c *Collection) Qmgo() *qmgo.Collection {
	return c.col
}

func (c *Collection) DB() *Database {
	return c.db
}

func (c *Collection) Context(ctx context.Context) *Collection {
	c.ctx = ctx
	return c
}

func (c *Collection) Bulk() *qmgo.Bulk {
	return c.col.Bulk()
}

func (c *Collection) Count() (int, error) {
	count, err := c.Find(bson.M{}).Count()
	return int(count), err
}

func (c *Collection) DropCollection() error {
	return c.col.DropCollection(c.ctx)
}

func (c *Collection) DropIndex(key ...string) error {
	return c.col.DropIndex(c.ctx, key)
}

func (c *Collection) DropIndexName(name string) error {
	return c.DropIndex(name)
}

func (c *Collection) FindId(id interface{}) qmgo.QueryI {
	return c.Find(bson.M{"_id": id})
}

func (c *Collection) Find(query interface{}) qmgo.QueryI {
	return c.col.Find(c.ctx, query)
}

func (c *Collection) Insert(docs ...interface{}) error {
	var err error
	if len(docs) > 1 {
		_, err = c.col.InsertOne(c.ctx, docs[0])
	} else {
		_, err = c.col.InsertMany(c.ctx, docs)
	}
	return err
}

func (c *Collection) Remove(selector interface{}) error {
	return c.col.Remove(c.ctx, selector)
}

func (c *Collection) RemoveId(id interface{}) error {
	return c.col.RemoveId(c.ctx, id)
}

func (c *Collection) RemoveAll(selector interface{}) (*qmgo.DeleteResult, error) {
	return c.col.RemoveAll(c.ctx, selector)
}

func (c *Collection) Update(selector interface{}, update interface{}) error {
	return c.col.UpdateOne(c.ctx, selector, update)
}

func (c *Collection) UpdateAll(selector interface{}, update interface{}) (*qmgo.UpdateResult, error) {
	return c.col.UpdateAll(c.ctx, selector, update)
}

func (c *Collection) UpdateId(id interface{}, update interface{}) error {
	return c.col.UpdateId(c.ctx, id, update)
}

func (c *Collection) Upsert(selector interface{}, update interface{}) (*qmgo.UpdateResult, error) {
	return c.col.Upsert(c.ctx, selector, update)
}

func (c *Collection) UpsertId(id interface{}, update interface{}) (*qmgo.UpdateResult, error) {
	return c.col.UpsertId(c.ctx, id, update)
}
