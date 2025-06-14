package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) handleDashboard(c *gin.Context) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>LocalCloud Dashboard</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <style>
        .status-running { color: #10b981; }
        .status-exited { color: #ef4444; }
        .live-dot { animation: pulse 2s infinite; }
    </style>
</head>
<body class="bg-gray-50 min-h-screen">
    <div class="container mx-auto px-4 py-8">
        <div class="flex justify-between items-center mb-8">
            <h1 class="text-3xl font-bold text-gray-900">LocalCloud Dashboard</h1>
            <div class="flex items-center space-x-2">
                <div class="live-dot w-3 h-3 bg-green-500 rounded-full"></div>
                <span class="text-sm text-gray-600">Live</span>
            </div>
        </div>

        <!-- Create Container Form -->
        <div class="bg-white rounded-lg shadow mb-6 p-6">
            <h2 class="text-xl font-semibold mb-4">Create New Container</h2>
            <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
                <input id="imageInput" type="text" placeholder="Image (e.g., nginx:latest)" 
                       class="border rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500">
                <input id="nameInput" type="text" placeholder="Name (optional)" 
                       class="border rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500">
                <input id="portsInput" type="text" placeholder="Ports (e.g., 80:80)" 
                       class="border rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500">
                <button onclick="createContainer()" 
                        class="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700">
                    Create
                </button>
            </div>
        </div>

        <!-- Containers Table -->
        <div class="bg-white rounded-lg shadow overflow-hidden">
            <div class="px-6 py-4 border-b">
                <h2 class="text-xl font-semibold">Containers</h2>
            </div>
            <div class="overflow-x-auto">
                <table class="min-w-full divide-y divide-gray-200">
                    <thead class="bg-gray-50">
                        <tr>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ID</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Image</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Ports</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Uptime</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                        </tr>
                    </thead>
                    <tbody id="containerTable" class="bg-white divide-y divide-gray-200">
                    </tbody>
                </table>
            </div>
        </div>
    </div>

    <!-- Logs Modal -->
    <div id="logsModal" class="fixed inset-0 bg-black bg-opacity-50 hidden items-center justify-center z-50">
        <div class="bg-white rounded-lg p-6 max-w-4xl w-full mx-4 max-h-[80vh] flex flex-col">
            <div class="flex justify-between items-center mb-4">
                <h3 class="text-lg font-semibold">Container Logs</h3>
                <button onclick="closeLogsModal()" class="text-gray-400 hover:text-gray-600">
                    <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                    </svg>
                </button>
            </div>
            <pre id="logsContent" class="bg-gray-900 text-green-400 rounded p-4 overflow-auto flex-1 text-sm font-mono"></pre>
        </div>
    </div>

    <!-- Metrics Modal -->
    <div id="metricsModal" class="fixed inset-0 bg-black bg-opacity-50 hidden items-center justify-center z-50">
        <div class="bg-white rounded-lg p-6 max-w-2xl w-full mx-4">
            <div class="flex justify-between items-center mb-4">
                <h3 class="text-lg font-semibold">Container Metrics</h3>
                <button onclick="closeMetricsModal()" class="text-gray-400 hover:text-gray-600">
                    <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                    </svg>
                </button>
            </div>
            <div id="metricsContent" class="space-y-4">
                <div class="grid grid-cols-2 gap-4">
                    <div class="bg-blue-50 p-4 rounded">
                        <div class="text-sm text-gray-600">CPU Usage</div>
                        <div id="cpuUsage" class="text-2xl font-bold text-blue-600">-</div>
                    </div>
                    <div class="bg-green-50 p-4 rounded">
                        <div class="text-sm text-gray-600">Memory Usage</div>
                        <div id="memoryUsage" class="text-2xl font-bold text-green-600">-</div>
                    </div>
                    <div class="bg-purple-50 p-4 rounded">
                        <div class="text-sm text-gray-600">Network RX</div>
                        <div id="networkRx" class="text-2xl font-bold text-purple-600">-</div>
                    </div>
                    <div class="bg-orange-50 p-4 rounded">
                        <div class="text-sm text-gray-600">Network TX</div>
                        <div id="networkTx" class="text-2xl font-bold text-orange-600">-</div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script>
        let ws;
        
        function connectWebSocket() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            ws = new WebSocket(protocol + '//' + window.location.host + '/ws');
            
            ws.onmessage = function(event) {
                const data = JSON.parse(event.data);
                updateContainerTable(data.containers || []);
            };
            
            ws.onclose = function() {
                setTimeout(connectWebSocket, 3000);
            };
        }

        function updateContainerTable(containers) {
            const tbody = document.getElementById('containerTable');
            tbody.innerHTML = '';
            
            containers.forEach(container => {
                const row = document.createElement('tr');
                const statusClass = container.status.includes('running') ? 'status-running' : 'status-exited';
                
                row.innerHTML = ` + "`" + `
                    <td class="px-6 py-4 text-sm font-mono text-gray-500">${container.id.substring(0, 12)}</td>
                    <td class="px-6 py-4 text-sm text-gray-900">${container.name}</td>
                    <td class="px-6 py-4 text-sm text-gray-500">${container.image}</td>
                    <td class="px-6 py-4 text-sm ${statusClass}">${container.status}</td>
                    <td class="px-6 py-4 text-sm text-gray-500">${container.ports || '-'}</td>
                    <td class="px-6 py-4 text-sm text-gray-500">${container.uptime || '-'}</td>
                    <td class="px-6 py-4 text-sm space-x-2">
                        <button onclick="viewLogs('${container.id}')" 
                                class="text-blue-600 hover:text-blue-900">Logs</button>
                        <button onclick="viewMetrics('${container.id}')" 
                                class="text-green-600 hover:text-green-900">Metrics</button>
                        <button onclick="deleteContainer('${container.id}')" 
                                class="text-red-600 hover:text-red-900">Delete</button>
                    </td>
                ` + "`" + `;
                tbody.appendChild(row);
            });
        }

        async function createContainer() {
            const image = document.getElementById('imageInput').value || 'nginx:latest';
            const name = document.getElementById('nameInput').value;
            const ports = document.getElementById('portsInput').value;

            try {
                const response = await fetch('/api/v1/containers', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ image, name, ports })
                });
                
                const result = await response.json();
                if (result.success) {
                    document.getElementById('imageInput').value = '';
                    document.getElementById('nameInput').value = '';
                    document.getElementById('portsInput').value = '';
                } else {
                    alert('Error: ' + result.error);
                }
            } catch (error) {
                alert('Error creating container: ' + error.message);
            }
        }

        async function deleteContainer(id) {
            if (!confirm('Are you sure you want to delete this container?')) return;
            
            try {
                const response = await fetch('/api/v1/containers/' + id, {
                    method: 'DELETE'
                });
                const result = await response.json();
                if (!result.success) {
                    alert('Error: ' + result.error);
                }
            } catch (error) {
                alert('Error deleting container: ' + error.message);
            }
        }

        async function viewLogs(id) {
            try {
                const response = await fetch('/api/v1/containers/' + id + '/logs');
                const result = await response.json();
                
                if (result.success) {
                    document.getElementById('logsContent').textContent = result.data || 'No logs available';
                    document.getElementById('logsModal').classList.remove('hidden');
                    document.getElementById('logsModal').classList.add('flex');
                } else {
                    alert('Error: ' + result.error);
                }
            } catch (error) {
                alert('Error fetching logs: ' + error.message);
            }
        }

        async function viewMetrics(id) {
            try {
                const response = await fetch('/api/v1/containers/' + id + '/metrics');
                const result = await response.json();
                
                if (result.success) {
                    const metrics = result.data;
                    document.getElementById('cpuUsage').textContent = metrics.cpu_percent.toFixed(2) + '%';
                    document.getElementById('memoryUsage').textContent = formatBytes(metrics.memory_usage) + ' / ' + formatBytes(metrics.memory_limit);
                    document.getElementById('networkRx').textContent = formatBytes(metrics.network_rx);
                    document.getElementById('networkTx').textContent = formatBytes(metrics.network_tx);
                    
                    document.getElementById('metricsModal').classList.remove('hidden');
                    document.getElementById('metricsModal').classList.add('flex');
                } else {
                    alert('Error: ' + result.error);
                }
            } catch (error) {
                alert('Error fetching metrics: ' + error.message);
            }
        }

        function closeMetricsModal() {
            document.getElementById('metricsModal').classList.add('hidden');
            document.getElementById('metricsModal').classList.remove('flex');
        }

        function formatBytes(bytes) {
            if (bytes === 0) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
        }

        function closeLogsModal() {
            document.getElementById('logsModal').classList.add('hidden');
            document.getElementById('logsModal').classList.remove('flex');
        }

        // Initialize
        connectWebSocket();
        
        // Load initial data
        fetch('/api/v1/containers')
            .then(response => response.json())
            .then(result => {
                if (result.success) {
                    updateContainerTable(result.data || []);
                }
            });
    </script>
</body>
</html>`

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}
