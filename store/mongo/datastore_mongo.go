package mongo

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/arpsch/xm/model"
	"github.com/arpsch/xm/store"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"go.mongodb.org/mongo-driver/mongo/options"
	mopts "go.mongodb.org/mongo-driver/mongo/options"
)

const (
	DbName          = "xm"
	DbCompaniesColl = "companies"

	//fields
	Name = "name"
)

type MongoStoreConfig struct {
	// MongoURL holds the URL to the MongoDB server.
	MongoURL *url.URL

	// DbName contains the name of the deviceconfig database.
	DbName string
}

// newClient returns a mongo client
func newClient(ctx context.Context, config MongoStoreConfig) (*mongo.Client, error) {

	clientOptions := mopts.Client()
	if config.MongoURL == nil {
		return nil, errors.New("mongo: missing URL")
	}
	clientOptions.ApplyURI(config.MongoURL.String())

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "mongo: failed to connect with server")
	}

	// Validate connection
	if err = client.Ping(ctx, nil); err != nil {
		return nil, errors.Wrap(err, "mongo: error reaching mongo server")
	}

	return client, nil
}

// MongoStore is the data storage service
type MongoStore struct {
	// client holds the reference to the client used to communicate with the
	// mongodb server.
	client *mongo.Client

	config MongoStoreConfig
}

// SetupDataStore returns the mongo data store and optionally runs migrations
func NewMongoStore(ctx context.Context, config MongoStoreConfig) (*MongoStore, error) {
	dbClient, err := newClient(ctx, config)
	if err != nil {
		return nil, err
	}
	return &MongoStore{
		client: dbClient,
		config: config,
	}, nil
}

func (db *MongoStore) Database(ctx context.Context, opt ...*mopts.DatabaseOptions) *mongo.Database {
	return db.client.Database(db.config.DbName, opt...)
}

// Ping verifies the connection to the database
func (db *MongoStore) Ping(ctx context.Context) error {
	res := db.client.
		Database(db.config.DbName).
		RunCommand(ctx, bson.M{"ping": 1})
	return res.Err()
}

// Close disconnects the client
func (db *MongoStore) Close(ctx context.Context) error {
	err := db.client.Disconnect(ctx)
	return err
}

func (db *MongoStore) DropDatabase(ctx context.Context) error {
	err := db.client.
		Database(db.config.DbName).
		Drop(ctx)
	return err
}
func mongoOperator(co store.ComparisonOperator) string {
	switch co {
	case store.Eq:
		return "$eq"
	}
	return ""
}

func (db *MongoStore) CreateIndex(ctx context.Context, collectionName string, field string, unique bool) error {

	mod := mongo.IndexModel{
		Keys:    bson.M{field: 1},
		Options: options.Index().SetUnique(unique),
	}

	c := db.Database(ctx).Collection(DbCompaniesColl)

	_, err := c.Indexes().CreateOne(ctx, mod)
	if err != nil {
		return err
	}

	return nil
}

func (db *MongoStore) CreateCompany(ctx context.Context, comp model.Company) (string, error) {

	if err := db.CreateIndex(ctx, DbCompaniesColl, Name, true); err != nil {
		return "", err
	}

	if comp.ID == "" {
		comp.ID = primitive.NewObjectID().Hex()
	}

	now := time.Now()

	comp.CreatedTs = now
	comp.UpdatedTs = now

	c := db.Database(ctx).Collection(DbCompaniesColl)
	_, err := c.InsertOne(ctx, comp)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key error") {
			return "", store.ErrCompanyExists
		}

		return "", err
	}

	return comp.ID, nil
}

func (db *MongoStore) ListCompanies(ctx context.Context, q store.ListQuery) ([]model.Company, int, error) {
	type CompanyResult struct {
		Company []model.Company `json:"results" bson:"results"`
		Count   int             `json:"totalCount" bson:"totalCount"`
	}

	c := db.Database(ctx).Collection(DbCompaniesColl)

	queryFilters := make([]bson.M, 0)
	for _, filter := range q.Filters {
		op := mongoOperator(filter.Operator)
		if filter.ValueFloat != nil {
			queryFilters = append(queryFilters, bson.M{"$or": []bson.M{
				{filter.AttrName: bson.M{op: filter.Value}},
				{filter.AttrName: bson.M{op: filter.ValueFloat}},
			}})
		} else {
			queryFilters = append(queryFilters, bson.M{filter.AttrName: bson.M{op: filter.Value}})
		}
	}

	findQuery := bson.M{}
	if len(queryFilters) > 0 {
		findQuery["$and"] = queryFilters
	}

	filter := bson.M{
		"$match": bson.M{
			"$and": []bson.M{
				findQuery,
			},
		},
	}

	sortQuery := bson.M{"$skip": 0}
	if q.Sort != nil {
		sortFieldQuery := bson.M{}
		sortFieldQuery[q.Sort.AttrName] = 1
		if !q.Sort.Ascending {
			sortFieldQuery[q.Sort.AttrName] = -1
		}
		sortQuery = bson.M{"$sort": sortFieldQuery}
	}

	limitQuery := bson.M{"$skip": 0}
	if q.Limit > 0 {
		limitQuery = bson.M{"$limit": q.Limit}
	}

	combinedQuery := bson.M{
		"$facet": bson.M{
			"results": []bson.M{
				sortQuery,
				bson.M{"$skip": q.Skip},
				limitQuery,
			},
			"totalCount": []bson.M{
				bson.M{"$count": "count"},
			},
		},
	}

	resultMap := bson.M{
		"$project": bson.M{
			"results": 1,
			"totalCount": bson.M{
				"$ifNull": []interface{}{
					bson.M{
						"$arrayElemAt": []interface{}{"$totalCount.count", 0},
					},
					0,
				},
			},
		},
	}

	queryPipeline := []bson.M{filter}
	queryPipeline = append(queryPipeline, combinedQuery, resultMap)

	cursor, err := c.Aggregate(ctx, queryPipeline, nil)
	if err != nil {
		return []model.Company{}, 0, err
	}

	var companies []CompanyResult
	if err := cursor.All(ctx, &companies); err != nil {
		return []model.Company{}, 0, err
	}

	if len(companies) <= 0 {
		return []model.Company{}, 0, nil
	}

	return companies[0].Company, companies[0].Count, nil
}

func (db *MongoStore) GetCompany(ctx context.Context, id string) (*model.Company, error) {

	c := db.Database(ctx).Collection(DbCompaniesColl)
	res := model.Company{}

	err := c.FindOne(ctx, bson.M{"_id": id}).Decode(&res)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, store.ErrCompanyNotFound
		} else {
			return nil, errors.Wrap(err, "failed to fetch company")
		}
	}
	return &res, nil
}

func (db *MongoStore) UpdateCompany(ctx context.Context, id string, cu model.CompanyUpdate) error {
	c := db.Database(ctx).Collection(DbCompaniesColl)
	cu.UpdatedTs = time.Now()

	update := bson.M{
		"$set": cu,
	}
	res, err := c.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return errors.Wrap(err, "failed to update company")
	} else if res.MatchedCount < 1 {
		return store.ErrCompanyNotFound
	}

	return nil
}

func (db *MongoStore) DeleteCompany(ctx context.Context, id string) error {
	c := db.Database(ctx).Collection(DbCompaniesColl)

	filter := bson.M{"_id": id}
	result, err := c.DeleteOne(ctx, filter)
	if err != nil {
		return errors.Wrap(err, "failed to remove company")
	} else if result.DeletedCount < 1 {
		return store.ErrCompanyNotFound
	}

	return nil
}
