# Calculator REST API

A simple REST API that provides basic calculator operations (add, subtract, multiply, divide) built with Go and the Gin framework.

## Features

- Four basic arithmetic operations: addition, subtraction, multiplication, and division
- RESTful API design
- Comprehensive error handling
- Health check endpoint
- >70% test coverage
- Graceful shutdown handling

## Prerequisites

- Go 1.21 or higher
- Make (optional, for using Makefile commands)

## Installation

1. Clone or download the repository
2. Navigate to the project directory:
   ```bash
   cd calculator-api
   ```
3. Install dependencies:
   ```bash
   go mod download
   ```

## Running the Server

### Using Make
```bash
make run
```

### Using Go directly
```bash
go run cmd/api/main.go
```

The server will start on port 8081.

## Running Tests

### Using Make
```bash
make test
```

### Using Go directly
```bash
go test ./... -v -cover
```

## Building the Application

### Using Make
```bash
make build
```

This creates a binary at `bin/calculator-api`.

### Using Go directly
```bash
go build -o bin/calculator-api cmd/api/main.go
```

## API Endpoints

### Health Check

Check if the API is running.

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "healthy",
  "version": "1.0.0"
}
```

**Example:**
```bash
curl http://localhost:8081/health
```

### Calculate

Perform arithmetic operations.

**Endpoint:** `POST /api/v1/calculate`

**Request Body:**
```json
{
  "operation": "add|subtract|multiply|divide",
  "operand1": <number>,
  "operand2": <number>
}
```

**Success Response (200 OK):**
```json
{
  "result": <number>,
  "operation": "<operation>",
  "operands": [<operand1>, <operand2>]
}
```

**Error Response (400 Bad Request):**
```json
{
  "error": {
    "code": "<ERROR_CODE>",
    "message": "<error message>"
  }
}
```

### Error Codes

- `INVALID_INPUT`: Missing or invalid fields in request body
- `INVALID_OPERATION`: Operation not in [add, subtract, multiply, divide]
- `DIVISION_BY_ZERO`: Attempt to divide by zero

## Usage Examples

### Addition
```bash
curl -X POST http://localhost:8081/api/v1/calculate \
  -H "Content-Type: application/json" \
  -d '{"operation": "add", "operand1": 5, "operand2": 3}'
```

**Response:**
```json
{
  "result": 8,
  "operation": "add",
  "operands": [5, 3]
}
```

### Subtraction
```bash
curl -X POST http://localhost:8081/api/v1/calculate \
  -H "Content-Type: application/json" \
  -d '{"operation": "subtract", "operand1": 10, "operand2": 4}'
```

**Response:**
```json
{
  "result": 6,
  "operation": "subtract",
  "operands": [10, 4]
}
```

### Multiplication
```bash
curl -X POST http://localhost:8081/api/v1/calculate \
  -H "Content-Type: application/json" \
  -d '{"operation": "multiply", "operand1": 6, "operand2": 7}'
```

**Response:**
```json
{
  "result": 42,
  "operation": "multiply",
  "operands": [6, 7]
}
```

### Division
```bash
curl -X POST http://localhost:8081/api/v1/calculate \
  -H "Content-Type: application/json" \
  -d '{"operation": "divide", "operand1": 15, "operand2": 3}'
```

**Response:**
```json
{
  "result": 5,
  "operation": "divide",
  "operands": [15, 3]
}
```

### Error Example: Division by Zero
```bash
curl -X POST http://localhost:8081/api/v1/calculate \
  -H "Content-Type: application/json" \
  -d '{"operation": "divide", "operand1": 10, "operand2": 0}'
```

**Response:**
```json
{
  "error": {
    "code": "DIVISION_BY_ZERO",
    "message": "Cannot divide by zero"
  }
}
```

### Error Example: Invalid Operation
```bash
curl -X POST http://localhost:8081/api/v1/calculate \
  -H "Content-Type: application/json" \
  -d '{"operation": "power", "operand1": 2, "operand2": 3}'
```

**Response:**
```json
{
  "error": {
    "code": "INVALID_OPERATION",
    "message": "Operation must be one of: add, subtract, multiply, divide"
  }
}
```

## Project Structure

```
calculator-api/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── handler/
│   │   ├── calculator.go        # Calculator endpoint handlers
│   │   ├── calculator_test.go   # Handler tests
│   │   └── health.go            # Health check handler
│   └── service/
│       ├── calculator.go        # Calculator business logic
│       └── calculator_test.go   # Service tests
├── go.mod                       # Go module definition
├── go.sum                       # Go dependencies checksums
├── Makefile                     # Build automation
└── README.md                    # This file
```

## Development

### Adding New Operations

1. Add the operation method to `internal/service/calculator.go`
2. Add a case in the switch statement in `internal/handler/calculator.go`
3. Write tests in `internal/service/calculator_test.go` and `internal/handler/calculator_test.go`
4. Run tests to verify: `make test`

## Clean Up

Remove built binaries:
```bash
make clean
```

## License

This project is provided as-is for educational purposes.
