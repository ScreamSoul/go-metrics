package routers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/screamsoul/go-metrics-tpl/internal/restapi/handlers"
)

func NewMetricRouter(
	mServer *handlers.MetricServer,
	middlewares ...func(http.Handler) http.Handler,
) chi.Router {

	r := chi.NewRouter()

	r.Use(middlewares...)

	r.Get("/", mServer.ListMetrics)
	r.Get("/ping", mServer.PingStorage)
	r.Post("/value/", mServer.GetMetricJSON)
	r.Get("/value/{metric_type}/{metric_name}", mServer.GetMetricValue)
	r.Post("/update/", mServer.UpdateMetric)
	r.Post("/updates/", mServer.UpdateMetricBulk)
	r.Post("/update/{metric_type}/{metric_name}/{metric_value}", mServer.UpdateMetric)

	return r
}
