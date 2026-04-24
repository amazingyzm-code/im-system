# IM System

A high-performance instant messaging backend built with Go, supporting single chat, group chat, offline messages, and multi-node routing.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                        Clients                          │
└──────────────────────┬──────────────────────────────────┘
                       │ WebSocket
┌──────────────────────▼──────────────────────────────────┐
│                    Gateway Layer                        │
│         WebSocket upgrade · JWT auth · Heartbeat        │
└──────┬───────────────────────────────────┬──────────────┘
       │                                   │
┌──────▼──────┐                   ┌────────▼────────┐
│   Message   │                   │      Group      │
│   Service   │                   │     Service     │
│  single chat│                   │  fan-out write  │
└──────┬──────┘                   └────────┬────────┘
       │                                   │
┌──────▼───────────────────────────────────▼──────────────┐
│                     Storage Layer                       │
│   Redis (online status · offline msgs · group members)  │
│   MySQL (message history · users · group members)       │
└─────────────────────────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│               Cross-node Router                         │
│              Redis Pub/Sub per node                     │
└─────────────────────────────────────────────────────────┘
```

## Features

- **WebSocket long connection** with heartbeat (Ping/Pong)
- **JWT authentication** on first packet
- **Message reliability** — ACK + sequence number
- **Offline messages** — Redis List, pulled on reconnect
- **Group chat** — write fan-out, Redis Set for members
- **Message persistence** — MySQL with cursor-based history query
- **Cross-node routing** — Redis Pub/Sub, stateless horizontal scaling
- **Rate limiting** — per-user token bucket (10 msg/s)
- **Snowflake ID** — globally unique message IDs

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.25 |
| WebSocket | gorilla/websocket |
| Auth | JWT (golang-jwt) |
| Cache | Redis 7 |
| Database | MySQL 8 + GORM |
| Logger | Uber Zap |
| Config | Viper |

## Quick Start

### With Docker Compose (recommended)

```bash
docker-compose up -d
```

### Local

```bash
# Start Redis and MySQL first, then:
go run main.go
```

### Test with CLI client

```bash
# Terminal 1 — user alice
go run cmd/client/main.go -user=alice -pass=123456

# Terminal 2 — user bob
go run cmd/client/main.go -user=bob -pass=123456

# In alice's terminal, send to bob (uid=2)
> s 2 hello bob!

# Send group message to group 1
> g 1 hello group!
```

## API

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/user/register` | Register |
| POST | `/api/user/login` | Login, returns JWT |
| GET | `/ws` | WebSocket connection |
| GET | `/api/history/single` | Single chat history |
| GET | `/api/history/group` | Group chat history |
| POST | `/api/group/join` | Join group |
| POST | `/api/group/leave` | Leave group |
| GET | `/health` | Health check |

## WebSocket Protocol

All messages are JSON. First packet must be auth:

```json
{"msg_type": 5, "token": "<jwt>"}
```

Send a message:

```json
{
  "msg_type": 1,
  "target_type": 1,
  "to_id": 2,
  "content": "hello",
  "timestamp": 1700000000000,
  "seq": 1
}
```

Server ACK:

```json
{"msg_type": 3, "seq": 1, "msg_id": 7891234567890}
```

## Benchmark

```bash
# 100 connections, 30s
go run benchmark/main.go -c 100 -d 30s -i 50ms

# 1000 connections, 60s
go run benchmark/main.go -c 1000 -d 60s -i 100ms
```

### Results

| Connections | Duration | Send QPS | Recv QPS | Msg Errors |
|-------------|----------|----------|----------|------------|
| 100         | 30s      | 997      | 1,973    | 0          |
| 500         | 30s      | 2,183    | 2,238    | 281        |
| 1000        | 60s      | 1,069    | 1,677    | 0          |

> recv QPS > send QPS because each message generates 1 ACK + 1 delivery via Kafka async pipeline.

## Project Structure

```
im-system/
├── api/            HTTP API handlers
├── benchmark/      Load testing tool
├── cmd/client/     CLI test client
├── gateway/        WebSocket layer (connection, handler)
├── group/          Group chat service
├── message/        Single/group message delivery
├── pkg/
│   ├── config/     Configuration
│   ├── db/         MySQL + GORM models
│   ├── limiter/    Token bucket rate limiter
│   ├── logger/     Zap logger
│   ├── redis/      Redis client, online status, offline store
│   └── snowflake/  Snowflake ID generator
├── proto/          Message protocol definitions
├── router/         Cross-node routing via Redis Pub/Sub
└── user/           User service, JWT auth
```
