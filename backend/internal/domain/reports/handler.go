package reports

import (
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thoriqzs/PARKIR/backend/internal/middleware"
	"github.com/thoriqzs/PARKIR/backend/internal/response"
	"github.com/thoriqzs/PARKIR/backend/internal/store"
)

type Handler struct {
	store *store.Store
}

func NewHandler(store *store.Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	reports := r.Group("/reports")
	{
		reports.GET("/daily-revenue", middleware.RequirePermission("reports:view_revenue"), h.DailyRevenue)
		reports.GET("/occupancy", middleware.RequirePermission("reports:view_occupancy"), h.Occupancy)
		reports.GET("/vehicle-breakdown", middleware.RequirePermission("reports:view_revenue"), h.VehicleBreakdown)
		reports.GET("/operator-activity", middleware.RequirePermission("reports:view_operators"), h.OperatorActivity)
	}
}

func (h *Handler) parseDateRange(c *gin.Context) store.DateRange {
	now := time.Now().UTC()
	dateTo := now
	dateFrom := now.AddDate(0, 0, -7)

	if v := c.Query("date_from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			dateFrom = t
		}
	}
	if v := c.Query("date_to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			dateTo = t.Add(24*time.Hour - time.Second)
		}
	}

	if dateTo.Sub(dateFrom) > 90*24*time.Hour {
		dateFrom = dateTo.AddDate(0, 0, -90)
	}

	return store.DateRange{DateFrom: dateFrom, DateTo: dateTo}
}

func (h *Handler) DailyRevenue(c *gin.Context) {
	locationID := c.Query("location_id")
	if locationID == "" {
		response.BadRequest(c, "MISSING_LOCATION", "location_id is required")
		return
	}
	dr := h.parseDateRange(c)
	includeVoided := c.Query("include_voided") == "true"
	format := c.DefaultQuery("format", "json")

	data, err := h.store.ReportDailyRevenue(c.Request.Context(), locationID, dr, includeVoided)
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	if format == "csv" {
		h.exportDailyRevenueCSV(c, data)
		return
	}

	type rowJSON struct {
		Date             string  `json:"date"`
		TotalRevenue     float64 `json:"total_revenue"`
		TransactionCount int     `json:"transaction_count"`
		AverageFee       float64 `json:"average_fee"`
		VoidedCount      int     `json:"voided_count"`
		VoidedAmount     float64 `json:"voided_amount"`
	}
	jsonData := make([]rowJSON, len(data))
	for i, r := range data {
		jsonData[i] = rowJSON{
			Date:             r.Date.Format("2006-01-02"),
			TotalRevenue:     r.TotalRevenue,
			TransactionCount: r.TransactionCount,
			AverageFee:       r.AverageFee,
			VoidedCount:      r.VoidedCount,
			VoidedAmount:     r.VoidedAmount,
		}
	}

	response.OK(c, jsonData)
}

func (h *Handler) Occupancy(c *gin.Context) {
	locationID := c.Query("location_id")
	if locationID == "" {
		response.BadRequest(c, "MISSING_LOCATION", "location_id is required")
		return
	}
	dr := h.parseDateRange(c)
	granularity := c.DefaultQuery("granularity", "day")
	format := c.DefaultQuery("format", "json")

	data, err := h.store.ReportOccupancy(c.Request.Context(), locationID, dr, granularity)
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	if format == "csv" {
		h.exportOccupancyCSV(c, data)
		return
	}

	response.OK(c, data)
}

func (h *Handler) VehicleBreakdown(c *gin.Context) {
	locationID := c.Query("location_id")
	if locationID == "" {
		response.BadRequest(c, "MISSING_LOCATION", "location_id is required")
		return
	}
	dr := h.parseDateRange(c)
	format := c.DefaultQuery("format", "json")

	data, err := h.store.ReportVehicleBreakdown(c.Request.Context(), locationID, dr)
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	if format == "csv" {
		h.exportVehicleBreakdownCSV(c, data)
		return
	}

	response.OK(c, data)
}

func (h *Handler) OperatorActivity(c *gin.Context) {
	locationID := c.Query("location_id")
	if locationID == "" {
		response.BadRequest(c, "MISSING_LOCATION", "location_id is required")
		return
	}
	dr := h.parseDateRange(c)
	operatorID := c.Query("operator_id")
	format := c.DefaultQuery("format", "json")

	data, err := h.store.ReportOperatorActivity(c.Request.Context(), locationID, dr, operatorID)
	if err != nil {
		_ = c.Error(err)
		response.InternalServerError(c)
		return
	}

	if format == "csv" {
		h.exportOperatorActivityCSV(c, data)
		return
	}

	response.OK(c, data)
}

func (h *Handler) exportDailyRevenueCSV(c *gin.Context, rows []store.DailyRevenueRow) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=daily_revenue.csv")
	w := csv.NewWriter(c.Writer)
	w.Write([]string{"Date", "Total Revenue", "Transaction Count", "Average Fee", "Voided Count", "Voided Amount"})
	for _, r := range rows {
		w.Write([]string{
			r.Date.Format("2006-01-02"),
			fmt.Sprintf("%.2f", r.TotalRevenue),
			fmt.Sprintf("%d", r.TransactionCount),
			fmt.Sprintf("%.2f", r.AverageFee),
			fmt.Sprintf("%d", r.VoidedCount),
			fmt.Sprintf("%.2f", r.VoidedAmount),
		})
	}
	w.Flush()
}

func (h *Handler) exportOccupancyCSV(c *gin.Context, rows []store.OccupancyRow) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=occupancy.csv")
	w := csv.NewWriter(c.Writer)
	w.Write([]string{"Bucket", "Count"})
	for _, r := range rows {
		w.Write([]string{r.Bucket.Format(time.RFC3339), fmt.Sprintf("%d", r.Count)})
	}
	w.Flush()
}

func (h *Handler) exportVehicleBreakdownCSV(c *gin.Context, rows []store.VehicleBreakdownRow) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=vehicle_breakdown.csv")
	w := csv.NewWriter(c.Writer)
	w.Write([]string{"Vehicle Type", "Count", "Total Revenue"})
	for _, r := range rows {
		w.Write([]string{r.VehicleType, fmt.Sprintf("%d", r.Count), fmt.Sprintf("%.2f", r.TotalRevenue)})
	}
	w.Flush()
}

func (h *Handler) exportOperatorActivityCSV(c *gin.Context, rows []store.OperatorActivityRow) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=operator_activity.csv")
	w := csv.NewWriter(c.Writer)
	w.Write([]string{"Operator ID", "Operator Name", "Session Count", "Total Revenue", "Shift Hours"})
	for _, r := range rows {
		w.Write([]string{
			r.OperatorID,
			r.OperatorName,
			fmt.Sprintf("%d", r.SessionCount),
			fmt.Sprintf("%.2f", r.TotalRevenue),
			strings.TrimRight(fmt.Sprintf("%.2f", r.ShiftHours), "0"),
		})
	}
	w.Flush()
}
