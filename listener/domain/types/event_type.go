package types

type EventType string

const (
	FileCreatedEvent EventType = "FILE_CREATED"
	FileWritedEvent  EventType = "FILE_WRITED"
	FileRemovedEvent EventType = "FILE_REMOVED"
	FileMovedEvent   EventType = "FILE_MOVED"
	FileRenamedEvent EventType = "FILE_RENAMED"

	DirectoryCreatedEvent EventType = "DIRECTORY_CREATED"
	DirectoryWritedEvent  EventType = "DIRECTORY_WRITED"
	DirectoryRemovedEvent EventType = "DIRECTORY_REMOVED"
	DirectoryMovedEvent   EventType = "DIRECTORY_MOVED"
	DirectoryRenamedEvent EventType = "DIRECTORY_RENAMED"

	SnapshotCreateEvent   EventType = "SNAPSHOT_CREATE"
	SnapshotSyncEvent     EventType = "SNAPSHOT_SYNC"
	HostDisconnectedEvent EventType = "HOST_DISCONNECTED"
	RoomCreatedEvent      EventType = "ROOM_CREATED"
)
