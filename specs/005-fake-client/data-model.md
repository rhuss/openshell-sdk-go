# Data Model: Fake Client Package

## Entities

### ObjectStore[T]

Generic thread-safe in-memory store for named objects.

| Field | Type | Description |
|-------|------|-------------|
| mu | sync.RWMutex | Protects concurrent access to objects map |
| objects | map[string]T | Objects keyed by name |
| nameFunc | func(T) string | Extracts the name from an object |
| copyFunc | func(T) T | Deep-copies an object for boundary isolation |

**Operations**: Create, Get, List, Delete, Update (all return deep copies)

### WatchBroadcaster[T]

Distributes typed events to registered watchers.

| Field | Type | Description |
|-------|------|-------------|
| mu | sync.RWMutex | Protects the watchers slice |
| watchers | []*watcher[T] | Active watchers receiving events |

### watcher[T]

Single registered watch consumer.

| Field | Type | Description |
|-------|------|-------------|
| ch | chan Event[T] | Buffered event channel (capacity 100) |
| name | string | Filter: empty = all events, non-empty = filter by name |
| stopped | atomic.Bool | Whether this watcher has been stopped |

### FakeClient

Top-level fake implementing ClientInterface.

| Field | Type | Description |
|-------|------|-------------|
| sandboxes | *fakeSandboxClient | Sandbox sub-client |
| providers | *fakeProviderClient | Provider sub-client |
| health | *fakeHealthClient | Health sub-client |
| exec | *fakeExecClient | Exec stub (returns Unimplemented) |
| files | *fakeFileClient | File stub (returns Unimplemented) |
| closed | atomic.Bool | Whether Close has been called |
| healthResult | *HealthResult | Configurable health response |

### fakeSandboxClient

Implements SandboxInterface.

| Field | Type | Description |
|-------|------|-------------|
| store | *ObjectStore[*Sandbox] | Sandbox object store |
| broadcaster | *WatchBroadcaster[*Sandbox] | Event broadcaster |
| associations | map[string][]string | sandbox name → attached provider names |
| mu | sync.RWMutex | Protects associations map |
| closed | *atomic.Bool | Reference to FakeClient.closed |

### fakeProviderClient

Implements ProviderInterface.

| Field | Type | Description |
|-------|------|-------------|
| store | *ObjectStore[*Provider] | Provider object store |
| closed | *atomic.Bool | Reference to FakeClient.closed |

### fakeHealthClient

Implements HealthInterface.

| Field | Type | Description |
|-------|------|-------------|
| result | *HealthResult | Configurable result |
| closed | *atomic.Bool | Reference to FakeClient.closed |

## State Transitions

### Sandbox Lifecycle in Fake

```
(not exists) --Create--> Provisioning --WaitReady--> Ready --Delete--> (not exists)
```

- Create: inserts with Phase=Provisioning, broadcasts ADDED
- WaitReady: updates Phase to Ready in store, broadcasts MODIFIED, returns sandbox
- Delete: removes from store, broadcasts DELETED

### FakeClient Lifecycle

```
Open --Close--> Closed
```

- Close: sets closed=true, stops all watchers
- Any method after Close: returns Unavailable StatusError

## Relationships

```
FakeClient
  ├── fakeSandboxClient (owns ObjectStore[*Sandbox], WatchBroadcaster[*Sandbox])
  ├── fakeProviderClient (owns ObjectStore[*Provider])
  ├── fakeHealthClient (configurable result)
  ├── fakeExecClient (stub)
  └── fakeFileClient (stub)
```

ObjectStore and WatchBroadcaster are internal to the fake package; they are not exported.
