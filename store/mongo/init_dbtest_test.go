package mongo_test

import (
	"context"
	"net/url"
	"os"
	"testing"

	"log"

	"github.com/arpsch/xm/store/mongo"
)

var ds *mongo.MongoStore

func testSetup() error {
	var err error

	mgoUrl, err := url.Parse("mongodb://localhost:27017")
	if err != nil {
		log.Fatal(err)
	}

	storeConfig := mongo.MongoStoreConfig{
		MongoURL: mgoUrl,
		DbName:   "xm",
	}
	ds, err = mongo.NewMongoStore(context.Background(), storeConfig)
	if err != nil {
		return err
	}

	if err := ds.Ping(context.Background()); err != nil {
		return err
	}

	return nil
}

func testTeardown() error {
	ctx := context.Background()
	err := ds.DropDatabase(ctx)
	if err != nil {
		return err
	}
	ds.Close(ctx)
	return nil
}

func TestMain(m *testing.M) {
	if err := testSetup(); err != nil {
		log.Fatalf("severe setup issue :%v", err)
		os.Exit(-1)
	}

	exitCode := m.Run()

	if err := testTeardown(); err != nil {
		os.Exit(-1)
	}

	// Exit
	os.Exit(exitCode)
}
