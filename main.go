package main

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/arpsch/xm/server"
	"github.com/arpsch/xm/store/mongo"
)

func main() {
	err := doMain()
	if err != nil {
		log.Fatal(err)
	}
}

func doMain() error {

	ctx := context.Background()

	mgoUrl, err := url.Parse("mongodb://127.0.0.1:27017")
	if err != nil {
		log.Fatal(err)
	}

	storeConfig := mongo.MongoStoreConfig{
		MongoURL: mgoUrl,
		DbName:   "xm",
	}
	ds, err := mongo.NewMongoStore(context.Background(), storeConfig)
	if err != nil {
		log.Fatal(err)
	}

	defer ds.Close(ctx)
	fmt.Printf("Starting the server...")
	return server.InitAndRun(ds)
}
