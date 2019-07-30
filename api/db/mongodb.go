package db

import (
	"context"
	"fmt"
	"time"

	"github.com/astaxie/beego"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Conn struct {
	Addr   string
	Port   int
	dbName string
	err    error
	*mongo.Client
	*mongo.Collection
	*mongo.Database
}

type MongoInterface interface {
	// SetDBName(string) MongoInterface
	GetConn() MongoInterface
	ConnDatabase(string) MongoInterface
	ConnCollection(string) *mongo.Collection
}

func Init() MongoInterface {
	host := beego.AppConfig.String("mongohost")
	port := beego.AppConfig.String("mongoport")
	mongodbName := beego.AppConfig.String("mongodbName")
	// url := fmt.Sprintf("%s:%s/%s", host, port, mongodbName)
	url := fmt.Sprintf("%s:%s", host, port)
	return New(url).GetConn().ConnDatabase(mongodbName)
}

func New(addr string) MongoInterface {
	return &Conn{Addr: addr}
}

// func (con *Conn) SetDBName(name string) MongoInterface {
// 	con.db = name
// 	return con
// }

func (con *Conn) GetConn() MongoInterface {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	url := fmt.Sprintf("mongodb://%s", con.Addr)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	con.Client = client
	con.err = err
	return con
}

func (con *Conn) ConnDatabase(dbs string) MongoInterface {
	if con.err != nil {
		return con
	}
	if dbs == "" {
		db := con.Client.Database(con.dbName)
		con.Database = db
	} else {
		con.dbName = dbs
		db := con.Client.Database(dbs)
		con.Database = db
	}

	return con
}

func (con *Conn) ConnCollection(table string) *mongo.Collection {
	if con.err != nil {
		return nil
	}
	// con.Client.Get
	collection := con.Database.Collection(table)
	con.Collection = collection
	return collection
}
