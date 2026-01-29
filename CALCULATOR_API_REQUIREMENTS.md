# Calculator API - Project Requirements

## Project Overview
Build a simple REST API that provides basic calculator operations.

## Functional Requirements

### Calculator Operations
The API should support the following operations:
1. **Addition**: Add two numbers
2. **Subtraction**: Subtract two numbers
3. **Multiplication**: Multiply two numbers
4. **Division**: Divide two numbers (with zero-division handling)

### API Endpoints

#### POST /api/v1/calculate
Perform a calculation operation.

**Request Body**:
```json
{
  "operation": "add|subtract|multiply|divide",
  "operand1": number,
  "operand2": number
}
```

**Success Response** (200 OK):
```json
{
  "result": number,
  "operation": "string",
  "operands": [number, number]
}
```

**Error Response** (400 Bad Request):
```json
{
  "error": {
    "code": "INVALID_OPERATION|DIVISION_BY_ZERO|INVALID_INPUT",
    "message": "Error description"
  }
}
```

#### GET /health
Health check endpoint.

**Success Response** (200 OK):
```json
{
  "status": "healthy",
  "version": "1.0.0"
}
```

## Technical Requirements

### Technology Stack
- **Language**: Go
- **Framework**: Gin (lightweight web framework)
- **Port**: 8081 (to avoid conflict with user-management-api on 8080)

### Project Structure
```
calculator-api/
├── cmd/
│   └── api/
│       └── main.go          # Application entry point
├── internal/
│   ├── handler/
│   │   ├── calculator.go    # Calculator handlers
│   │   └── health.go        # Health check handler
│   └── service/
│       └── calculator.go    # Calculator business logic
├── go.mod
├── go.sum
├── Makefile                 # Build and run commands
└── README.md               # Setup and usage instructions
```

### Error Handling
- Invalid operations (not add/subtract/multiply/divide)
- Division by zero
- Invalid input types (non-numeric values)
- Missing required fields

### Response Format
- All responses in JSON format
- Consistent error structure
- HTTP status codes:
  - 200: Success
  - 400: Bad Request (validation errors)
  - 500: Internal Server Error

## Non-Functional Requirements

### Simplicity
- Clean, readable code
- Minimal dependencies
- Easy to understand and extend

### Testing
- Unit tests for calculator service
- Basic handler tests
- Test coverage >70%

### Documentation
- README with setup instructions
- API usage examples with curl
- Code comments for complex logic

## Deliverables

1. **Source Code**:
   - Complete working calculator API
   - Proper Go project structure
   - Clean, idiomatic Go code

2. **Tests**:
   - Unit tests for calculator logic
   - Tests pass with `go test ./...`

3. **Documentation**:
   - README.md with setup and usage
   - Example API calls

4. **Build Scripts**:
   - Makefile with targets:
     - `make run`: Run the server
     - `make test`: Run tests
     - `make build`: Build binary

## Success Criteria
- [ ] Server starts on port 8081
- [ ] GET /health returns 200 OK
- [ ] All four calculator operations work correctly
- [ ] Division by zero returns proper error
- [ ] All tests pass
- [ ] Code is clean and well-structured

## Timeline
- **Target Completion**: 2-3 hours
- **Priority**: Normal

## Example Usage

```bash
# Start the server
make run

# Health check
curl http://localhost:8081/health

# Addition
curl -X POST http://localhost:8081/api/v1/calculate \
  -H "Content-Type: application/json" \
  -d '{"operation": "add", "operand1": 5, "operand2": 3}'

# Division
curl -X POST http://localhost:8081/api/v1/calculate \
  -H "Content-Type: application/json" \
  -d '{"operation": "divide", "operand1": 10, "operand2": 2}'

# Division by zero (should return error)
curl -X POST http://localhost:8081/api/v1/calculate \
  -H "Content-Type: application/json" \
  -d '{"operation": "divide", "operand1": 10, "operand2": 0}'
```

---

**Document Version**: 1.0
**Created**: 2026-01-29
**Engineering Manager**: locke (engineering-manager-1769715322358)
**Status**: Ready for Implementation
