# IM System

A high-performance instant messaging system built with Go + React, supporting single chat, group chat, offline messages, and multi-node routing.

![Auth](https://img.shields.io/badge/Auth-JWT-blue)
![WebSocket](https://img.shields.io/badge/Transport-WebSocket-green)
![Kafka](https://img.shields.io/badge/MQ-Kafka-orange)
![Redis](https://img.shields.io/badge/Cache-Redis-red)
![MySQL](https://img.shields.io/badge/DB-MySQL-blue)

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│              React Frontend (Vite + Tailwind)           │
└──────────────────────┬──────────────────────────────────┘
                       │ WebSocket + HTTP
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
       │         Kafka Async Pipeline      │
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

- **React Web UI** — dark theme chat interface with real-time messaging
- **WebSocket long connection** with heartbeat (Ping/Pong)
- **JWT authentication** on first packet
- **Message reliability** — ACK + sequence number
- **Kafka async pipeline** — decouple send from deliver, 8-worker consumer pool
- **Offline messages** — Redis List, pulled on reconnect
- **Group chat** — write fan-out, Redis Set for members
- **Message persistence** — MySQL with cursor-based history query
- **Cross-node routing** — Redis Pub/Sub, stateless horizontal scaling
- **Rate limiting** — per-user token bucket (10 msg/s)
- **Snowflake ID** — globally unique message IDs

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | React 18 + Vite + Tailwind CSS |
| Language | Go 1.25 |
| WebSocket | gorilla/websocket |
| Message Queue | Apache Kafka 3.7 |
| Auth | JWT (golang-jwt) |
| Cache | Redis 7 |
| Database | MySQL 8 + GORM |
| Logger | Uber Zap |
| Config | Viper |

## Quick Start

### Prerequisites
- Go 1.21+
- Node.js 18+
- Redis
- MySQL
- Kafka

### Backend

```bash
# Copy and edit config
cp config.example.yaml config.yaml

# Start server
go run main.go
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Open `http://localhost:5173`, register an account and start chatting.

### With Docker Compose

```bash
docker-compose up -d
```

### CLI Test Client

```bash
# Terminal 1 — alice
go run cmd/client/main.go -user=alice -pass=123456

# Terminal 2 — bob
go run cmd/client/main.go -user=bob -pass=123456

# Send single message
> s 2 hello bob!

# Send group message
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

First packet must be auth:

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
go run benchmark/main.go -c 100 -d 30s

# 1000 connections, 60s
go run benchmark/main.go -c 1000 -d 60s
```

### Results

| Connections | Duration | Send QPS | Recv QPS | Msg Errors |
|-------------|----------|----------|----------|------------|
| 100         | 30s      | 997      | 1,973    | 0          |
| 500         | 30s      | 2,183    | 2,238    | 0          |
| 1000        | 60s      | 1,069    | 1,677    | 0          |

> recv QPS > send QPS because each message generates 1 ACK + 1 delivery via Kafka async pipeline.

## Project Structure

```
im-system/
├── frontend/           React frontend (Vite + Tailwind)
│   ├── src/
│   │   ├── api/        HTTP + WebSocket client
│   │   ├── components/ Sidebar, ChatWindow
│   │   ├── hooks/      useWebSocket
│   │   ├── pages/      AuthPage, ChatPage
│   │   └── store/      Global state (useReducer)
├── api/                HTTP API handlers
├── benchmark/          Load testing tool
├── cmd/client/         CLI test client
├── gateway/            WebSocket layer (connection, handler)
├── group/              Group chat service
├── message/            Single/group message delivery
├── pkg/
│   ├── config/         Configuration
│   ├── db/             MySQL + GORM models
│   ├── limiter/        Token bucket rate limiter
│   ├── logger/         Zap logger
│   ├── mq/             Kafka producer + consumer
│   ├── redis/          Online status, offline store
│   └── snowflake/      Snowflake ID generator
├── proto/              Message protocol definitions
├── router/             Cross-node routing via Redis Pub/Sub
└── user/               User service, JWT auth
```
