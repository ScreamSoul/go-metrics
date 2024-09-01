package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/gojuno/minimock/v3"
	"github.com/screamsoul/go-metrics-tpl/internal/models/metrics"
	"github.com/screamsoul/go-metrics-tpl/internal/restapi/handlers"
	"github.com/screamsoul/go-metrics-tpl/internal/restapi/routers"
	"github.com/stretchr/testify/suite"
)

type MetricRouterSuite struct {
	suite.Suite
	server *httptest.Server
	mockDB *MetricStorageMock
}

func TestMemStorageSuite(t *testing.T) {
	suite.Run(t, new(MetricRouterSuite))
}

func (s *MetricRouterSuite) serverRequest(
	method,
	path string,
	body interface{},
	headers ...http.Header,
) *resty.Response {

	req := resty.New().R()

	req.Method = method
	req.URL = s.server.URL + path

	if body != nil {
		req.Body = body
	}
	for _, header := range headers {
		for key, values := range header {
			for _, value := range values {
				req.SetHeader(key, value)
			}
		}
	}

	resp, err := req.Send()
	s.Require().NoError(err, "error making HTTP request")

	return resp
}

func (s *MetricRouterSuite) SetupTest() {
	mc := minimock.NewController(s.T())

	s.mockDB = NewMetricStorageMock(mc)
	s.server = httptest.NewServer(
		routers.NewMetricRouter(
			handlers.NewMetricServer(s.mockDB),
		),
	)
}

func (s *MetricRouterSuite) TearDownTest() {
	s.server.Close()
	s.mockDB.MinimockFinish()
}

func (s *MetricRouterSuite) TestUpdateRouter() {

	s.mockDB.AddMock.Set(
		func(ctx context.Context, m metrics.Metrics) error {
			return nil
		},
	)

	var testTable = []struct {
		name   string
		body   interface{}
		method string
		status int
	}{
		{
			name:   "update ok",
			body:   map[string]interface{}{"type": "counter", "delta": 1, "ID": "someMetric1"},
			method: "POST",
			status: http.StatusOK,
		},
		{
			name:   "update mothod not allowed",
			body:   map[string]interface{}{"type": "counter", "delta": 1, "id": "someMetric1"},
			method: "GET",
			status: http.StatusMethodNotAllowed,
		},
		{
			name:   "update bad type",
			body:   map[string]interface{}{"type": "asd", "delta": 1, "id": "someMetric1"},
			method: "POST",
			status: http.StatusBadRequest,
		},
		{
			name:   "update bad value",
			body:   map[string]interface{}{"type": "counter", "delta": "asd", "id": "someMetric1"},
			method: "POST",
			status: http.StatusBadRequest,
		},
		{
			name:   "update incorrect value",
			body:   map[string]interface{}{"type": "counter", "value": 1.2, "id": "someMetric1"},
			method: "POST",
			status: http.StatusBadRequest,
		},
	}
	for _, v := range testTable {
		s.Suite.Run(v.name, func() {
			resp := s.serverRequest(v.method, "/update/", v.body, http.Header{"Content-Type": {"application/json"}})
			s.Equal(
				v.status,
				resp.StatusCode(),
				fmt.Sprintf("Resp body: %s", string(resp.Body())),
			)
		})
	}
}

func (s *MetricRouterSuite) TestUpdateBulkRouter() {

	mockBulkOk := func() {
		s.mockDB.BulkAddMock.Set(
			func(ctx context.Context, m []metrics.Metrics) error {
				return nil
			},
		)
	}

	mockBulkErr := func() {
		s.mockDB.BulkAddMock.Set(
			func(ctx context.Context, m []metrics.Metrics) error {
				return fmt.Errorf("some err")
			},
		)
	}

	bodyMore100Rows := make([]map[string]interface{}, 150)
	for i := range bodyMore100Rows {
		bodyMore100Rows[i] = map[string]interface{}{"type": "counter", "delta": 1, "id": "someMetric1"}
	}

	var testTable = []struct {
		name   string
		body   interface{}
		method string
		status int
		mock   func()
	}{
		{
			name: "update batch",
			body: []map[string]interface{}{
				{"type": "counter", "delta": 1, "id": "someMetric1"},
				{"type": "gauge", "value": 1.2, "id": "someMetric2"},
				{"type": "gauge", "value": 1.3, "id": "someMetric3"},
			},
			method: "POST",
			status: http.StatusOK,
			mock:   mockBulkOk,
		},
		{
			name:   "update batch",
			body:   map[string]interface{}{"type": "counter", "delta": 1, "id": "someMetric1"},
			method: "POST",
			status: http.StatusBadRequest,
			mock:   mockBulkOk,
		},
		{
			name: "update batch",
			body: []map[string]interface{}{
				{"type": "asd", "delta": 1, "id": "someMetric1"},
			},
			method: "POST",
			status: http.StatusBadRequest,
			mock:   mockBulkOk,
		},
		{
			name: "update batch",
			body: []map[string]interface{}{
				{"type": "counter", "value": 1.2, "id": "someMetric1"},
			},
			method: "POST",
			status: http.StatusBadRequest,
			mock:   mockBulkOk,
		},
		{
			name: "update batch",
			body: []map[string]interface{}{
				{"type": "counter", "delta": 1, "id": "someMetric1"},
			},
			method: "POST",
			status: http.StatusInternalServerError,
			mock:   mockBulkErr,
		},

		{
			name:   "update batch",
			body:   bodyMore100Rows,
			method: "POST",
			status: http.StatusInternalServerError,
			mock:   mockBulkErr,
		},
	}
	for _, v := range testTable {
		s.Suite.Run(v.name, func() {
			v.mock()
			resp := s.serverRequest(v.method, "/updates/", v.body, http.Header{"Content-Type": {"application/json"}})
			s.Equal(
				v.status,
				resp.StatusCode(),
				fmt.Sprintf("Resp body: %s", string(resp.Body())),
			)
		})
	}
}

func (s *MetricRouterSuite) TestUpdateFromPath() {
	s.mockDB.AddMock.Set(
		func(ctx context.Context, m metrics.Metrics) error {
			return nil
		},
	)

	var testTable = []struct {
		name   string
		path   string
		status int
	}{
		{
			name:   "status ok",
			path:   "/update/gauge/sume_metric1/1.2",
			status: http.StatusOK,
		},
		{
			name:   "bad type",
			path:   "/update/sume_gauge/sume_metric1/1.2",
			status: http.StatusBadRequest,
		},
		{
			name:   "bad value",
			path:   "/update/gauge/sume_metric1/asd",
			status: http.StatusBadRequest,
		},
	}

	for _, v := range testTable {
		s.Suite.Run(v.name, func() {
			resp := s.serverRequest("POST", v.path, nil)
			s.Equal(
				v.status,
				resp.StatusCode(),
				fmt.Sprintf("Resp body: %s", string(resp.Body())),
			)
		})
	}
}

func (s *MetricRouterSuite) TestGetMetricJSON() {

	s.mockDB.GetMock.Set(func(ctx context.Context, m *metrics.Metrics) (err error) {
		var intValue int64 = 1
		if m.MType == metrics.Counter && m.ID == "someMetric1" {
			m.Delta = &intValue
		} else {
			return errors.New("not found")
		}

		return nil
	})

	var testTable = []struct {
		name           string
		body           map[string]interface{}
		method         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "fail type",
			body:           map[string]interface{}{"type": "fake_type", "id": "someMetric1"},
			method:         "POST",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name:           "value ok",
			body:           map[string]interface{}{"type": "counter", "id": "someMetric1"},
			method:         "POST",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"type": "counter", "id": "someMetric1", "delta": 1}`,
		},
		{
			name:   "value not found",
			body:   map[string]interface{}{"type": "counter", "id": "someMetric2"},
			method: "POST",

			expectedStatus: http.StatusNotFound,
			expectedBody:   "",
		},
	}
	for _, v := range testTable {
		s.Suite.Run(v.name, func() {
			resp := s.serverRequest(v.method, "/value/", v.body, http.Header{"Content-Type": {"application/json"}})
			s.Require().Equal(v.expectedStatus, resp.StatusCode(), fmt.Sprintf("Resp body: %s", string(resp.Body())))
			if v.expectedBody != "" {
				s.JSONEq(v.expectedBody, string(resp.Body()))
			}
		})

	}

}

func (s *MetricRouterSuite) TestGetMetric() {

	s.mockDB.GetMock.Set(func(ctx context.Context, m *metrics.Metrics) (err error) {
		var intValue int64 = 1
		if m.MType == metrics.Counter && m.ID == "someMetric1" {
			m.Delta = &intValue
		} else {
			return errors.New("not found")
		}

		return nil
	})

	var testTable = []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "fail type",
			body:           map[string]interface{}{"type": "fake_type", "id": "someMetric1"},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name:           "value ok",
			body:           map[string]interface{}{"type": "counter", "id": "someMetric1"},
			expectedStatus: http.StatusOK,
			expectedBody:   "1",
		},
		{
			name:           "value not found",
			body:           map[string]interface{}{"type": "counter", "id": "someMetric2"},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "",
		},
	}
	for _, v := range testTable {
		s.Suite.Run(v.name, func() {
			path := fmt.Sprintf("/value/%s/%s", v.body["type"], v.body["id"])
			resp := s.serverRequest("GET", path, nil)
			s.Require().Equal(v.expectedStatus, resp.StatusCode(), fmt.Sprintf("Resp body: %s", string(resp.Body())))
			if v.expectedBody != "" {
				s.Equal(v.expectedBody, string(resp.Body()))
			}
		})

	}

}

func (s *MetricRouterSuite) TestListMetrics() {

	var intValue int64 = 1
	var floatValue = 1.1

	mockMetricList := []metrics.Metrics{
		{ID: "MetricGauge1", Value: &floatValue, MType: metrics.Gauge},
		{ID: "MetricCounter1", Delta: &intValue, MType: metrics.Counter},
	}

	mockWithDataList := func() {
		s.mockDB.ListMock.Return(
			mockMetricList, nil,
		)
	}

	mockWithErr := func() {
		mockMetricList = []metrics.Metrics{}
		s.mockDB.ListMock.Return(
			mockMetricList, fmt.Errorf("some err"),
		)
	}

	var testTable = []struct {
		name   string
		method string
		status int
		result []metrics.Metrics
		mock   func()
	}{
		{"list json ok", "GET", http.StatusOK, mockMetricList, mockWithDataList},
		{"list json err", "GET", http.StatusInternalServerError, mockMetricList, mockWithErr},
	}
	for _, v := range testTable {
		s.Suite.Run(v.name, func() {
			v.mock()
			resp := s.serverRequest(v.method, "/", nil)
			s.Require().Equal(v.status, resp.StatusCode())
			if v.status == http.StatusOK {
				var mList []metrics.Metrics
				err := json.Unmarshal(resp.Body(), &mList)
				s.Require().NoError(err, string(resp.Body()))
				s.Equal(v.result, mList)
			}
		})

	}

}

func (s *MetricRouterSuite) TestPingStorage() {

	var testTable = []struct {
		name   string
		method string
		status int
		mock   func()
	}{
		{
			name:   "connect db ok",
			method: "GET",
			status: http.StatusOK,
			mock: func() {
				s.mockDB.PingMock.Return(true)
			},
		},
		{
			name:   "connect db fail",
			method: "GET",
			status: http.StatusInternalServerError,
			mock: func() {
				s.mockDB.PingMock.Return(false)
			},
		},
	}
	for _, v := range testTable {
		s.Suite.Run(v.name, func() {
			v.mock()
			resp := s.serverRequest(v.method, "/ping", nil)
			s.Require().Equal(v.status, resp.StatusCode())
		})

	}
}
