# Program Manager

A comprehensive program management system with web interface for managing and monitoring applications.

## Features

- **Program Management**: Add, edit, delete, and manage programs
- **Process Monitoring**: Real-time monitoring of running processes
- **Log Management**: Centralized logging with search and filtering
- **Web Interface**: Modern React-based web UI with Ant Design
- **Cross-Platform**: Support for Windows and Linux

## Quick Start

### Prerequisites

- **Go 1.19+**: For backend development
- **Node.js 18+**: For frontend development
- **pnpm**: Package manager for frontend dependencies

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd program-manager
   ```

2. **Install dependencies**
   ```bash
   make.bat deps
   ```

3. **Build the project**
   ```bash
   make.bat all
   ```

4. **Run the application**
   ```bash
   program-manager.exe
   ```

### Development

- **Start frontend dev server**: `make.bat dev`
- **Build Go backend**: `make.bat go`
- **Build frontend**: `make.bat web`
- **Clean build files**: `make.bat clean`

### Usage

1. **Start the server**: Run `program-manager.exe`
2. **Access web interface**: Open http://localhost:8080 in your browser
3. **Manage programs**: Use the web interface to add and manage your programs

## Build Commands

The project includes a comprehensive build tool (`make.bat`) with the following commands:

- `make.bat help` - Show help information
- `make.bat deps` - Install frontend dependencies
- `make.bat go` - Build Go backend
- `make.bat web` - Build frontend
- `make.bat dev` - Start frontend development server
- `make.bat all` - Full project build (dependencies + frontend + backend)
- `make.bat clean` - Clean build files

## Automated Build and Release

This project includes GitHub Actions workflows for automated building and releasing:

### 🚀 Quick Release

To create a new release, simply push a version tag:

```bash
git tag v1.0.0
git push origin v1.0.0
```

### 📦 Available Builds

- **Linux AMD64**: `program-manager-linux-amd64`
- **Docker Image**: `ghcr.io/{owner}/{repository}:{tag}`
- **Release Package**: Complete tar.gz package with all files

### 🐳 Docker Usage

```bash
# Pull the latest image
docker pull ghcr.io/{owner}/{repository}:latest

# Run with Docker
docker run -p 8080:8080 -v $(pwd)/data:/app/data ghcr.io/{owner}/{repository}:latest
```

### 🔧 Manual Build

```bash
# Build optimized Linux binary
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o program-manager-linux .

# Build Docker image locally
docker build -t program-manager:latest .
```

## Project Structure

```
program-manager/
├── main.go              # Main application entry point
├── handlers/            # HTTP request handlers
├── models/              # Data models
├── services/            # Business logic services
├── utils/               # Utility functions
├── web/                 # Frontend React application
├── config/              # Configuration files
├── storage/             # Data persistence
└── docs/                # Documentation
```

## API Documentation

- **GET /api/programs** - List all programs
- **POST /api/programs** - Create new program
- **PUT /api/programs/:id** - Update program
- **DELETE /api/programs/:id** - Delete program
- **POST /api/programs/:id/start** - Start program
- **POST /api/programs/:id/stop** - Stop program
- **GET /api/logs** - Get program logs
- **GET /api/status** - Get system status

## Configuration

Configuration can be managed through the web interface or by editing `config/config.go`.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly using `make.bat all`
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.