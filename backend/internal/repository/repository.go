package repository

import (
	"context"
	"time"

	"github.com/MorningBlossom/nexus/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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


// --- Environment Repository ---

func (r *MongoRepository) ListEnvironmentsByApp(ctx context.Context, appID string) ([]*models.Environment, error) {
	cursor, err := r.db.Collection("environments").Find(ctx, bson.M{"app_id": appID, "deleted": bson.M{"$ne": true}}, options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	environments := make([]*models.Environment, 0)
	if err := cursor.All(ctx, &environments); err != nil {
		return nil, err
	}
	return environments, nil
}

func (r *MongoRepository) GetEnvironment(ctx context.Context, appID string, id primitive.ObjectID) (*models.Environment, error) {
	var environment models.Environment
	if err := r.db.Collection("environments").FindOne(ctx, bson.M{"_id": id, "app_id": appID}).Decode(&environment); err != nil {
		return nil, err
	}
	return &environment, nil
}

func (r *MongoRepository) SaveEnvironment(ctx context.Context, environment *models.Environment) error {
	if environment.ID.IsZero() {
		environment.ID = primitive.NewObjectID()
	}
	if environment.CreatedAt.IsZero() {
		environment.CreatedAt = time.Now().UTC()
	}
	environment.UpdatedAt = time.Now().UTC()

	filter := bson.M{"_id": environment.ID, "app_id": environment.AppID}
	update := bson.M{"$set": bson.M{
		"app_id": environment.AppID,
		"name": environment.Name,
		"url": environment.URL,
		"secret": environment.Secret,
		"service": environment.Service,
		"artifact": environment.Artifact,
		"created_at": environment.CreatedAt,
		"updated_at": environment.UpdatedAt,
		"deleted": false,
	}}
	_, err := r.db.Collection("environments").UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	return err
}

func (r *MongoRepository) DeleteEnvironment(ctx context.Context, appID string, id primitive.ObjectID) error {
	result, err := r.db.Collection("environments").UpdateOne(ctx, bson.M{"_id": id, "app_id": appID, "deleted": bson.M{"$ne": true}}, bson.M{"$set": bson.M{"deleted": true, "updated_at": time.Now().UTC()}})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *MongoRepository) RestoreEnvironment(ctx context.Context, appID string, id primitive.ObjectID) error {
	result, err := r.db.Collection("environments").UpdateOne(ctx, bson.M{"_id": id, "app_id": appID, "deleted": true}, bson.M{"$set": bson.M{"deleted": false, "updated_at": time.Now().UTC()}})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
