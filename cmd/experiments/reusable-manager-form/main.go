package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/Masterminds/sprig/v3"
)

// Models
type Manager struct {
	ID   int
	Name string
}

type Report struct {
	ID        int
	Name      string
	ManagerID int
}

// Mock database
var managers = []Manager{
	{ID: 1, Name: "Alex Johnson"},
	{ID: 2, Name: "Sam Williams"},
	{ID: 3, Name: "Taylor Smith"},
}

var reports = []Report{
	{ID: 101, Name: "Q1 Sales Report", ManagerID: 1},
	{ID: 102, Name: "Team Performance Review", ManagerID: 1},
	{ID: 103, Name: "Customer Satisfaction Survey", ManagerID: 1},
	{ID: 201, Name: "Marketing Campaign Analysis", ManagerID: 2},
	{ID: 202, Name: "Budget Forecast", ManagerID: 2},
	{ID: 301, Name: "Product Roadmap", ManagerID: 3},
	{ID: 302, Name: "Development Sprint Review", ManagerID: 3},
	{ID: 303, Name: "Quality Assurance Metrics", ManagerID: 3},
}

// Template data structure
type TemplateData struct {
	Managers []Manager
	Reports  []Report
	FormID   string
}

func main() {
	// Set up templates with Sprig functions
	tmpl := template.Must(template.New("").Funcs(sprig.FuncMap()).ParseGlob("templates/*.html"))

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Main page route
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := TemplateData{
			Managers: managers,
			Reports:  []Report{}, // Empty initially
		}
		tmpl.ExecuteTemplate(w, "index.html", data)
	})

	// Reports by manager API endpoint
	http.HandleFunc("/api/reports-by-manager", func(w http.ResponseWriter, r *http.Request) {
		// Get the manager ID from the request
		managerIDStr := r.URL.Query().Get("manager_id")

		// Get the form ID to maintain context
		formID := r.URL.Query().Get("form_id")

		// If no manager selected, return empty list
		if managerIDStr == "" {
			data := TemplateData{
				Reports: []Report{},
				FormID:  formID,
			}
			tmpl.ExecuteTemplate(w, "report_dropdown", data)
			return
		}

		// Convert manager ID to integer
		managerID, err := strconv.Atoi(managerIDStr)
		if err != nil {
			http.Error(w, "Invalid manager ID", http.StatusBadRequest)
			return
		}

		// Filter reports for the selected manager
		var filteredReports []Report
		for _, report := range reports {
			if report.ManagerID == managerID {
				filteredReports = append(filteredReports, report)
			}
		}

		// Render only the report dropdown template
		data := TemplateData{
			Reports: filteredReports,
			FormID:  formID,
		}
		tmpl.ExecuteTemplate(w, "report_dropdown", data)
	})

	// Start the server
	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
