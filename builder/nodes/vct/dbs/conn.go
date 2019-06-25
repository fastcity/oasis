package dbs

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Conn struct {
	Host string
	Port int
	*mongo.Client
}

func New(host string, port int) *Conn {
	return &Conn{Host: host, Port: port}
}

func (con *Conn) GetConn() error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	con.Client = client
	return err
}

func (con *Conn) GetCollection(db, table string) *mongo.Collection {
	collection := con.Client.Database(db).Collection(table)
	return collection
}
