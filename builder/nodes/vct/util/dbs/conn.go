package dbs

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Conn struct {
	Addr string
	Port int
	*mongo.Client
}

type MongoI interface {
	GetConn() error
	GetCollection(db, table string) *mongo.Collection
	Close()
}

func New(addr string) MongoI {
	return &Conn{Addr: addr}
}

func (con *Conn) GetConn() error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	url := fmt.Sprintf("mongodb://%s", con.Addr)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	con.Client = client
	return err
}

func (con *Conn) GetCollection(db, table string) *mongo.Collection {
	collection := con.Client.Database(db).Collection(table)
	return collection
}

func CreateIndexView(coll *mongo.Collection) mongo.IndexView {
	//  coll.Indexes().CreateOne()
	return coll.Indexes()
}

func (con *Conn) Close() {
	if con.Client != nil {
		con.Client.Disconnect(context.TODO())
	}

}
