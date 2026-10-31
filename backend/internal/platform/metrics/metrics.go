// Package metrics 提供 Prometheus 指标。
package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics 指标集合。
type Metrics struct {
	Registry *prometheus.Registry
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
	inflight prometheus.Gauge
	dbUp     prometheus.Gauge
	redisUp  prometheus.Gauge
	dbConns  *prometheus.GaugeVec
	pool     *pgxpool.Pool
}

// New 创建指标集合。
func New() *Metrics {
	reg := prometheus.NewRegistry()
	m := &Metrics{
		Registry: reg,
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "wmc_http_requests_total",
			Help: "HTTP 请求总数",
		}, []string{"method", "route", "status"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "wmc_http_request_duration_seconds",
			Help:    "HTTP 请求耗时",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		}, []string{"method", "route"}),
		inflight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "wmc_http_inflight_requests",
			Help: "处理中的请求数",
		}),
		dbUp:    prometheus.NewGauge(prometheus.GaugeOpts{Name: "wmc_db_up", Help: "数据库可用性"}),
		redisUp: prometheus.NewGauge(prometheus.GaugeOpts{Name: "wmc_redis_up", Help: "Redis 可用性"}),
		dbConns: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "wmc_db_pool_conns",
			Help: "数据库连接池状态",
		}, []string{"state"}),
	}
	reg.MustRegister(m.requests, m.duration, m.inflight, m.dbUp, m.redisUp, m.dbConns)
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	return m
}

// AttachPool 绑定数据库连接池，用于采集连接数。
func (m *Metrics) AttachPool(pool *pgxpool.Pool) { m.pool = pool }

// Handler 暴露 /metrics。
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.Registry, promhttp.HandlerOpts{})
}

// SetChecks 更新健康状态指标。
func (m *Metrics) SetChecks(dbOK, redisOK bool) {
	if dbOK {
		m.dbUp.Set(1)
	} else {
		m.dbUp.Set(0)
	}
	if redisOK {
		m.redisUp.Set(1)
	} else {
		m.redisUp.Set(0)
	}
	if m.pool != nil {
		st := m.pool.Stat()
		m.dbConns.WithLabelValues("acquired").Set(float64(st.AcquiredConns()))
		m.dbConns.WithLabelValues("idle").Set(float64(st.IdleConns()))
		m.dbConns.WithLabelValues("total").Set(float64(st.TotalConns()))
		m.dbConns.WithLabelValues("max").Set(float64(st.MaxConns()))
	}
}

// Middleware 记录 HTTP 指标。
func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		m.inflight.Inc()
		defer m.inflight.Dec()
		rec := &statusRecorder{ResponseWriter: w, status: 200}
		next.ServeHTTP(rec, r)
		route := r.URL.Path
		m.requests.WithLabelValues(r.Method, route, strconv.Itoa(rec.status)).Inc()
		m.duration.WithLabelValues(r.Method, route).Observe(time.Since(start).Seconds())
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}
