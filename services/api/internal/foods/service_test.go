package foods

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockFoodService struct{}

func (mockFoodService) ListGeneralFoods(q string) ([]FoodDTO, error) { return nil, nil }
func (mockFoodService) Get(id uint) (*FoodDTO, error)                { return nil, nil }
func (mockFoodService) Create(req UpsertRequest) (*FoodDTO, error) {
	return &FoodDTO{ID: 1, DictionaryItemID: req.DictionaryItemID, DictionaryItemLabel: "rice", Protein: req.Protein, Fat: req.Fat, Carbs: req.Carbs, Callories: req.Callories}, nil
}
func (mockFoodService) Update(id uint, req UpsertRequest) (*FoodDTO, error) { return nil, nil }
func (mockFoodService) Delete(id uint) error                                { return nil }

func TestCreateGeneralFoodSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/admin/general-foods", NewHandler(mockFoodService{}).CreateGeneralFood)
	req := httptest.NewRequest(http.MethodPost, "/admin/general-foods", bytes.NewBufferString(`{"dictionaryItemId":10,"protein":2.7,"fat":0.3,"carbs":28,"callories":1.3}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), `"dictionaryItemLabel":"rice"`)
}
