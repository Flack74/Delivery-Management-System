# Architecture Overview

## System Design

The Delivery Management System follows a clean architecture pattern with clear separation of concerns:

```
├── cmd/                    # Application entry point
├── internal/
│   ├── handlers/          # HTTP request handlers
│   ├── services/          # Business logic layer
│   ├── models/            # Data models and DTOs
│   ├── middleware/        # HTTP middleware (auth, CSRF, etc.)
│   ├── db/               # Database connection and migrations
│   ├── cache/            # Redis cache implementation
│   ├── config/           # Configuration management
│   ├── routes/           # Route definitions
│   └── utils/            # Utility functions
└── tests/                # Test suites
```

## Key Design Decisions

### 1. Concurrency Model
- **Worker Pool Pattern**: 10 goroutines process orders concurrently
- **Channel-based Communication**: Orders queued via buffered channels
- **Atomic Operations**: Thread-safe status updates using sync/atomic

### 2. Data Storage
- **PostgreSQL**: Persistent storage for users and orders
- **Redis**: Real-time pub/sub for order tracking and caching
- **GORM**: ORM for database operations with connection pooling

### 3. Security
- **JWT Authentication**: Stateless token-based authentication
- **CSRF Protection**: Configurable cross-site request forgery protection
- **Input Validation**: Comprehensive sanitization and validation
- **Rate Limiting**: Token bucket algorithm for API protection

### 4. Real-time Features
- **Order Tracking**: Redis pub/sub channels for live updates
- **Status Progression**: Automatic state transitions every 60 seconds
- **Event Publishing**: Order status changes broadcast to subscribers

## Status Flow

```
Created → Dispatched → In Transit → Delivered
   ↓         ↓
Cancelled  Cancelled
```

## API Endpoints

### Authentication
- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User login
- `POST /api/auth/refresh` - Token refresh

### Orders
- `POST /api/orders` - Create order
- `GET /api/orders` - List orders
- `GET /api/orders/:id` - Get order details
- `PUT /api/orders/:id/cancel` - Cancel order

### Admin
- `GET /api/admin/orders` - List all orders
- `POST /api/admin/orders/:id/status` - Update order status

### System
- `GET /health` - Health check
- `GET /metrics` - System metrics
- `GET /csrf-token` - Get CSRF token