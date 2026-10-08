package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type App struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	AppID       string             `bson:"app_id" json:"app_id"`
	Name        string             `bson:"name" json:"name"`
	Mark        string             `bson:"mark" json:"mark"`
	Description string             `bson:"description" json:"description"`
	Repo        string             `bson:"repo" json:"repo"`
}

type Service struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	AppID      string             `bson:"app_id" json:"app_id"`
	ServiceID  string             `bson:"service_id" json:"service_id"`
	Runtime    string             `bson:"runtime" json:"runtime"`
	Port       int                `bson:"port" json:"port"`
	Capacity   string             `bson:"capacity" json:"capacity"`
	DeployedAt time.Time         `bson:"deployed_at" json:"deployed_at"`
}

type ServiceVersion struct {
	ServiceID string `bson:"service_id" json:"service_id"`
	Version   string `bson:"version" json:"version"`
	Digest    string `bson:"digest" json:"digest"`
}

type Release struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	AppID      string             `bson:"app_id" json:"app_id"`
	Tag        string             `bson:"tag" json:"tag"`
	DeployedAt time.Time           `bson:"deployed_at" json:"deployed_at"`
	Services   []ServiceVersion   `bson:"services" json:"services"`
}

type ProductionConfig struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	CurrentReleaseID primitive.ObjectID `bson:"current_release_id" json:"current_release_id"`
	Status           string             `bson:"status" json:"status"`
	UpdatedAt        time.Time           `bson:"updated_at" json:"updated_at"`
}

type ProductionSecret struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Key       string              `bson:"key" json:"key"`
	Value     string              `bson:"value" json:"value"`
	UpdatedAt time.Time           `bson:"updated_at" json:"updated_at"`
}
