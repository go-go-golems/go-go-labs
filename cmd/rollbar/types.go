package main

type Instance struct {
	ID        int64 `json:"id"`
	ProjectID int64 `json:"project_id"`
	Timestamp int64 `json:"timestamp"`
	Version   int   `json:"version"`
	Data      Data  `json:"data"`
	Billable  int   `json:"billable"`
	ItemID    int64 `json:"item_id"`
}

type Data struct {
	Environment   string      `json:"environment"`
	Body          Body        `json:"body"`
	Level         string      `json:"level"`
	Timestamp     int64       `json:"timestamp"`
	CodeVersion   string      `json:"code_version"`
	Platform      string      `json:"platform"`
	Language      string      `json:"language"`
	Request       Request     `json:"request"`
	Server        Server      `json:"server"`
	Custom        interface{} `json:"custom"`
	Uuid          string      `json:"uuid"`
	Notifier      Notifier    `json:"notifier"`
	Metadata      interface{} `json:"metadata"`
	Framework     string      `json:"framework"`
	RetentionDays int         `json:"retentionDays"`
}

type Body struct {
	Trace Trace `json:"trace"`
	// Add other fields like trace_chain, message, or crash_report if needed
}

type Trace struct {
	Frames    []Frame   `json:"frames"`
	Exception Exception `json:"exception"`
}

type Frame struct {
	// Define the fields based on your requirement
}

type Exception struct {
	// Define the fields based on your requirement
}

type Request struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	GET     interface{}       `json:"GET"`
	POST    interface{}       `json:"POST"`
	UserIP  string            `json:"user_ip"`
}

type Server struct {
	Host string `json:"host"`
	// Add other fields as needed
}

type Notifier struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
