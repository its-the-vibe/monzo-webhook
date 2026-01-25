# monzo-webhook

A simple web service which consumes Monzo Bank API events and publishes to Redis pub/sub

## Features

- Receives and parses Monzo webhook POST requests
- Event filtering with configuration file support
- Publishes webhook payloads to event-specific Redis pub/sub channels
- Configurable log levels (DEBUG, INFO, WARN, ERROR)
- Configurable port via environment variable
- Configurable Redis connection via environment variables
- Docker and Docker Compose support for easy deployment

## Configuration

### Event Configuration

The webhook server uses a JSON configuration file to specify the Redis pub/sub channel where all webhook events will be published. All event types from Monzo are accepted and published to this channel.

**Configuration File Format:**

Create a `config.json` file (or specify a custom path via `CONFIG_FILE` environment variable):

```json
{
  "channel": "monzo-webhook"
}
```

**Environment Variables:**

- `CONFIG_FILE`: Path to the configuration file (default: `config.json`)

The server will examine the `type` field in the incoming webhook payload for logging purposes and publish all events to the configured Redis channel.

**Example:**

```bash
# Use default config.json
./webhook-server

# Use custom configuration file
CONFIG_FILE=/path/to/my-config.json ./webhook-server
```

### Log Level Configuration

Control the verbosity of logging with the `LOG_LEVEL` environment variable.

**Available Log Levels:**

- `DEBUG`: Most verbose, includes webhook payloads
- `INFO`: Standard operational messages (default)
- `WARN`: Warning messages only
- `ERROR`: Error messages only

**Environment Variables:**

- `LOG_LEVEL`: Sets the logging level (default: `INFO`)

**Note:** Webhook payloads are only logged when `LOG_LEVEL` is set to `DEBUG`. This prevents sensitive data from appearing in logs during normal operation.

**Example:**

```bash
# Use INFO level (default)
./webhook-server

# Use DEBUG level to see webhook payloads
LOG_LEVEL=DEBUG ./webhook-server

# Use WARN level for minimal logging
LOG_LEVEL=WARN ./webhook-server
```

### Port Configuration

The server port can be configured via the `PORT` environment variable. If not set, it defaults to `8080`.

```bash
# Run on default port 8080
./webhook-server

# Run on custom port
PORT=3000 ./webhook-server
```

### Redis Configuration

The webhook service publishes all received webhooks to a single Redis pub/sub channel specified in the configuration file.

**Environment Variables:**

- `REDIS_HOST`: Redis server hostname (default: `localhost`)
- `REDIS_PORT`: Redis server port (default: `6379`)
- `REDIS_PASSWORD`: Redis server password (optional; default: unset)

**Note:** If the Redis connection fails, the application will log a warning and continue to work without Redis publishing. This ensures the webhook service remains operational even if Redis is unavailable.

```bash
# Run with Redis configuration
REDIS_HOST=redis.example.com REDIS_PORT=6379 ./webhook-server

# Run with Redis password
REDIS_HOST=redis.example.com REDIS_PORT=6379 REDIS_PASSWORD=yourpassword ./webhook-server

# Run with default Redis settings (connects to localhost:6379, no password)
./webhook-server
```

## Building and Running

### Local Development

```bash
# Build the application
go build -o webhook-server

# Run the server (requires config.json)
./webhook-server

# Run with custom configuration file
CONFIG_FILE=my-config.json ./webhook-server

# Run with custom port
PORT=3000 ./webhook-server

# Run with DEBUG logging
LOG_LEVEL=DEBUG ./webhook-server

# Run with Redis configuration
REDIS_HOST=redis.example.com REDIS_PORT=6379 ./webhook-server

# Run with all options
LOG_LEVEL=DEBUG CONFIG_FILE=config.json REDIS_HOST=localhost PORT=8080 ./webhook-server
```

### Using Docker

```bash
# Build the Docker image
docker build -t monzo-webhook .

# Run the container (mount config.json)
docker run -p 8080:8080 -v $(pwd)/config.json:/app/config.json:ro monzo-webhook

# Run with custom log level
docker run -p 8080:8080 -e LOG_LEVEL=DEBUG -v $(pwd)/config.json:/app/config.json:ro monzo-webhook

# Run with custom port
docker run -p 3000:8080 -e PORT=8080 -v $(pwd)/config.json:/app/config.json:ro monzo-webhook

# Run with Redis configuration (connecting to Redis on host machine)
docker run -p 8080:8080 -e REDIS_HOST=host.docker.internal -e REDIS_PORT=6379 -v $(pwd)/config.json:/app/config.json:ro monzo-webhook

# Run with Redis password
docker run -p 8080:8080 -e REDIS_HOST=host.docker.internal -e REDIS_PORT=6379 -e REDIS_PASSWORD=yourpassword -v $(pwd)/config.json:/app/config.json:ro monzo-webhook
```

### Using Docker Compose

The easiest way to run the application:

```bash
# Start the service
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the service
docker-compose down
```

To use custom configuration with Docker Compose, you can set environment variables:

```bash
# Custom port
PORT=3000 docker-compose up -d

# Custom log level
LOG_LEVEL=DEBUG docker-compose up -d

# Custom configuration file
CONFIG_FILE=/path/to/config.json docker-compose up -d

# Redis configuration
REDIS_HOST=192.168.1.100 REDIS_PORT=6379 docker-compose up -d

# Redis password
REDIS_PASSWORD=yourpassword docker-compose up -d
```

The docker-compose configuration automatically mounts the `config.json` file if it exists.

## API Endpoints

### POST /webhook

Accepts Monzo webhook notifications. All event types are accepted and published to the Redis channel specified in the configuration file.

**Request Body Example:**
```json
{
  "type": "transaction.created",
  "data": {
    "id": "tx_00009LyMQT7N7VJi7SaFCN",
    "created": "2015-09-04T14:28:40Z",
    "description": "Coffee at Starbucks",
    "amount": -350,
    "currency": "GBP",
    "merchant": {
      "id": "merch_00008zIcpbAKe8shBxXUtl",
      "name": "Starbucks"
    }
  }
}
```

**Response:**
- `200 OK`: Webhook received and processed successfully
- `405 Method Not Allowed`: Non-POST request
- `400 Bad Request`: Invalid JSON or request body error

## Testing

### Manual Testing with curl

```bash
# Test with a transaction.created event type
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "type": "transaction.created",
    "data": {
      "id": "tx_test123",
      "created": "2026-01-24T12:00:00Z",
      "description": "Test transaction",
      "amount": -100,
      "currency": "GBP"
    }
  }'

# Test with a different event type - all event types are now accepted
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{"type": "account.balance_updated", "data": {}}'
```

### Testing with Redis

If you want to verify that messages are being published to Redis:

```bash
# In one terminal, subscribe to the Redis channel
redis-cli
> SUBSCRIBE monzo-webhook

# In another terminal, send a test webhook
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{"type": "transaction.created", "data": {"id": "test"}}'
```

## Monzo Webhook Setup

To receive webhooks from Monzo:

1. Register a webhook with Monzo API using your access token and account ID
2. Provide your server's `/webhook` endpoint URL
3. Monzo will send POST requests to this endpoint when transactions occur

For more information, see the [Monzo API documentation](https://docs.monzo.com/#webhooks).

## Development

The project follows standard Go conventions:
- Use `gofmt` for code formatting
- Explicit error handling
- Standard library packages preferred

## License

MIT
