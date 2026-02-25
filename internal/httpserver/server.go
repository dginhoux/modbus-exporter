package httpserver

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/atrabilis/modbus-exporter/internal/store"
)

type Server struct {
	addr  string
	store *store.Store
}

func New(addr string, store *store.Store) *Server {
	return &Server{
		addr:  addr,
		store: store,
	}
}

func (s *Server) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", s.handleMetrics)
	mux.HandleFunc("/health", s.handleHealth)
	log.Printf("http server listening on %s", s.addr)
	return http.ListenAndServe(s.addr, mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	samples := s.store.Snapshot()
	now := time.Now()

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")

	// Comentarios Prometheus: HELP y TYPE para cada métrica.
	fmt.Fprint(w, "# HELP modbus_value Current value of the Modbus register (numeric).\n")
	fmt.Fprint(w, "# TYPE modbus_value gauge\n")
	fmt.Fprint(w, "# HELP modbus_sample_age_seconds Age in seconds since the last successful poll.\n")
	fmt.Fprint(w, "# TYPE modbus_sample_age_seconds gauge\n")
	fmt.Fprint(w, "# HELP modbus_register_info Info from UTF-8 string registers (string value in label).\n")
	fmt.Fprint(w, "# TYPE modbus_register_info gauge\n")

	for _, sm := range samples {
		if sm.StringValue != nil {
			// Registro UTF8: exponer como info (valor en etiqueta, gauge=1).
			fmt.Fprintf(
				w,
				"modbus_register_info{device=%q,slave=%q,register=%q,name=%q,unit=%q,ip_address=%q,value=%q} 1\n",
				sm.Device,
				fmt.Sprintf("%d", sm.SlaveID),
				fmt.Sprintf("%d", sm.Register),
				sm.Name,
				sm.Unit,
				sm.IpAddress,
				*sm.StringValue,
			)
		} else {
			// Registro numérico.
			fmt.Fprintf(
				w,
				"modbus_value{device=%q,slave=%q,register=%q,name=%q,unit=%q,ip_address=%q} %f\n",
				sm.Device,
				fmt.Sprintf("%d", sm.SlaveID),
				fmt.Sprintf("%d", sm.Register),
				sm.Name,
				sm.Unit,
				sm.IpAddress,
				sm.Value,
			)
		}
	}

	for _, sm := range samples {
		age := now.Sub(sm.Timestamp).Seconds()
		fmt.Fprintf(
			w,
			"modbus_sample_age_seconds{device=%q,slave=%q,register=%q,ip_address=%q} %f\n",
			sm.Device,
			fmt.Sprintf("%d", sm.SlaveID),
			fmt.Sprintf("%d", sm.Register),
			sm.IpAddress,
			age,
		)
	}
}
