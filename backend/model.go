package main

import "github.com/kc-clintone/study-coach/model"

// Type aliases to keep existing code compiling while using model definitions.
type (
	Task                 = model.Task
	Session              = model.Session
	CreateSessionRequest = model.CreateSessionRequest
	Submission           = model.Submission
	StudyPlan            = model.StudyPlan
	TopicSnapshot        = model.TopicSnapshot
	Badge                = model.Badge
	ProfileResponse      = model.ProfileResponse
	TopicImprovement     = model.TopicImprovement
)
