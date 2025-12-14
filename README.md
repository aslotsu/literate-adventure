# Exobook Notification Worker

A NATS subscriber service that asynchronously processes notification events and stores them in DynamoDB.

## 🎯 Purpose

This service decouples notification creation from user-facing API endpoints, providing:
- **Performance**: API responses are faster (no blocking on notification creation)
- **Reliability**: Notifications are created even if users are offline
- **Security**: Notifications can only be created server-side, not by clients
- **Scalability**: Can scale independently of main API

## 🏗️ Architecture

```
API (dynamodb-go-api) → NATS → Notification Worker → DynamoDB
                         ↓
                    (async, non-blocking)
```

### Event Flow

1. User action triggers API call (e.g., like a post)
2. API creates like in DynamoDB + publishes NATS event
3. API returns immediately (fast response)
4. Notification worker receives NATS event
5. Worker creates notification in DynamoDB
6. User sees notification in their feed

## 📦 Events Handled

The worker subscribes to `notifications.>` (all notification events):

- `notifications.post.like` - User likes a post
- `notifications.post.unlike` - User unlikes a post
- `notifications.comment.like` - User likes a comment
- `notifications.reply.post` - User replies to a post
- `notifications.reply.comment` - User replies to a comment
- More can be added easily...

## 🚀 Getting Started

### Prerequisites

- Go 1.23+
- AWS credentials with DynamoDB access
- NATS credentials file (`NGS-Default-exobook.creds`)

### Installation

```bash
# Clone repository
cd backend/notification-worker

# Install dependencies
go mod download

# Copy environment file
cp .env.example .env

# Edit .env with your credentials
nano .env
```

### Running Locally

```bash
# Run directly
go run main.go

# Or use Makefile
make run

# With auto-reload (requires air)
make dev
```

### Building

```bash
# Build binary
make build

# Run binary
./bin/notification-worker
```

## 🐳 Docker

```bash
# Build image
make docker-build

# Run container
make docker-run

# Or manually
docker build -t notification-worker .
docker run --env-file .env notification-worker
```

## ⚙️ Configuration

All configuration is done via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `NATS_URL` | `nats://connect.ngs.global` | NATS server URL |
| `NATS_CREDS_FILE` | `NGS-Default-exobook.creds` | NATS credentials file |
| `AWS_REGION` | `ca-central-1` | AWS region |
| `AWS_ACCESS_KEY_ID` | - | AWS access key |
| `AWS_SECRET_ACCESS_KEY` | - | AWS secret key |
| `NOTIF_TABLE_NAME` | `exobook-notifications` | DynamoDB table name |
| `ENVIRONMENT` | `development` | Environment (development/production) |
| `LOG_LEVEL` | `info` | Log level |

## 📊 Notification Event Schema

```json
{
  "owner": "user-123",           // User receiving notification
  "trigger_user": "user-456",    // User who triggered action
  "username": "John Doe",        // Trigger user's name
  "user_picture": "https://...", // Trigger user's picture
  "user_bio": "Software dev",    // Trigger user's bio
  "action": 1,                   // 1=like post, 2=like comment, etc.
  "resource_type": "POST",       // POST, COMMENT, etc.
  "resource_id": "post-789",     // ID of the resource
  "excerpt": "Great post!...",   // Optional preview text
  "created_at": 1765318000       // Unix timestamp
}
```

## 🔧 Development

### Project Structure

```
notification-worker/
├── main.go                 # Entry point
├── config/
│   └── config.go          # Configuration management
├── models/
│   ├── event.go           # NATS event models
│   └── notification.go    # DynamoDB notification models
├── handlers/
│   ├── worker.go          # NATS subscriber
│   └── notification_service.go  # DynamoDB operations
├── Dockerfile             # Container image
├── Makefile              # Development commands
└── README.md             # This file
```

### Adding New Event Types

1. Define event topic in `models/event.go`:
```go
const TopicNewEventType = "notifications.new.event"
```

2. Publish from API:
```go
natsConn.Publish("notifications.new.event", eventData)
```

3. Worker automatically handles it! (subscribes to `notifications.>`)

### Testing

```bash
# Run tests
make test

# Test with specific event
go run main.go

# In another terminal, publish test event
nats pub notifications.post.like '{"owner":"user1","trigger_user":"user2",...}'
```

## 📈 Monitoring

The worker logs all events:

```
📨 Received event on subject: notifications.post.like
✅ Created notification: owner=user-123, action=1, resource=post-789
```

Monitor these logs for:
- Event processing time
- Failed events
- Duplicate notifications (skipped)

## 🚨 Error Handling

- **Invalid events**: Logged and skipped
- **DynamoDB errors**: Logged (TODO: add retry logic)
- **NATS disconnection**: Auto-reconnects infinitely
- **Duplicate notifications**: Detected and skipped using `action_key`

## 🔐 Security

- Notifications can only be created by this service (not by clients)
- NATS requires authentication via credentials file
- AWS credentials required for DynamoDB access
- Validates all event fields before processing

## 🚀 Deployment

### Railway

```bash
# Push to GitHub
git push origin main

# Railway will auto-deploy from Dockerfile
```

### Environment Variables in Production

Set in Railway dashboard:
- `NATS_URL`
- `NATS_CREDS_FILE` (upload file separately)
- `AWS_REGION`
- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`
- `NOTIF_TABLE_NAME`
- `ENVIRONMENT=production`

## 📝 TODO

- [ ] Add retry logic for failed DynamoDB writes
- [ ] Implement dead letter queue for failed events
- [ ] Add metrics/observability (Prometheus)
- [ ] Add health check endpoint
- [ ] Implement rate limiting per user
- [ ] Add support for notification batching
- [ ] Add support for email/push notifications

## 🤝 Contributing

1. Create feature branch
2. Make changes
3. Run tests: `make test`
4. Format code: `make fmt`
5. Submit PR

## 📄 License

MIT License - See LICENSE file

## 💬 Support

For issues or questions:
- Check logs for error messages
- Verify NATS connection
- Verify AWS credentials
- Check DynamoDB table exists

---

**Built with ❤️ for Exobook**
