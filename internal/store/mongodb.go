package store

import (
	"card-service/internal/models"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type Store struct {
	Client       *mongo.Client
	Db           *mongo.Database
	Customers    *mongo.Collection
	Accounts     *mongo.Collection
	Cards        *mongo.Collection
	Transactions *mongo.Collection
	logger       *zap.Logger
}

// NewStore initializes a new Store instance with the provided MongoDB client and database name.
func NewStore(dsn, dbName string) (*Store, error) {
	//connect to mongoDb
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(dsn))
	if err != nil {
		return nil, err
	}

	// initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	//initialize database and collection
	db := client.Database(dbName)
	store := &Store{
		Client:       client,
		Db:           db,
		Customers:    db.Collection("customers"),
		Accounts:     db.Collection("accounts"),
		Cards:        db.Collection("cards"),
		Transactions: db.Collection("transactions"),
		logger:       logger,
	}

	//create indexes for efficient queries
	store.createIndexes()
	return store, nil
}

// createIndex sets up indexes for fast lookups on the customers collection.
func (s *Store) createIndexes() {
	// Index on customer ID, email, and virtual account ID
	s.Customers.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "customerId", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "sub_account_id", Value: 1}}, Options: options.Index().SetUnique(true)},
	})
	s.Cards.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "_id", Value: 1}}, Options: options.Index().SetUnique(true)},
	})
}

// Close disconnects the MongoDB client.
func (s *Store) Close() {
	s.Client.Disconnect(context.Background())
}

// Insert Transaction inserts a new transaction into the transactions collection.
func (s *Store) InsertTransaction(ctx context.Context, transaction models.Transaction) (*mongo.InsertOneResult, error) {
	return s.Transactions.InsertOne(ctx, transaction)
}

// Insert update transaction updates an existing transaction in the transactions collection.
func (s *Store) UpdateTransaction(ctx context.Context, authorizationID string, update bson.M) (*mongo.UpdateResult, error) {
	return s.Transactions.UpdateOne(ctx, bson.M{"authorizationId": authorizationID}, update)

}
func (s *Store) GetTransaction(ctx context.Context, authorizationID string) (*models.Transaction, error) {
	var transaction models.Transaction
	err := s.Transactions.FindOne(ctx, bson.M{"authorizationId": authorizationID}).Decode(&transaction)
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}
