package taskmanager

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strings"
	"time"
)

// WebHandler creates an HTTP handler for the task manager web interface.
// The baseURL parameter should be the URL prefix where the handler is mounted (e.g., "/tasks").
// Returns a ServeMux that can be integrated into your HTTP server.
func (tm *TaskManager) WebHandler(baseURL string) *http.ServeMux {
	mux := http.NewServeMux()

	// Normalize baseURL
	baseURL = strings.TrimSuffix(baseURL, "/")
	if baseURL == "" {
		baseURL = "/"
	}

	// Register routes
	mux.HandleFunc(baseURL+"/", tm.handleIndex(baseURL))
	mux.HandleFunc(baseURL+"/task/", tm.handleTaskAction(baseURL))
	mux.HandleFunc(baseURL+"/stats", tm.handleStats(baseURL))

	return mux
}

// handleIndex renders the main task list page
func (tm *TaskManager) handleIndex(baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		tm.renderIndexWithMessage(w, baseURL, "", "")
	}
}

// handleTaskAction handles task operations (enable, disable, run, remove)
func (tm *TaskManager) handleTaskAction(baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		taskName := r.FormValue("name")
		action := r.FormValue("action")

		if taskName == "" || action == "" {
			http.Error(w, "Missing name or action parameter", http.StatusBadRequest)
			return
		}

		var err error
		var message string
		var messageType string

		switch action {
		case "enable":
			err = tm.EnableTask(taskName)
			message = fmt.Sprintf("Task '%s' enabled successfully", taskName)
			messageType = "success"
		case "disable":
			err = tm.DisableTask(taskName)
			message = fmt.Sprintf("Task '%s' disabled successfully", taskName)
			messageType = "success"
		case "run":
			err = tm.RunTaskNow(taskName)
			message = fmt.Sprintf("Task '%s' triggered successfully", taskName)
			messageType = "success"
		case "remove":
			err = tm.RemoveTask(taskName)
			message = fmt.Sprintf("Task '%s' removed successfully", taskName)
			messageType = "success"
		default:
			http.Error(w, "Invalid action", http.StatusBadRequest)
			return
		}

		if err != nil {
			message = fmt.Sprintf("Failed to %s task '%s': %s", action, taskName, err.Error())
			messageType = "error"
		}

		// Render the page with inline message
		tm.renderIndexWithMessage(w, baseURL, message, messageType)
	}
}

// renderIndexWithMessage renders the index page with a message banner
func (tm *TaskManager) renderIndexWithMessage(w http.ResponseWriter, baseURL, message, messageType string) {
	tasks := tm.ListTasks()

	// Sort tasks by name
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Name < tasks[j].Name
	})

	data := struct {
		Tasks       []*TaskInfo
		BaseURL     string
		Stats       map[string]interface{}
		Message     string
		MessageType string
	}{
		Tasks:       tasks,
		BaseURL:     baseURL,
		Stats:       tm.GetStats(),
		Message:     message,
		MessageType: messageType,
	}

	tmpl := template.Must(template.New("index").Funcs(templateFuncs).Parse(indexTemplate))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleStats renders the statistics page
func (tm *TaskManager) handleStats(baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		stats := tm.GetStats()
		tasks := tm.ListTasks()

		// Sort tasks by error count (descending)
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].ErrorCount > tasks[j].ErrorCount
		})

		data := struct {
			Stats   map[string]interface{}
			Tasks   []*TaskInfo
			BaseURL string
		}{
			Stats:   stats,
			Tasks:   tasks,
			BaseURL: baseURL,
		}

		tmpl := template.Must(template.New("stats").Funcs(templateFuncs).Parse(statsTemplate))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// Template functions
var templateFuncs = template.FuncMap{
	"formatTime": func(t time.Time) string {
		if t.IsZero() {
			return "Never"
		}
		return t.Format("2006-01-02 15:04:05")
	},
	"formatDuration": func(t time.Time) string {
		if t.IsZero() {
			return "N/A"
		}
		duration := time.Until(t)
		if duration < 0 {
			return "Overdue"
		}
		if duration < time.Minute {
			return fmt.Sprintf("%ds", int(duration.Seconds()))
		}
		if duration < time.Hour {
			return fmt.Sprintf("%dm", int(duration.Minutes()))
		}
		if duration < 24*time.Hour {
			return fmt.Sprintf("%dh %dm", int(duration.Hours()), int(duration.Minutes())%60)
		}
		return fmt.Sprintf("%dd %dh", int(duration.Hours()/24), int(duration.Hours())%24)
	},
	"successRate": func(runCount, errorCount int64) string {
		if runCount == 0 {
			return "N/A"
		}
		rate := float64(runCount-errorCount) / float64(runCount) * 100
		return fmt.Sprintf("%.1f%%", rate)
	},
	"statusBadge": func(enabled, running bool) template.HTML {
		if running {
			return `<span class="badge badge-running">Running</span>`
		}
		if enabled {
			return `<span class="badge badge-enabled">Enabled</span>`
		}
		return `<span class="badge badge-disabled">Disabled</span>`
	},
}

// HTML Templates
const indexTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Task Manager</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 1400px;
            margin: 0 auto;
        }
        .header {
            background: white;
            border-radius: 12px;
            padding: 30px;
            margin-bottom: 20px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
        }
        .header h1 {
            color: #333;
            font-size: 32px;
            margin-bottom: 10px;
        }
        .header p {
            color: #666;
            font-size: 14px;
        }
        .nav {
            display: flex;
            gap: 10px;
            margin-top: 20px;
        }
        .nav a {
            padding: 10px 20px;
            background: #667eea;
            color: white;
            text-decoration: none;
            border-radius: 6px;
            font-size: 14px;
            transition: background 0.3s;
        }
        .nav a:hover {
            background: #5568d3;
        }
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
            margin-bottom: 20px;
        }
        .stat-card {
            background: white;
            padding: 20px;
            border-radius: 12px;
            box-shadow: 0 4px 15px rgba(0,0,0,0.1);
        }
        .stat-card h3 {
            color: #999;
            font-size: 12px;
            text-transform: uppercase;
            letter-spacing: 1px;
            margin-bottom: 8px;
        }
        .stat-card .value {
            color: #333;
            font-size: 28px;
            font-weight: bold;
        }
        .tasks-container {
            background: white;
            border-radius: 12px;
            padding: 30px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
        }
        .tasks-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 20px;
            padding-bottom: 15px;
            border-bottom: 2px solid #f0f0f0;
        }
        .tasks-header h2 {
            color: #333;
            font-size: 24px;
        }
        .task-count {
            color: #999;
            font-size: 14px;
        }
        .task-grid {
            display: grid;
            gap: 15px;
        }
        .task-card {
            border: 2px solid #f0f0f0;
            border-radius: 8px;
            padding: 20px;
            transition: all 0.3s;
        }
        .task-card:hover {
            border-color: #667eea;
            box-shadow: 0 4px 12px rgba(102, 126, 234, 0.2);
        }
        .task-header {
            display: flex;
            justify-content: space-between;
            align-items: start;
            margin-bottom: 15px;
        }
        .task-name {
            font-size: 18px;
            font-weight: 600;
            color: #333;
            margin-bottom: 5px;
        }
        .task-schedule {
            color: #666;
            font-size: 13px;
            font-family: monospace;
            background: #f8f8f8;
            padding: 3px 8px;
            border-radius: 4px;
            display: inline-block;
        }
        .task-info {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
            gap: 12px;
            margin-bottom: 15px;
            padding: 15px;
            background: #f8f8f8;
            border-radius: 6px;
        }
        .info-item {
            font-size: 13px;
        }
        .info-label {
            color: #999;
            display: block;
            margin-bottom: 3px;
        }
        .info-value {
            color: #333;
            font-weight: 500;
        }
        .task-actions {
            display: flex;
            gap: 8px;
            flex-wrap: wrap;
        }
        .btn {
            padding: 8px 16px;
            border: none;
            border-radius: 5px;
            font-size: 13px;
            font-weight: 500;
            cursor: pointer;
            transition: all 0.3s;
            text-decoration: none;
            display: inline-block;
        }
        .btn-enable {
            background: #10b981;
            color: white;
        }
        .btn-enable:hover {
            background: #059669;
        }
        .btn-disable {
            background: #f59e0b;
            color: white;
        }
        .btn-disable:hover {
            background: #d97706;
        }
        .btn-run {
            background: #3b82f6;
            color: white;
        }
        .btn-run:hover {
            background: #2563eb;
        }
        .btn-remove {
            background: #ef4444;
            color: white;
        }
        .btn-remove:hover {
            background: #dc2626;
        }
        .badge {
            padding: 4px 10px;
            border-radius: 12px;
            font-size: 11px;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        .badge-enabled {
            background: #d1fae5;
            color: #065f46;
        }
        .badge-disabled {
            background: #fee2e2;
            color: #991b1b;
        }
        .badge-running {
            background: #dbeafe;
            color: #1e40af;
            animation: pulse 2s infinite;
        }
        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.7; }
        }
        .error-info {
            background: #fef2f2;
            border-left: 3px solid #ef4444;
            padding: 10px;
            margin-top: 10px;
            border-radius: 4px;
        }
        .error-info strong {
            color: #991b1b;
            font-size: 12px;
            display: block;
            margin-bottom: 5px;
        }
        .error-info span {
            color: #dc2626;
            font-size: 12px;
        }
        .empty-state {
            text-align: center;
            padding: 60px 20px;
            color: #999;
        }
        .empty-state svg {
            width: 80px;
            height: 80px;
            margin-bottom: 20px;
            opacity: 0.3;
        }
        .success-message {
            background: linear-gradient(135deg, #d1fae5 0%, #a7f3d0 100%);
            border-left: 4px solid #10b981;
            color: #065f46;
            padding: 16px 20px;
            margin-bottom: 20px;
            border-radius: 8px;
            display: flex;
            align-items: center;
            gap: 12px;
            box-shadow: 0 4px 12px rgba(16, 185, 129, 0.2);
            animation: slideIn 0.3s ease-out;
        }
        .error-message {
            background: linear-gradient(135deg, #fee2e2 0%, #fecaca 100%);
            border-left: 4px solid #ef4444;
            color: #991b1b;
            padding: 16px 20px;
            margin-bottom: 20px;
            border-radius: 8px;
            display: flex;
            align-items: center;
            gap: 12px;
            box-shadow: 0 4px 12px rgba(239, 68, 68, 0.2);
            animation: slideIn 0.3s ease-out;
        }
        .message-icon {
            font-size: 24px;
            flex-shrink: 0;
        }
        .message-content {
            flex: 1;
        }
        @keyframes slideIn {
            from {
                transform: translateY(-10px);
                opacity: 0;
            }
            to {
                transform: translateY(0);
                opacity: 1;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>⚙️ Task Manager</h1>
            <p>Monitor and manage your scheduled tasks</p>
            <div class="nav">
                <a href="{{.BaseURL}}/">📋 Tasks</a>
                <a href="{{.BaseURL}}/stats">📊 Statistics</a>
            </div>
        </div>

        {{if .Stats}}
        <div class="stats-grid">
            <div class="stat-card">
                <h3>Total Tasks</h3>
                <div class="value">{{.Stats.total_tasks}}</div>
            </div>
            <div class="stat-card">
                <h3>Enabled</h3>
                <div class="value">{{.Stats.enabled_tasks}}</div>
            </div>
            <div class="stat-card">
                <h3>Running</h3>
                <div class="value">{{.Stats.running_tasks}}</div>
            </div>
            <div class="stat-card">
                <h3>Total Executions</h3>
                <div class="value">{{.Stats.total_runs}}</div>
            </div>
            <div class="stat-card">
                <h3>Total Errors</h3>
                <div class="value">{{.Stats.total_errors}}</div>
            </div>
        </div>
        {{end}}

        {{if .Message}}
        <div class="{{if eq .MessageType "error"}}error-message{{else}}success-message{{end}}">
            <div class="message-icon">{{if eq .MessageType "error"}}⚠️{{else}}✅{{end}}</div>
            <div class="message-content">{{.Message}}</div>
        </div>
        {{end}}

        <div class="tasks-container">
            <div class="tasks-header">
                <h2>Tasks</h2>
                <span class="task-count">{{len .Tasks}} task(s)</span>
            </div>

            {{if .Tasks}}
            <div class="task-grid">
                {{range .Tasks}}
                <div class="task-card">
                    <div class="task-header">
                        <div>
                            <div class="task-name">{{.Name}}</div>
                            <span class="task-schedule">{{.Schedule}}</span>
                        </div>
                        <div>
                            {{statusBadge .Enabled .Running}}
                        </div>
                    </div>

                    <div class="task-info">
                        <div class="info-item">
                            <span class="info-label">Last Run</span>
                            <span class="info-value">{{formatTime .LastRun}}</span>
                        </div>
                        <div class="info-item">
                            <span class="info-label">Next Run</span>
                            <span class="info-value">{{formatTime .NextRun}} ({{formatDuration .NextRun}})</span>
                        </div>
                        <div class="info-item">
                            <span class="info-label">Executions</span>
                            <span class="info-value">{{.RunCount}}</span>
                        </div>
                        <div class="info-item">
                            <span class="info-label">Errors</span>
                            <span class="info-value">{{.ErrorCount}}</span>
                        </div>
                        <div class="info-item">
                            <span class="info-label">Success Rate</span>
                            <span class="info-value">{{successRate .RunCount .ErrorCount}}</span>
                        </div>
                        <div class="info-item">
                            <span class="info-label">Added</span>
                            <span class="info-value">{{formatTime .AddedAt}}</span>
                        </div>
                    </div>

                    {{if .LastError}}
                    <div class="error-info">
                        <strong>Last Error:</strong>
                        <span>{{.LastError}}</span>
                    </div>
                    {{end}}

                    <div class="task-actions">
                        {{if .Enabled}}
                        <form method="POST" action="{{$.BaseURL}}/task/" style="display: inline;">
                            <input type="hidden" name="name" value="{{.Name}}">
                            <input type="hidden" name="action" value="disable">
                            <button type="submit" class="btn btn-disable">⏸️ Disable</button>
                        </form>
                        {{else}}
                        <form method="POST" action="{{$.BaseURL}}/task/" style="display: inline;">
                            <input type="hidden" name="name" value="{{.Name}}">
                            <input type="hidden" name="action" value="enable">
                            <button type="submit" class="btn btn-enable">▶️ Enable</button>
                        </form>
                        {{end}}

                        {{if not .Running}}
                        <form method="POST" action="{{$.BaseURL}}/task/" style="display: inline;">
                            <input type="hidden" name="name" value="{{.Name}}">
                            <input type="hidden" name="action" value="run">
                            <button type="submit" class="btn btn-run">🚀 Run Now</button>
                        </form>
                        {{end}}

                        <form method="POST" action="{{$.BaseURL}}/task/" style="display: inline;" onsubmit="return confirm('Are you sure you want to remove this task?');">
                            <input type="hidden" name="name" value="{{.Name}}">
                            <input type="hidden" name="action" value="remove">
                            <button type="submit" class="btn btn-remove">🗑️ Remove</button>
                        </form>
                    </div>
                </div>
                {{end}}
            </div>
            {{else}}
            <div class="empty-state">
                <svg fill="currentColor" viewBox="0 0 20 20">
                    <path d="M9 2a1 1 0 000 2h2a1 1 0 100-2H9z"/>
                    <path fill-rule="evenodd" d="M4 5a2 2 0 012-2 3 3 0 003 3h2a3 3 0 003-3 2 2 0 012 2v11a2 2 0 01-2 2H6a2 2 0 01-2-2V5zm3 4a1 1 0 000 2h.01a1 1 0 100-2H7zm3 0a1 1 0 000 2h3a1 1 0 100-2h-3zm-3 4a1 1 0 100 2h.01a1 1 0 100-2H7zm3 0a1 1 0 100 2h3a1 1 0 100-2h-3z" clip-rule="evenodd"/>
                </svg>
                <h3>No tasks yet</h3>
                <p>Add tasks programmatically using the TaskManager API</p>
            </div>
            {{end}}
        </div>
    </div>
</body>
</html>
`

const statsTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Task Manager - Statistics</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        .header {
            background: white;
            border-radius: 12px;
            padding: 30px;
            margin-bottom: 20px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
        }
        .header h1 {
            color: #333;
            font-size: 32px;
            margin-bottom: 10px;
        }
        .nav {
            display: flex;
            gap: 10px;
            margin-top: 20px;
        }
        .nav a {
            padding: 10px 20px;
            background: #667eea;
            color: white;
            text-decoration: none;
            border-radius: 6px;
            font-size: 14px;
            transition: background 0.3s;
        }
        .nav a:hover {
            background: #5568d3;
        }
        .stats-container {
            background: white;
            border-radius: 12px;
            padding: 30px;
            margin-bottom: 20px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
        }
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
        }
        .stat-box {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 25px;
            border-radius: 10px;
            box-shadow: 0 4px 15px rgba(0,0,0,0.1);
        }
        .stat-box h3 {
            font-size: 14px;
            opacity: 0.9;
            margin-bottom: 10px;
            text-transform: uppercase;
            letter-spacing: 1px;
        }
        .stat-box .value {
            font-size: 36px;
            font-weight: bold;
        }
        .stat-box .description {
            font-size: 12px;
            opacity: 0.8;
            margin-top: 5px;
        }
        .tasks-table {
            background: white;
            border-radius: 12px;
            padding: 30px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
            overflow-x: auto;
        }
        .tasks-table h2 {
            color: #333;
            margin-bottom: 20px;
            padding-bottom: 15px;
            border-bottom: 2px solid #f0f0f0;
        }
        table {
            width: 100%;
            border-collapse: collapse;
        }
        th {
            background: #f8f8f8;
            padding: 12px;
            text-align: left;
            font-size: 13px;
            color: #666;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            font-weight: 600;
        }
        td {
            padding: 15px 12px;
            border-bottom: 1px solid #f0f0f0;
            font-size: 14px;
        }
        tr:hover {
            background: #f8f8f8;
        }
        .progress-bar {
            width: 100%;
            height: 8px;
            background: #e5e7eb;
            border-radius: 4px;
            overflow: hidden;
        }
        .progress-fill {
            height: 100%;
            background: linear-gradient(90deg, #10b981, #059669);
            transition: width 0.3s;
        }
        .error-bar .progress-fill {
            background: linear-gradient(90deg, #ef4444, #dc2626);
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📊 Statistics</h1>
            <div class="nav">
                <a href="{{.BaseURL}}/">📋 Tasks</a>
                <a href="{{.BaseURL}}/stats">📊 Statistics</a>
            </div>
        </div>

        <div class="stats-container">
            <div class="stats-grid">
                <div class="stat-box">
                    <h3>Total Tasks</h3>
                    <div class="value">{{.Stats.total_tasks}}</div>
                    <div class="description">Registered in the system</div>
                </div>
                <div class="stat-box">
                    <h3>Enabled Tasks</h3>
                    <div class="value">{{.Stats.enabled_tasks}}</div>
                    <div class="description">Currently active</div>
                </div>
                <div class="stat-box">
                    <h3>Running Tasks</h3>
                    <div class="value">{{.Stats.running_tasks}}</div>
                    <div class="description">Executing right now</div>
                </div>
                <div class="stat-box">
                    <h3>Total Executions</h3>
                    <div class="value">{{.Stats.total_runs}}</div>
                    <div class="description">All time</div>
                </div>
                <div class="stat-box">
                    <h3>Total Errors</h3>
                    <div class="value">{{.Stats.total_errors}}</div>
                    <div class="description">Failed executions</div>
                </div>
                <div class="stat-box">
                    <h3>Max Concurrent</h3>
                    <div class="value">{{.Stats.max_concurrent}}</div>
                    <div class="description">{{if eq .Stats.max_concurrent 0}}Unlimited{{else}}Limit{{end}}</div>
                </div>
            </div>
        </div>

        <div class="tasks-table">
            <h2>Task Performance</h2>
            {{if .Tasks}}
            <table>
                <thead>
                    <tr>
                        <th>Task Name</th>
                        <th>Executions</th>
                        <th>Errors</th>
                        <th>Success Rate</th>
                        <th>Last Run</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Tasks}}
                    <tr>
                        <td><strong>{{.Name}}</strong></td>
                        <td>{{.RunCount}}</td>
                        <td>{{.ErrorCount}}</td>
                        <td>
                            {{if gt .RunCount 0}}
                            <div class="progress-bar {{if gt .ErrorCount 0}}error-bar{{end}}">
                                <div class="progress-fill" style="width: {{successRate .RunCount .ErrorCount}}"></div>
                            </div>
                            <small>{{successRate .RunCount .ErrorCount}}</small>
                            {{else}}
                            N/A
                            {{end}}
                        </td>
                        <td>{{formatTime .LastRun}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
            {{else}}
            <p style="text-align: center; color: #999; padding: 40px;">No task data available</p>
            {{end}}
        </div>
    </div>
</body>
</html>
`
