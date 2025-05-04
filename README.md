# Omnichannel Inventory Backend (Golang)

A real-time multi-warehouse inventory synchronization system for omnichannel retail. Supports stock management across multiple warehouses and sales channels (Amazon, Flipkart, POS), with Redis Streams for event-driven updates and PostgreSQL for durability.

## Features

- Add/update product stock in specific warehouses
- Real-time consolidated stock per product
- Simulate order events from any channel
- Inventory change history log
- RESTful APIs (Gin)
- Redis Streams for real-time events
- PostgreSQL for persistence
- Webhook notifications for low stock
- OpenAPI (Swagger) docs
- Dockerized for easy deployment

## Prerequisites

- Docker & Docker Compose
- Git
- Go 1.24 or later (for development)

## Installation

1. Clone the repository:

```bash
git clone <repository-url>
cd omnichannel_inventory
```

2. Create a `.env` file in the project root (copy from `configs/config.sample.env`):

```bash
cp configs/config.sample.env .env
```

3. Configure environment variables in `.env`:

```env
# PostgreSQL
DB_HOST=db
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=inventory

# Redis
REDIS_ADDR=redis:6379
REDIS_PASSWORD=
REDIS_DB=0

# App
APP_PORT=8081

# Webhook (optional)
WEBHOOK_URL=http://your-webhook-url
```

## Running the Application

1. Start the application using Docker Compose:

```bash
docker-compose up --build
```

This will start:

- PostgreSQL database on port 5432
- Redis server on port 6379
- Application server on port 8081

2. Access the application:

- API Documentation: http://localhost:8081/swagger/index.html
- Web Interface: http://localhost:8081

## API Endpoints

### Stock Management

- `POST /api/stock` - Add or update stock

  ```json
  {
    "sku": "PROD001",
    "warehouse_id": 1,
    "quantity": 100
  }
  ```

- `GET /api/stock/:sku` - Get consolidated stock for a product

### Order Simulation

- `POST /api/orders/simulate` - Simulate an order
  ```json
  {
    "sku": "PROD001",
    "channel": "amazon",
    "quantity": 5
  }
  ```

### History

- `GET /api/history/:sku` - Get inventory history for a product

## Features in Detail

### Webhook Notifications

The system sends webhook notifications when stock levels fall below the threshold (10 units). To enable this:

1. Set the `WEBHOOK_URL` in your `.env` file
2. The webhook will receive POST requests with the following payload:

```json
{
  "sku": "PROD001",
  "warehouse_id": 1,
  "stock": 5,
  "timestamp": "2024-05-04T01:20:12Z"
}
```

### Inventory History Log

The system maintains a complete history of all inventory changes in the `inventory_transactions` table. Each transaction includes:

- SKU
- Warehouse ID
- Quantity change
- Transaction type (stock_update or order)
- Sales channel (for orders)
- Timestamp

## Development

1. Install Go dependencies:

```bash
go mod download
```

2. Run tests:

```bash
go test ./...
```

3. Build the application:

```bash
go build -o omnichannel_inventory cmd/main.go
```

## Project Structure

- `cmd/` - Main application entry point
- `internal/`
  - `handlers/` - HTTP request handlers
  - `services/` - Business logic
  - `models/` - Data structures
  - `db/` - Database connections
  - `events/` - Redis Streams event handling
  - `webhooks/` - Webhook notifications
- `configs/` - Configuration files
- `scripts/` - Database initialization scripts
- `static/` - Static web files
- `docs/` - API documentation

## Troubleshooting

1. Port Conflicts:

   - If port 8081 is in use, change `APP_PORT` in `.env`
   - If port 5432 is in use, change PostgreSQL port in `docker-compose.yml`

2. Database Issues:

   - Check PostgreSQL logs: `docker-compose logs db`
   - Verify database initialization: `docker-compose exec db psql -U postgres -d inventory`

3. Redis Issues:
   - Check Redis logs: `docker-compose logs redis`
   - Test Redis connection: `docker-compose exec redis redis-cli ping`

## License

MIT
