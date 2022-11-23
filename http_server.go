package mybench

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"time"
)

//go:embed webui
var webuiFiles embed.FS

type StatusData struct {
	CurrentTime   float64
	Note          string
	Workloads     []string
	DataSnapshots []*DataSnapshot
}

type HttpServer struct {
	benchmark *Benchmark
	note      string
	mux       *http.ServeMux
	port      int
}

func NewHttpServer(benchmark *Benchmark, note string, port int) *HttpServer {
	s := &HttpServer{
		benchmark: benchmark,
		note:      note,
		mux:       http.NewServeMux(),
		port:      port,
	}

	subFS, err := fs.Sub(webuiFiles, "webui")
	if err != nil {
		panic(err)
	}

	s.mux.Handle("/", http.FileServer(http.FS(subFS)))
	s.mux.HandleFunc("/api/status", s.apiStatus)
	return s
}

func (s *HttpServer) apiStatus(w http.ResponseWriter, req *http.Request) {
	var statusData StatusData
	statusData.DataSnapshots = s.benchmark.DataSnapshots()
	statusData.Note = s.note

	statusData.Workloads = make([]string, 0, len(s.benchmark.workloads))
	for workloadName := range s.benchmark.workloads {
		statusData.Workloads = append(statusData.Workloads, workloadName)
	}

	statusData.CurrentTime = time.Since(s.benchmark.startTime).Seconds()

	w.Header().Add("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(statusData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

func (h *HttpServer) Run() {
	host := fmt.Sprintf("localhost:%d", h.port)
	fmt.Printf("Starting HTTP server at http://%s\n", host)
	http.ListenAndServe(host, h.mux)
}
