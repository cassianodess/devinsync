# devinsync

Real-time file sync for pair programming sessions — no screen sharing, no branches, no commit noise.

---

## What is this?

devinsync lets two developers work on the same codebase at the same time, on different machines. One person becomes the **host** — they point the tool at their project directory and share a room ID. The other person runs as a **guest** — they connect with that ID and get an instant copy of all the host's files, then keep receiving every change as it happens.

It's not a collaborative editor. It's not Git. There's no conflict resolution, no merging, no history. Think of it more like a live mirror: whatever the host saves, deletes, renames, or moves, the guest sees in near real-time inside a local `.devinsync/` folder.

I built this because I wanted something lightweight for pair debugging and code review sessions. Screen sharing adds too much friction, and pushing temporary WIP commits just to share context felt wrong. A simple "here's my current state, follow along" tool was the gap I wanted to fill.

---

## How it works

There are two components:

**notifier** — a stateless WebSocket server that manages rooms and broadcasts events. Hosts create a room, guests join by ID. The notifier never reads file content directly; it just relays whatever the listener sends.

**listener** — a CLI binary that runs on each developer's machine. The host side watches the filesystem for changes and sends structured events to the notifier. The guest side receives those events and applies them locally.

When a guest connects, the notifier immediately requests a full snapshot from the host. The host walks all its watched files and sends their contents in a single batch. After that, incremental events take over.

```
HOST machine                           GUEST machine
  │                                        │
  │  ./devinsync -target=host              │  ./devinsync -target=guest
  │  -path=./myproject                     │  -room=<room-id>
  │                                        │
  │────── file change ──────►  notifier  ──────► apply change ──►  .devinsync/
  │                                        │
  │◄────── SNAPSHOT_CREATE ──  notifier                            (on connect)
  │─────── SNAPSHOT_SYNC ───►  notifier  ─────────────────────►   (all files)
```

Events the host tracks: `FILE_CREATED`, `FILE_WRITED`, `FILE_REMOVED`, `FILE_MOVED`, `FILE_RENAMED`, and the directory equivalents for all of those. The watcher polls every 500ms, and hidden files (dotfiles) are ignored by default.

When a session ends — or the host disconnects — the guest's `.devinsync/` folder is automatically removed.

---

## Caveats

A few things to be upfront about:

- **Sync is one-directional.** Host changes flow to guests, not the other way around. Guests can edit their local copies but nothing propagates back.
- **No conflict resolution.** If the same file is edited on both sides, the host's version wins on the next event.
- **It's not designed for simultaneous edits on the same file.** That path leads to torn reads and overwritten work.
- **Latency depends on the notifier's round-trip time.** The hosted instance on Render may have cold starts.
- **The watcher is poll-based** (500ms interval), not inotify/FSEvents-based, so there's a small inherent delay on each event.
- **Binary files sync fine**, but very large files will bloat the WebSocket messages since content is transmitted inline as JSON.

---

## Requirements

To use the pre-built binaries, you just need a compatible OS:

- macOS (Apple Silicon / ARM64)
- Linux (x86-64)

To build from source or run the notifier locally:

- Go 1.25+
- Docker (optional, for the notifier)

---

## Getting started

### Using the pre-built binaries (quickest path)

Download the binary for your platform from the `listener/` folder:

- `devinsync-mac` — macOS ARM64
- `devinsync-linux` — Linux x86-64

Create a `.env` file next to the binary:

```env
SERVER_URL=wss://devinsync-notifier.onrender.com/ws
```

Make it executable and run:

**Host:**
```bash
chmod +x ./devinsync-mac
./devinsync-mac -target=host -path=/path/to/your/project
```

The tool will print a room ID to your terminal. Share it with the guest.

**Guest:**
```bash
chmod +x ./devinsync-mac
./devinsync-mac -target=guest -room=<room-id>
```

The guest's copy of the project lands in `~/.devinsync/` (or `$HOME/.devinsync/`). Open that folder in your editor and you're synced.

---

### Building from source

```bash
cd listener

# macOS ARM64
make build-mac

# Linux x86-64
make build-linux
```

The binaries land directly in `listener/`.

---

### Running the notifier locally

If you want to run everything locally instead of using the hosted instance:

```bash
# with Go
make run-notifier

# or with Docker
cd notifier
docker build -t devinsync-notifier .
docker run -p 8080:8080 devinsync-notifier
```

Then set your `.env` to:

```env
SERVER_URL=ws://localhost:8080/ws
```

You can also use the root Makefile shortcuts to run host and guest via `go run` during development:

```bash
make run-host path=/path/to/project
make run-guest room=<room-id>
```

---

## Ignoring files

Create a `.disignore` file in the same directory as the binary. One pattern per line. The watcher already ignores dotfiles by default, so you mainly need this for things like `node_modules` or build artifacts:

```
.node_modules/*
.git/*
dist/*
```

---

## Project structure

```
devinsync/
├── listener/           # CLI client — runs on each developer's machine
│   ├── domain/
│   │   ├── entities/   # Connector (WebSocket), SyncManager, Event structs
│   │   ├── types/      # Event types, target types (host/guest)
│   │   └── constants/  # Guest workspace path constant
│   └── services/       # File system event handling, file I/O, server listener
│
└── notifier/           # WebSocket server — the relay hub
    ├── domain/
    │   └── entities/   # Room (broadcast hub), Client, Event
    ├── handlers/        # WebSocket upgrade, host/guest connection handlers
    └── routes/          # Route setup (/ws/host, /ws/join/:id)
```

---

## Tech stack

| Component | Stack |
|-----------|-------|
| listener  | Go, `gorilla/websocket`, `radovskyb/watcher` |
| notifier  | Go, Fiber v3, `gofiber/websocket`, `google/uuid` |
| deploy    | Docker (Alpine), hosted on Render |

---

## Why I built this

This started as a personal challenge. I wanted to understand what it takes to build a real-time sync system from scratch — dealing with filesystem events, WebSocket state, race conditions between the watcher and incoming events, and the snapshot/catch-up problem on new connections.

My first instinct was to write the listener in C. The filesystem watching part is doable — `inotify` on Linux, `kqueue` on macOS, `ReadDirectoryChangesW` on Windows — but each one has a completely different API and behavior. Abstracting over all three in C would have turned the interesting problem into a portability project. I switched to Go and used `radovskyb/watcher` to handle that layer, which let me focus on the actual sync logic instead.

It's a focused project with a specific scope, and I'm happy with how the deduplication strategy (ignoring events we applied ourselves for 2 seconds) ended up solving the feedback loop problem simply. No distributed lock, no CRDT, no overhead — just a timestamp map and a mutex.
