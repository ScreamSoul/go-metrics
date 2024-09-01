package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/screamsoul/go-metrics-tpl/internal/models/metrics"
	"github.com/screamsoul/go-metrics-tpl/internal/repositories"
	"github.com/screamsoul/go-metrics-tpl/pkg/logging"
	"go.uber.org/zap"
)

type MetricServer struct {
	store  repositories.MetricStorage
	logger *zap.Logger
}

func NewMetricServer(metricRepo repositories.MetricStorage) *MetricServer {
	logger := logging.GetLogger()

	return &MetricServer{store: metricRepo, logger: logger}
}

// PingStorage checks the connection to the database.
func (ms *MetricServer) PingStorage(w http.ResponseWriter, r *http.Request) {
	if !ms.store.Ping(r.Context()) {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

// UpdateMetricBulk handler, updates metrics, can accept multiple metrics in json format at once.
func (ms *MetricServer) UpdateMetricBulk(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(w, "content type must be application/json", http.StatusBadRequest)
		return
	}

	var metricsListChunk []metrics.Metrics

	decoder := json.NewDecoder(r.Body)

	if _, err := decoder.Token(); err != nil {
		http.Error(w, "bad json body", http.StatusBadRequest)
		return
	}

	for decoder.More() {
		var currentMetric metrics.Metrics
		if err := decoder.Decode(&currentMetric); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := currentMetric.ValidateValue(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		metricsListChunk = append(metricsListChunk, currentMetric)
		if len(metricsListChunk) == 100 {
			if err := ms.store.BulkAdd(r.Context(), metricsListChunk); err != nil {
				ms.logger.Error("Error update metrics chunk", zap.Error(err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			metricsListChunk = metricsListChunk[:0]
		}
	}
	if len(metricsListChunk) > 0 {
		if err := ms.store.BulkAdd(r.Context(), metricsListChunk); err != nil {
			ms.logger.Error("Error update metrics chunk", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	if _, err := decoder.Token(); err != nil {
		http.Error(w, "bad json body", http.StatusBadRequest)
		return
	}
}

// UpdateMetric handler, updates one metric.
func (ms *MetricServer) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	var metricObj metrics.Metrics

	contentType := r.Header.Get("Content-Type")
	if contentType == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&metricObj); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		metriObjPoint, err := metrics.NewMetric(
			r.PathValue("metric_type"),
			r.PathValue("metric_name"),
			r.PathValue("metric_value"),
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		metricObj = *metriObjPoint
	}

	if err := metricObj.ValidateValue(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := ms.store.Add(r.Context(), metricObj); err != nil {
		ms.logger.Error("Error update metric", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetMetricValue handler, returns the metric value by type and name.
func (ms *MetricServer) GetMetricValue(w http.ResponseWriter, r *http.Request) {

	metricObj, err := metrics.NewMetric(
		r.PathValue("metric_type"),
		r.PathValue("metric_name"),
		"",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = ms.store.Get(r.Context(), metricObj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if _, err := w.Write([]byte(metricObj.GetValue())); err != nil {
		ms.logger.Error("Error writing response", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// GetMetricJSON handler, returns the metric value by type and name in the josn format.
func (ms *MetricServer) GetMetricJSON(w http.ResponseWriter, r *http.Request) {

	var metricObj metrics.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metricObj); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := ms.store.Get(r.Context(), &metricObj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(&metricObj); err != nil {
		ms.logger.Error("Error writing response", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ListMetrics handler, returns all current metrics
func (ms *MetricServer) ListMetrics(w http.ResponseWriter, r *http.Request) {

	metrics, err := ms.store.List(r.Context())

	if err != nil {
		ms.logger.Error("error read metrics", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		ms.logger.Error("Error writing response", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
