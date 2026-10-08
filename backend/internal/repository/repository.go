package repository

import (
	"context"

	"github.com/MorningBlossom/nexus/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepository struct {
	db *mongo.Database
}

func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{db: db}
}

// --- App Repository ---

func (r *MongoRepository) SaveApp(ctx context.Context, app *models.App) error {
	coll := r.db.Collection("apps")
	filter := bson.M{"app_id": app.AppID}
	update := bson.M{"$set": app}
	opts := options.Update().SetUpsert(true)
	_, err := coll.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *MongoRepository) GetApp(ctx context.Context, appID string) (*models.App, error) {
	var app models.App
	err := r.db.Collection("apps").FindOne(ctx, bson.M{"app_id": appID}).Decode(&app)
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *MongoRepository) ListApps(ctx context.Context) ([]*models.App, error) {
	cursor, err := r.db.Collection("apps").Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var apps []*models.App
	if err := cursor.All(ctx, &apps); err != nil {
		return nil, err
	}
	return apps, nil
}

// --- Service Repository ---

func (r *MongoRepository) SaveService(ctx context.Context, svc *models.Service) error {
	coll := r.db.Collection("services")
	filter := bson.M{"service_id": svc.ServiceID}
	update := bson.M{"$set": svc}
	opts := options.Update().SetUpsert(true)
	_, err := coll.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *MongoRepository) ListServicesByApp(ctx context.Context, appID string) ([]*models.Service, error) {
	cursor, err := r.db.Collection("services").Find(ctx, bson.M{"app_id": appID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var services []*models.Service
	if err := cursor.All(ctx, &services); err != nil {
		return nil, err
	}
	return services, nil
}

// --- Release Repository ---

func (r *MongoRepository) CreateRelease(ctx context.Context, rel *models.Release) error {
	coll := r.db.Collection("releases")
	_, err := coll.InsertOne(ctx, rel)
	return err
}

func (r *MongoRepository) GetLatestRelease(ctx context.Context, appID string) (*models.Release, error) {
	opts := options.FindOneOptions{
		Sort: bson.M{"deployed_at": -1},
	}
	var rel models.Release
	err := r.db.Collection("releases").FindOne(ctx, bson.M{"app_id": appID}, &opts).Decode(&rel)
	if err != nil {
		return nil, err
	}
	return &rel, nil
}

// --- Production Config Repository ---

func (r *MongoRepository) UpdateProdConfig(ctx context.Context, config *models.ProductionConfig) error {
	coll := r.db.Collection("production_config")
	filter := bson.M{"_id": "prod_global"}
	update := bson.M{"$set": config}
	opts := options.Update().SetUpsert(true)
	_, err := coll.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *MongoRepository) GetProdConfig(ctx context.Context) (*models.ProductionConfig, error) {
	var config models.ProductionConfig
	err := r.db.Collection("production_config").FindOne(ctx, bson.M{"_id": "prod_global"}).Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// --- Secrets Repository ---

func (r *MongoRepository) SaveSecret(ctx context.Context, secret *models.ProductionSecret) error {
	coll := r.db.Collection("production_secrets")
	filter := bson.M{"key": secret.Key}
	update := bson.M{"$set": secret}
	opts := options.Update().SetUpsert(true)
	_, err := coll.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *MongoRepository) GetSecrets(ctx context.Context) ([]*models.ProductionSecret, error) {
	cursor, err := r.db.Collection("production_secrets").Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var secrets []*models.ProductionSecret
	if err := cursor.All(ctx, &secrets); err != nil {
		return nil, err
	}
	return secrets, nil
}
