package handlers

import (
	"forum/ui"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"path/filepath"
	"time"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path", "method"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration)
}

func (h *handler) Routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(neuteredFileSystem{http.FS(ui.Files)})
	mux.Handle("/static", http.NotFoundHandler())
	mux.Handle("/static/", fileServer)

	mux.HandleFunc("/", h.checkCookie(h.home))
	mux.HandleFunc("/post/", h.checkCookie(h.postView))
	mux.HandleFunc("/post/create", h.requireAuthentication(h.postCreate))
	mux.HandleFunc("/login", h.notRegistered(h.login))
	mux.HandleFunc("/signup", h.notRegistered(h.signup))
	mux.HandleFunc("/logout", h.requireAuthentication(h.logoutPost))
	mux.HandleFunc("/user/posts", h.requireAuthentication(h.PostByUser))
	mux.HandleFunc("/user/liked", h.requireAuthentication(h.LikedPosts))
	mux.HandleFunc("/post/reaction", h.requireAuthentication(h.postReaction))
	mux.HandleFunc("/comment/post", h.requireAuthentication(h.commentPost))
	mux.HandleFunc("/comment/reaction", h.requireAuthentication(h.commentReaction))

	// Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Apply the Prometheus middleware
	return prometheusMiddleware(h.secureHeaders(mux))
}

type neuteredFileSystem struct {
	fs http.FileSystem
}

func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}
	s, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := nfs.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}
			return nil, err
		}

	}

	return f, nil
}

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := r.URL.Path
		method := r.Method

		next.ServeHTTP(w, r)

		duration := time.Since(start).Seconds()
		httpRequestsTotal.WithLabelValues(path, method).Inc()
		httpRequestDuration.WithLabelValues(path, method).Observe(duration)
	})
}
