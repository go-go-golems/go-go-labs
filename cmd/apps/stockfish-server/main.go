package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type StockfishServer struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Scanner
	mutex  sync.Mutex
}

type AnalysisRequest struct {
	Fen   string `json:"fen,omitempty"`
	Depth int    `json:"depth"`
	Moves string `json:"moves,omitempty"` // e2e4 e7e5 ...
}

type AnalysisResponse struct {
	BestMove           string   `json:"bestMove"`
	Score              int      `json:"score"`
	Depth              int      `json:"depth"`
	PrincipalVariation []string `json:"pv"`
	TimeMs             int      `json:"timeMs"`
	Nodes              int64    `json:"nodes"`
}

func NewStockfishServer() (*StockfishServer, error) {
	cmd := exec.Command("stockfish")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start stockfish: %v", err)
	}

	server := &StockfishServer{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewScanner(stdout),
	}

	// Initialize UCI mode
	server.sendCommand("uci")
	server.waitForResponse("uciok")
	server.sendCommand("isready")
	server.waitForResponse("readyok")

	return server, nil
}

func (s *StockfishServer) sendCommand(cmd string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, err := io.WriteString(s.stdin, cmd+"\n")
	return err
}

func (s *StockfishServer) waitForResponse(expected string) string {
	var lastLine string
	for s.stdout.Scan() {
		line := s.stdout.Text()
		if strings.Contains(line, expected) {
			return line
		}
		lastLine = line
	}
	return lastLine
}

func (s *StockfishServer) analyze(req AnalysisRequest) (*AnalysisResponse, error) {
	// Set up position
	var posCmd string
	if req.Fen != "" {
		posCmd = fmt.Sprintf("position fen %s", req.Fen)
	} else {
		posCmd = "position startpos"
	}
	if req.Moves != "" {
		posCmd += " moves " + req.Moves
	}

	if err := s.sendCommand(posCmd); err != nil {
		return nil, err
	}

	// Start analysis
	if err := s.sendCommand(fmt.Sprintf("go depth %d", req.Depth)); err != nil {
		return nil, err
	}

	response := &AnalysisResponse{}
	startTime := time.Now()

	// Parse output until we get a bestmove
	for s.stdout.Scan() {
		line := s.stdout.Text()

		if strings.HasPrefix(line, "bestmove") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				response.BestMove = parts[1]
			}
			break
		}

		if strings.HasPrefix(line, "info") {
			s.parseInfoLine(line, response)
		}
	}

	response.TimeMs = int(time.Since(startTime).Milliseconds())
	return response, nil
}

func (s *StockfishServer) parseInfoLine(line string, response *AnalysisResponse) {
	fields := strings.Fields(line)

	for i := 0; i < len(fields); i++ {
		switch fields[i] {
		case "depth":
			if i+1 < len(fields) {
				fmt.Sscanf(fields[i+1], "%d", &response.Depth)
			}
		case "score":
			if i+2 < len(fields) && fields[i+1] == "cp" {
				fmt.Sscanf(fields[i+2], "%d", &response.Score)
			}
		case "nodes":
			if i+1 < len(fields) {
				fmt.Sscanf(fields[i+1], "%d", &response.Nodes)
			}
		case "pv":
			if i+1 < len(fields) {
				response.PrincipalVariation = fields[i+1:]
				return // PV is always last, so we can return
			}
		}
	}
}

func (s *StockfishServer) Close() error {
	s.sendCommand("quit")
	return s.cmd.Wait()
}

func main() {
	server, err := NewStockfishServer()
	if err != nil {
		log.Fatalf("Failed to start Stockfish: %v", err)
	}
	defer server.Close()

	http.HandleFunc("/analyze", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req AnalysisRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Depth <= 0 {
			req.Depth = 20 // default depth
		}

		response, err := server.analyze(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	log.Println("Starting Stockfish API server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
