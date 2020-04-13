package main

// This is based on https://github.com/enricofoltran/simple-go-server
// and on https://github.com/mjibson/esc

// The following generates the cmd.FS http.FileSystem which embeds the webui/dist assets.
// The webui must be built first. Refer to the README.md in the webui folder.

//go:generate esc -o webui.go -pkg main -prefix ../../webui/dist ../../webui/dist

// From the command line, run "go generate ./cmd/srvils"

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/asgaut/dumpils/pkg/demod2"
)

type channelType struct {
	Name string  `json:"name"`
	LOC  float32 `json:"loc"`
	GP   float32 `json:"gp"`
}

type httpapi struct {
	commands   chan interface{}
	channel    channelType
	processors map[string]*processor
}

// ServeAPI serves webapi until the context is done
func (s *httpapi) ServeAPI(ctx context.Context, listenAddr string) error {
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)

	router := http.NewServeMux()
	router.Handle("/", http.FileServer(FS(false)))
	router.Handle("/spectrum", spectrum(s))
	router.Handle("/measurements", meas(s))
	router.Handle("/channel", channel(s))
	router.Handle("/samples", samples(s))

	server := &http.Server{
		Addr:         listenAddr,
		Handler:      tracing()(logging(logger)(router)),
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	done := make(chan bool)

	go func() {
		<-ctx.Done()
		logger.Println("Server is shutting down...")
		ctx2, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx2); err != nil {
			logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	logger.Println("Server is ready to handle requests at", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("could not listen on '%s': %v", listenAddr, err)
	}

	<-done
	logger.Println("Server stopped")
	return nil
}

func spectrum(s *httpapi) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: https://stackoverflow.com/questions/22972066/how-to-handle-preflight-cors-requests-on-a-go-server
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET") // POST, GET, OPTIONS, PUT, DELETE
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodGet {
			source, ok := r.URL.Query()["source"]
			if !ok || len(source) != 1 {
				http.Error(w, "'source' argument missing", http.StatusBadRequest)
				return
			}

			stage, ok := r.URL.Query()["stage"]
			if !ok || len(stage) != 1 {
				http.Error(w, "'stage' argument missing", http.StatusBadRequest)
				return
			}

			p, ok := s.processors[source[0]]
			if !ok {
				http.Error(w, fmt.Sprintf("'%s' input not defined", source[0]), http.StatusBadRequest)
				return
			}

			p.mu.Lock()
			defer p.mu.Unlock()
			var ret []float32
			if stage[0] == "if" {
				ret = p.demodulator.Spectrum1()
			} else {
				ret = p.demodulator.Spectrum2()
			}
			buf, err := json.Marshal(ret)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(buf)
		}
	})
}

func meas(s *httpapi) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := map[string]demod2.Meas{}
		for key, p := range s.processors {
			p.mu.Lock()
			data[key] = p.demodulator.Meas
			p.mu.Unlock()
		}
		buf, err := json.Marshal(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf)
	})
}

func channel(s *httpapi) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
		if r.Method == http.MethodPut {
			var newChannel channelType
			err := json.NewDecoder(r.Body).Decode(&newChannel)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			s.commands <- newChannel
			s.channel = newChannel
		}
		buf, err := json.Marshal(s.channel)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, PUT") // POST, GET, OPTIONS, PUT, DELETE
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf)
	})
}

func samples(s *httpapi) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, PUT") // POST, GET, OPTIONS, PUT, DELETE
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		source, ok := r.URL.Query()["source"]
		if !ok || len(source) != 1 {
			http.Error(w, "'source' argument missing", http.StatusBadRequest)
			return
		}
		p, ok := s.processors[source[0]]
		if !ok {
			http.Error(w, fmt.Sprintf("'%s' input not defined", source[0]), http.StatusBadRequest)
			return
		}
		p.mu.Lock()
		defer p.mu.Unlock()
		if r.Method == http.MethodGet {
			//w.Header().Set("Content-Type", "application/octet-binary")
		}
		if r.Method == http.MethodPut {
			if r.Header.Get("Content-Type") != "application/octet-binary" {
				http.Error(w, "Invalid Content-Type", http.StatusBadRequest)
				return
			}
			if r.ContentLength != int64(len(p.iqRawData)) {
				http.Error(w, fmt.Sprintf("Content-Length must be %d bytes", len(p.iqRawData)), http.StatusBadRequest)
				return
			}
			_, err := io.ReadFull(r.Body, p.iqRawData)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusAccepted)
		}
	})
}

func logging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			/*defer func() {
				logger.Println(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
			}()*/
			next.ServeHTTP(w, r)
		})
	}
}

func tracing() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}
