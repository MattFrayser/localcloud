# LocalCloud

LocalCloud is a lightweight local cloud platform that provides container services similar to popular cloud providers, but runs entirely on your local machine. Learn cloud opperations without risk!

## Features

### Container Management
- Create and manage Docker containers
- Execute commands inside containers
- View container logs and statistics


## Installation

1. Ensure you have Go 1.21 or later installed
2. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/localcloud.git
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

### Container Operations

```bash
# Start the web interface
localcloud web 
# Create new 
localcloud new
# List containers
localcloud list
# Run commands
localcloud exec --id <ID> --c <COMMAND>
# Delete
localcloud delete --id <ID>
```

### Web View

```bash
# Start the web interface
localcloud web 
```

Then open http://localhost:8080 in your browser to view active container:
- IDs
- Names
- Images
- Up time
- Logs


## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 