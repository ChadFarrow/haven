package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type MonitorStatus struct {
	Status        string    `json:"status"`
	Uptime        string    `json:"uptime"`
	GoVersion     string    `json:"go_version"`
	Memory        string    `json:"memory"`
	Goroutines    int       `json:"goroutines"`
	LastLogEntry  string    `json:"last_log_entry"`
	LogFileSize   string    `json:"log_file_size"`
	DatabaseSize  string    `json:"database_size"`
	BlossomFiles  int       `json:"blossom_files"`
	Timestamp     time.Time `json:"timestamp"`
}

var startTime = time.Now()

func startMonitor() {
	http.HandleFunc("/monitor", monitorHandler)
	http.HandleFunc("/monitor/api/status", statusAPIHandler)
	http.HandleFunc("/monitor/api/logs", logsAPIHandler)
	http.HandleFunc("/monitor/api/upload", uploadHandler)
	http.HandleFunc("/monitor/static/", monitorStaticHandler)
	
	log.Println("🔍 Monitor server starting on :8082")
	go func() {
		if err := http.ListenAndServe(":8082", nil); err != nil {
			log.Printf("Monitor server error: %v", err)
		}
	}()
}

func monitorHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Haven Monitor</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script>
        async function updateStatus() {
            try {
                const response = await fetch('/monitor/api/status');
                const data = await response.json();
                
                // Smooth update function to prevent glitches
                function smoothUpdate(elementId, newValue) {
                    const element = document.getElementById(elementId);
                    if (element && element.textContent !== newValue) {
                        element.style.opacity = '0.7';
                        setTimeout(() => {
                            element.textContent = newValue;
                            element.style.opacity = '1';
                        }, 50);
                    }
                }
                
                smoothUpdate('status', data.status);
                smoothUpdate('uptime', data.uptime);
                smoothUpdate('memory', data.memory);
                smoothUpdate('goroutines', data.goroutines);
                smoothUpdate('log-size', data.log_file_size);
                smoothUpdate('db-size', data.database_size);
                smoothUpdate('blossom-files', data.blossom_files);
                
                // Update status indicator
                const statusEl = document.getElementById('status-indicator');
                statusEl.className = data.status === 'running' ? 
                    'w-3 h-3 bg-green-500 rounded-full animate-pulse' : 
                    'w-3 h-3 bg-red-500 rounded-full';
            } catch (error) {
                console.error('Failed to update status:', error);
                document.getElementById('status').textContent = 'error';
                document.getElementById('status-indicator').className = 'w-3 h-3 bg-red-500 rounded-full';
            }
        }

        async function loadLogs() {
            try {
                const response = await fetch('/monitor/api/logs?lines=50');
                const data = await response.json();
                document.getElementById('logs').textContent = data.logs;
                document.getElementById('last-updated').textContent = new Date().toLocaleTimeString();
            } catch (error) {
                document.getElementById('logs').textContent = 'Error loading logs: ' + error.message;
            }
        }

        // Auto-refresh every 5 seconds
        setInterval(updateStatus, 5000);
        setInterval(loadLogs, 10000);
        
        // File upload functionality
        function setupFileUpload() {
            const fileInput = document.getElementById('fileInput');
            const dropzone = document.getElementById('dropzone');
            const uploadStatus = document.getElementById('uploadStatus');
            const fileList = document.getElementById('fileList');

            // Click to select files
            dropzone.addEventListener('click', () => fileInput.click());

            // Drag and drop
            dropzone.addEventListener('dragover', (e) => {
                e.preventDefault();
                dropzone.classList.add('border-purple-500', 'bg-gray-700');
            });

            dropzone.addEventListener('dragleave', () => {
                dropzone.classList.remove('border-purple-500', 'bg-gray-700');
            });

            dropzone.addEventListener('drop', (e) => {
                e.preventDefault();
                dropzone.classList.remove('border-purple-500', 'bg-gray-700');
                const files = e.dataTransfer.files;
                handleFiles(files);
            });

            fileInput.addEventListener('change', (e) => {
                handleFiles(e.target.files);
            });

            function handleFiles(files) {
                if (files.length === 0) return;

                const formData = new FormData();
                Array.from(files).forEach(file => {
                    formData.append('files', file);
                });

                uploadStatus.className = 'mt-4 text-sm text-blue-400';
                uploadStatus.textContent = 'Uploading files...';
                uploadStatus.classList.remove('hidden');

                fetch('/monitor/api/upload', {
                    method: 'POST',
                    body: formData
                })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        uploadStatus.className = 'mt-4 text-sm text-green-400';
                        uploadStatus.textContent = 'Successfully uploaded ' + data.files.length + ' file(s)';
                        
                        fileList.innerHTML = '';
                        data.files.forEach(file => {
                            const fileItem = document.createElement('div');
                            fileItem.className = file.success ? 
                                'text-sm text-green-400 bg-gray-700 p-2 rounded' : 
                                'text-sm text-red-400 bg-gray-700 p-2 rounded';
                            fileItem.textContent = file.success ? 
                                '✓ ' + file.filename : 
                                '✗ ' + file.filename + ': ' + file.error;
                            fileList.appendChild(fileItem);
                        });
                    } else {
                        uploadStatus.className = 'mt-4 text-sm text-red-400';
                        uploadStatus.textContent = 'Upload failed';
                    }
                })
                .catch(error => {
                    uploadStatus.className = 'mt-4 text-sm text-red-400';
                    uploadStatus.textContent = 'Upload error: ' + error.message;
                });
            }
        }

        // Load initial data
        window.onload = function() {
            updateStatus();
            loadLogs();
            setupFileUpload();
        };
    </script>
</head>
<body class="bg-gray-900 text-white min-h-screen">
    <div class="container mx-auto px-4 py-8">
        <div class="flex items-center gap-3 mb-8">
            <div id="status-indicator" class="w-3 h-3 bg-green-500 rounded-full animate-pulse"></div>
            <h1 class="text-3xl font-bold text-purple-400">Haven Monitor</h1>
        </div>
        
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
            <div class="bg-gray-800 rounded-lg p-6">
                <h3 class="text-lg font-semibold mb-2 text-gray-300">Status</h3>
                <p id="status" class="text-2xl font-bold text-green-400">loading...</p>
            </div>
            
            <div class="bg-gray-800 rounded-lg p-6">
                <h3 class="text-lg font-semibold mb-2 text-gray-300">Uptime</h3>
                <p id="uptime" class="text-2xl font-bold text-blue-400">loading...</p>
            </div>
            
            <div class="bg-gray-800 rounded-lg p-6">
                <h3 class="text-lg font-semibold mb-2 text-gray-300">Memory</h3>
                <p id="memory" class="text-2xl font-bold text-yellow-400">loading...</p>
            </div>
            
            <div class="bg-gray-800 rounded-lg p-6">
                <h3 class="text-lg font-semibold mb-2 text-gray-300">Goroutines</h3>
                <p id="goroutines" class="text-2xl font-bold text-cyan-400">loading...</p>
            </div>
        </div>
        
        <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
            <div class="bg-gray-800 rounded-lg p-6">
                <h3 class="text-lg font-semibold mb-2 text-gray-300">Log File Size</h3>
                <p id="log-size" class="text-xl font-bold text-orange-400">loading...</p>
            </div>
            
            <div class="bg-gray-800 rounded-lg p-6">
                <h3 class="text-lg font-semibold mb-2 text-gray-300">Database Size</h3>
                <p id="db-size" class="text-xl font-bold text-pink-400">loading...</p>
            </div>
            
            <div class="bg-gray-800 rounded-lg p-6">
                <h3 class="text-lg font-semibold mb-2 text-gray-300">Blossom Files</h3>
                <p id="blossom-files" class="text-xl font-bold text-indigo-400">loading...</p>
            </div>
        </div>
        
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
            <div class="bg-gray-800 rounded-lg p-6">
                <h3 class="text-lg font-semibold mb-4 text-gray-300">File Upload</h3>
                <div class="border-2 border-dashed border-gray-600 rounded-lg p-6 text-center">
                    <input type="file" id="fileInput" class="hidden" multiple accept="*/*">
                    <div id="dropzone" class="cursor-pointer">
                        <svg class="w-12 h-12 mx-auto mb-3 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"></path>
                        </svg>
                        <p class="text-gray-400">Click to select files or drag and drop</p>
                        <p class="text-sm text-gray-500 mt-1">Multiple files supported</p>
                    </div>
                </div>
                <div id="uploadStatus" class="mt-4 text-sm hidden"></div>
                <div id="fileList" class="mt-4 space-y-2"></div>
            </div>
            
            <div class="bg-gray-800 rounded-lg p-6">
                <div class="flex justify-between items-center mb-4">
                    <h3 class="text-lg font-semibold text-gray-300">Recent Logs</h3>
                    <p class="text-sm text-gray-500">Last updated: <span id="last-updated">never</span></p>
                </div>
                <pre id="logs" class="bg-gray-900 p-4 rounded text-sm text-green-400 font-mono overflow-x-auto max-h-96 overflow-y-auto">Loading logs...</pre>
            </div>
        </div>
    </div>
</body>
</html>`

	t, err := template.New("monitor").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, nil)
}

func statusAPIHandler(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	uptime := time.Since(startTime)
	
	// Get log file info
	logSize := getFileSize("haven.log")
	lastLogEntry := getLastLogEntry("haven.log")
	
	// Get database size
	dbSize := getDirSize("db")
	
	// Get blossom files count
	blossomCount := getFileCount("blossom")
	
	status := MonitorStatus{
		Status:       "running",
		Uptime:       formatDuration(uptime),
		GoVersion:    runtime.Version(),
		Memory:       formatBytes(int64(m.Alloc)),
		Goroutines:   runtime.NumGoroutine(),
		LastLogEntry: lastLogEntry,
		LogFileSize:  formatBytes(logSize),
		DatabaseSize: formatBytes(dbSize),
		BlossomFiles: blossomCount,
		Timestamp:    time.Now(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func logsAPIHandler(w http.ResponseWriter, r *http.Request) {
	lines := 50
	if l := r.URL.Query().Get("lines"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			lines = parsed
		}
	}
	
	logs := getTailLogs("haven.log", lines)
	
	response := map[string]string{
		"logs": logs,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form with 32MB max memory
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	uploadDir := "./uploads"
	
	// Ensure upload directory exists
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		http.Error(w, "Failed to create upload directory", http.StatusInternalServerError)
		return
	}

	var results []map[string]interface{}
	
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			results = append(results, map[string]interface{}{
				"filename": fileHeader.Filename,
				"success":  false,
				"error":    "Failed to open file",
			})
			continue
		}
		defer file.Close()

		// Create destination file
		destPath := filepath.Join(uploadDir, fileHeader.Filename)
		destFile, err := os.Create(destPath)
		if err != nil {
			results = append(results, map[string]interface{}{
				"filename": fileHeader.Filename,
				"success":  false,
				"error":    "Failed to create destination file",
			})
			continue
		}
		defer destFile.Close()

		// Copy file contents
		_, err = io.Copy(destFile, file)
		if err != nil {
			results = append(results, map[string]interface{}{
				"filename": fileHeader.Filename,
				"success":  false,
				"error":    "Failed to copy file",
			})
			continue
		}

		results = append(results, map[string]interface{}{
			"filename": fileHeader.Filename,
			"success":  true,
			"path":     destPath,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"files":   results,
	})
}

func monitorStaticHandler(w http.ResponseWriter, r *http.Request) {
	// Simple static file serving for any additional assets
	http.ServeFile(w, r, r.URL.Path[1:])
}

func getFileSize(filename string) int64 {
	info, err := os.Stat(filename)
	if err != nil {
		return 0
	}
	return info.Size()
}

func getDirSize(dirname string) int64 {
	var size int64
	err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0
	}
	return size
}

func getFileCount(dirname string) int {
	files, err := os.ReadDir(dirname)
	if err != nil {
		return 0
	}
	return len(files)
}

func getLastLogEntry(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		return "No log file found"
	}
	defer file.Close()
	
	// Read last 1KB to get the last line
	stat, err := file.Stat()
	if err != nil {
		return "Error reading log"
	}
	
	size := stat.Size()
	if size == 0 {
		return "Log file is empty"
	}
	
	start := size - 1024
	if start < 0 {
		start = 0
	}
	
	_, err = file.Seek(start, 0)
	if err != nil {
		return "Error seeking log"
	}
	
	buf := make([]byte, 1024)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "Error reading log"
	}
	
	lines := strings.Split(string(buf[:n]), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			return strings.TrimSpace(lines[i])
		}
	}
	
	return "No recent log entries"
}

func getTailLogs(filename string, lines int) string {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Sprintf("Error opening log file: %v", err)
	}
	defer file.Close()
	
	// Simple implementation - read entire file and get last N lines
	content, err := io.ReadAll(file)
	if err != nil {
		return fmt.Sprintf("Error reading log file: %v", err)
	}
	
	allLines := strings.Split(string(content), "\n")
	start := len(allLines) - lines
	if start < 0 {
		start = 0
	}
	
	return strings.Join(allLines[start:], "\n")
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
	return fmt.Sprintf("%.1fd", d.Hours()/24)
}