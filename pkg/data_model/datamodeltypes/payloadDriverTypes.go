package datamodeltypes

type CommitType func(isCompletePayload bool)
type RejectType func()
