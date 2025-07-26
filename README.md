# Multi-Project Workspace

This workspace contains both client and server applications for a full-stack project.

## Project Structure

```
local/
├── client/          # Frontend application
├── server/          # Go backend server
├── README.md        # This file
└── .gitignore       # Git ignore patterns
```

## Getting Started

### Prerequisites

- Go 1.19+ for the server
- Node.js 16+ for the client
- Git

### Development

1. **Server Development**
   ```bash
   cd server
   go run main.go
   ```

2. **Client Development**
   ```bash
   cd client
   npm start
   ```

3. **Run Both Projects**
   ```bash
   # From the root directory
   ./scripts/dev.sh
   ```

## Server (Go)

The server is built with Go and provides the backend API.

- **Port**: 8080 (default)
- **API Base**: `http://localhost:8080/api`

## Client

The client is a frontend application that consumes the server API.

- **Port**: 3000 (default)
- **Dev Server**: `http://localhost:3000`

## Development Workflow

1. Start the server in one terminal
2. Start the client in another terminal
3. Both will hot-reload on file changes

## Contributing

1. Create a feature branch
2. Make your changes
3. Test both client and server
4. Create a pull request 