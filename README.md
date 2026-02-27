# Casheer Report Service

Report and Finance management service for Casheer POS system.

## Features

- 📊 Daily, monthly, and yearly reports
- 💰 Revenue and expense tracking
- 📝 Expense management with categories
- 🖨️ Receipt template management
- 📄 Receipt printing with ESC/POS support
- 📡 Bluetooth printer support
- 🔌 RabbitMQ integration for real-time updates
- 📈 Profit & loss calculation

## Tech Stack

- Go 1.21
- Fiber v2
- GORM
- PostgreSQL
- RabbitMQ
- Bluetooth (go-bluetooth)
- ESC/POS protocol

## API Endpoints

### Report Endpoints
- `GET /api/v1/reports/daily` - Get daily report
- `GET /api/v1/reports/monthly` - Get monthly report
- `GET /api/v1/reports/yearly` - Get yearly report
- `GET /api/v1/reports/revenue` - Get revenue summary
- `GET /api/v1/reports/expenses` - Get expense summary

### Expense Endpoints
- `POST /api/v1/expenses` - Create expense
- `GET /api/v1/expenses` - Get all expenses
- `GET /api/v1/expenses/:id` - Get expense by ID
- `PUT /api/v1/expenses/:id` - Update expense
- `DELETE /api/v1/expenses/:id` - Delete expense
- `GET /api/v1/expenses/categories/summary` - Get expenses by category

### Template Endpoints
- `POST /api/v1/templates` - Create template
- `GET /api/v1/templates` - Get all templates
- `GET /api/v1/templates/:id` - Get template by ID
- `PUT /api/v1/templates/:id` - Update template
- `DELETE /api/v1/templates/:id` - Delete template
- `PATCH /api/v1/templates/:id/default` - Set as default

### Print Endpoints
- `POST /api/v1/print/receipt/:orderId` - Print receipt
- `POST /api/v1/print/test` - Print test page

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| DB_HOST | Database host | localhost |
| DB_PORT | Database port | 5432 |
| DB_USER | Database user | postgres |
| DB_PASSWORD | Database password | postgres |
| DB_NAME | Database name | casheer_db |
| PORT | Service port | 3003 |
| JWT_SECRET | JWT secret key | - |
| RABBITMQ_URL | RabbitMQ URL | amqp://localhost:5672 |
| PRINTER_DEFAULT_WIDTH | Default paper width | 58mm |

## Bluetooth Printer Support

Supports ESC/POS thermal printers via Bluetooth:
- 58mm and 80mm paper widths
- Logo printing
- QR Code printing
- Custom templates

## Running the Service

```bash
go mod tidy
go run cmd/main.go
```

### **22. `Tiltfile`**
```python
# Tiltfile for Casheer Report Service

print("🚀 Starting Casheer Report Service...")

# Build configuration
docker_build(
    'casheer-report-service',
    '.',
    dockerfile='Dockerfile',
    build_args={
        'GO_VERSION': '1.21'
    }
)

# Kubernetes deployment
k8s_yaml('k8s/deployment.yaml')

# Port forwarding for local development
k8s_resource(
    'casheer-report-service',
    port_forwards='3003:3003',
    labels=['report-service']
)

# Watch for changes
watch_file('internal/**/*.go')
watch_file('cmd/**/*.go')
watch_file('pkg/**/*.go')
watch_file('.env')

print("✅ Report Service configuration loaded")