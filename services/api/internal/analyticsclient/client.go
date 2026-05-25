package analyticsclient

import (
	"net/http"
	"strconv"

	"activitydiary/api/internal/auth"
	"activitydiary/api/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

type Client struct {
	http *resty.Client
}

func New(baseURL string) *Client {
	client := resty.New().SetBaseURL(baseURL)
	return &Client{http: client}
}

func (c *Client) GetChartTypes(tagID uint) ([]string, error) {
	var data []string
	_, err := c.http.R().SetResult(&data).Get("/internal/tags/" + strconv.Itoa(int(tagID)) + "/chart-types")
	return data, err
}

func (c *Client) GetChart(userID uint, query map[string]string) (map[string]interface{}, error) {
	var data map[string]interface{}
	req := c.http.R().SetHeader("X-User-ID", strconv.Itoa(int(userID))).SetResult(&data)
	for key, value := range query {
		req.SetQueryParam(key, value)
	}
	_, err := req.Get("/internal/charts")
	return data, err
}

type Handler struct {
	client *Client
}

func NewHandler(client *Client) *Handler {
	return &Handler{client: client}
}

func (h *Handler) GetChartTypes(c *gin.Context) {
	items, err := h.client.GetChartTypes(uint(parseInt(c.Param("tagId"), 0)))
	if err != nil {
		response.Error(c, http.StatusBadGateway, "Failed to load analytics chart types")
		return
	}
	response.OK(c, http.StatusOK, items)
}

func (h *Handler) GetChart(c *gin.Context) {
	query := map[string]string{
		"tagId":     c.Query("tagId"),
		"chartType": c.Query("chartType"),
		"dateFrom":  c.Query("dateFrom"),
		"dateTo":    c.Query("dateTo"),
	}
	data, err := h.client.GetChart(auth.UserIDFromContext(c), query)
	if err != nil {
		response.Error(c, http.StatusBadGateway, "Failed to load analytics chart")
		return
	}
	response.OK(c, http.StatusOK, data)
}

func parseInt(raw string, fallback int) int {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}
