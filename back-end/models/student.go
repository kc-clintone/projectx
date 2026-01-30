package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Student struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Weaknesses     []string           `bson:"weaknesses" json:"weaknesses"`             // e.g., ["Calculus", "Grammar"]
	TotalStudyTime int                `bson:"total_study_time" json:"total_study_time"` // in minutes
}
