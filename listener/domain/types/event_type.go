package types

type EventType string

const (
	FileCreatedEvent EventType = "FILE_CREATED"
	FileWritedEvent  EventType = "FILE_WRITED"
	FileDeletedEvent EventType = "FILE_DELETED"
	FileMovedEvent   EventType = "FILE_MOVED"

	DirectoryCreatedEvent EventType = "DIRECTORY_CREATED"
	DirectoryDeletedEvent EventType = "DIRECTORY_DELETED"
)
