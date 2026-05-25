package metriclinks

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"activitydiary/api/internal/common"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockMetricLinkService struct{}

func (mockMetricLinkService) ListUnits(metricNameID uint, page, limit int) (*common.PageResponse[LinkResponse], error) {
	return &common.PageResponse[LinkResponse]{Items: []LinkResponse{{ID: 2, Label: "reps"}}, Page: 0, Limit: 10, TotalElements: 1, TotalPages: 1}, nil
}
func (mockMetricLinkService) Create(req LinkRequest) (*LinkResponse, error) {
	return &LinkResponse{ID: req.MetricUnitID, Label: "reps"}, nil
}
func (mockMetricLinkService) Delete(metricNameID, metricUnitID uint) error { return nil }

func TestCreateMetricLinkSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/admin/metric-links", NewHandler(mockMetricLinkService{}).Create)
	req := httptest.NewRequest(http.MethodPost, "/admin/metric-links", bytes.NewBufferString(`{"metricNameId":1,"metricUnitId":2}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), `"label":"reps"`)
}
