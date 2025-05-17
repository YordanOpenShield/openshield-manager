# OpenShield Manager

OpenShield Manager is a Go-based backend for managing agents, jobs, and scripts in a distributed security automation environment. It uses gRPC for agent communication and Gin for its REST API.

## Features
- Agent registration, heartbeat, and state tracking
- Job and task assignment to agents
- Script synchronization and management
- PostgreSQL database integration
- Docker Compose support for development

## Getting Started

### Prerequisites
- Go 1.20+
- Docker & Docker Compose
- protoc (Protocol Buffers compiler)

### Development Setup

1. **Clone the repository:**
   ```cmd
   git clone <your-repo-url>
   cd openshield-manager
   ```

2. **Generate gRPC code:**
   ```cmd
   protoc --go_out=./proto --go-grpc_out=./proto proto/rpc.proto
   ```

3. **Start services with Docker Compose:**
   ```cmd
   cd docker-compose
   docker-compose -f docker-compose-dev.yml up --build
   ```
   The manager will be available at `http://localhost:9000` and PostgreSQL at `localhost:5432`.

4. **Environment Variables:**
   Create a `.env` file in the project root:
   ```env
   DB_HOST=localhost
   DB_USER=user
   DB_PASSWORD=pass
   DB_NAME=openshield
   DB_PORT=5432
   ```

## API Endpoints

- `POST /api/agents/register` — Register a new agent
- `POST /api/agents/unregister` — Unregister an agent
- `POST /api/agents/heartbeat` — Agent heartbeat
- `GET /api/agents/:agent_id/tasks` — List tasks for an agent
- `GET /api/tasks` — List all tasks
- `POST /api/tasks/assign` — Assign a task to an agent
- `GET /api/jobs/available` — List available jobs
- `POST /api/jobs/create` — Create a new job

## gRPC

- Service definitions are in `proto/rpc.proto`.
- To regenerate Go code after editing proto files:
  ```cmd
  protoc --go_out=./proto --go-grpc_out=./proto proto/rpc.proto
  ```

## License

MIT License