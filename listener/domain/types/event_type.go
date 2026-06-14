package types

type EventType string

const (
	FileCreatedEvent EventType = "FILE_CREATED"
	FileWritedEvent  EventType = "FILE_WRITED"
	FileRemovedEvent EventType = "FILE_REMOVED"
	FileMovedEvent   EventType = "FILE_MOVED"
	FileRenamedEvent EventType = "FILE_RENAMED"

	DirectoryCreatedEvent EventType = "FILE_CREATED"
	DirectoryWritedEvent  EventType = "FILE_WRITED"
	DirectoryRemovedEvent EventType = "FILE_REMOVED"
	DirectoryMovedEvent   EventType = "FILE_MOVED"
	DirectoryRenamedEvent EventType = "FILE_RENAMED"

	SnapshotEvent         EventType = "SNAPSHOT"
	HostDisconnectedEvent EventType = "HOST_DISCONNECTED"
	RoomCreatedEvent      EventType = "ROOM_CREATED"
)
