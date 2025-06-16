# LocalCloud

LocalCloud is a lightweight local cloud platform that provides container services similar to popular cloud providers, but runs entirely on your local machine. Learn cloud opperations without risk!

## Features

### Container Management
- Create and manage Docker containers
- Execute commands inside containers
- View real time container logs and metrics
- Monitor container status and uptime

### Web interface
- Easy management of containers
- Real-time updates via WebSocket

### CLI interface
- Full command line support for all Operationsions
- Compatible with Docker workflows

## Installation

1. Ensure you have Go 1.21 or later installed, and Docker running on your machine.
2. Clone the repository:
   ```bash
   git clone https://github.com/mattfrayser/localcloud.git
   cd localcloud
   ```
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Build the project:
   ```bash
   go build -o localcloud ./cmd/localcloud
   ```

## Usage

### Start Web interface
```bash
# Start web interface (default port 8080)
./localcloud web

# Start web interface on a specific port
./localcloud web --port 8081
```
### CLI Commands
```bash
# Create new 
localcloud new

# List containers
localcloud list

# Run commands
localcloud exec --id <ID> --c <COMMAND>

# Delete
localcloud delete --id <ID>
```
