package tags

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"activitydiary/api/internal/common"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockTagService struct{}

func (mockTagService) ListPublic(q string) ([]TagDTO, error) {
	return []TagDTO{{ID: 1, Name: "training", Status: "APPROVED"}}, nil
}
func (mockTagService) ListMetricsByTags(tagIDs []uint, page, limit int, q string) (*common.PageResponse[map[string]interface{}], error) {
	return &common.PageResponse[map[string]interface{}]{Items: []map[string]interface{}{}, Page: 0, Limit: 10}, nil
}
func (mockTagService) ListAdmin(page, size int, q string) (*common.Slice[TagDTO], error) {
	return &common.Slice[TagDTO]{Content: []TagDTO{}, Number: 0, Size: size, First: true, Last: true, Empty: true}, nil
}
func (mockTagService) Create(name string) (*TagDTO, error) {
	return &TagDTO{ID: 1, Name: name, Status: "PROPOSED"}, nil
}
func (mockTagService) Update(id uint, name string) (*TagDTO, error)      { return &TagDTO{}, nil }
func (mockTagService) SetStatus(id uint, status string) error            { return nil }
func (mockTagService) Delete(id uint) error                              { return nil }
func (mockTagService) ListTagMetrics(tagID uint) ([]TagMetricDTO, error) { return nil, nil }
func (mockTagService) ReplaceTagMetrics(tagID uint, metricNameIDs []uint) ([]TagMetricDTO, error) {
	return nil, nil
}

func TestGetTagsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/tags", NewHandler(mockTagService{}).ListPublic)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/tags?q=training", nil)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"name":"training"`)
}
