package handler

import (
	"appliance-recycle/internal/dto"
	"appliance-recycle/internal/pkg/response"
	"appliance-recycle/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func mapCodeToHTTP(code int) int {
	switch code {
	case response.CodeParamError, response.CodeImageTooMany, response.CodeImageTooFew,
		response.CodeImageInvalidExt, response.CodeImageTooLarge,
		response.CodeSlotFull, response.CodeSlotPast, response.CodeAppointmentDone,
		response.CodeAppointmentCanceled, response.CodeStatusInvalid,
		response.CodeUserExist, response.CodePasswordWrong:
		return http.StatusBadRequest
	case response.CodeUnauthorized, response.CodeTokenInvalid, response.CodeTokenExpired, response.CodeTokenMalformed:
		return http.StatusUnauthorized
	case response.CodeForbidden:
		return http.StatusForbidden
	case response.CodeNotFound, response.CodeSlotNotFound, response.CodeApplianceInvalid,
		response.CodeAppointmentNotFound, response.CodeUserNotFound:
		return http.StatusNotFound
	case response.CodeDBError, response.CodeServerError, response.CodeImageSaveFailed:
		return http.StatusInternalServerError
	default:
		return http.StatusBadRequest
	}
}

type AdminHandler struct{}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

func (h *AdminHandler) Login(c *gin.Context) {
	var req dto.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.FailWithMsg(response.CodeParamError, err.Error()))
		return
	}

	token, code, err := service.AdminLogin(&req)
	if code != response.CodeSuccess {
		c.JSON(mapCodeToHTTP(code), response.Fail(code))
		return
	}
	_ = err
	c.JSON(http.StatusOK, response.Success(dto.LoginResponse{Token: token}))
}

func (h *AdminHandler) ListAppointments(c *gin.Context) {
	var req dto.AdminAppointmentListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.FailWithMsg(response.CodeParamError, err.Error()))
		return
	}

	data, code, err := service.GetAdminAppointments(&req)
	if code != response.CodeSuccess {
		c.JSON(mapCodeToHTTP(code), response.Fail(code))
		return
	}
	_ = err
	c.JSON(http.StatusOK, response.Success(data))
}

func (h *AdminHandler) GetAppointmentDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, response.Fail(response.CodeParamError))
		return
	}
	data, code, err := service.GetAppointmentDetail(id)
	if code != response.CodeSuccess {
		c.JSON(mapCodeToHTTP(code), response.Fail(code))
		_ = err
		return
	}
	c.JSON(http.StatusOK, response.Success(data))
}

func (h *AdminHandler) UpdateAppointmentStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, response.Fail(response.CodeParamError))
		return
	}

	var req dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.FailWithMsg(response.CodeParamError, err.Error()))
		return
	}

	code, err := service.UpdateAppointmentStatus(id, &req)
	if code != response.CodeSuccess {
		c.JSON(mapCodeToHTTP(code), response.Fail(code))
		return
	}
	_ = err
	c.JSON(http.StatusOK, response.Success(nil))
}

func (h *AdminHandler) Statistics(c *gin.Context) {
	data, code, err := service.GetMonthlyStatistics()
	if code != response.CodeSuccess {
		c.JSON(http.StatusInternalServerError, response.Fail(code))
		return
	}
	_ = err
	c.JSON(http.StatusOK, response.Success(data))
}

func (h *AdminHandler) GetApplianceTypes(c *gin.Context) {
	data, code, err := service.GetApplianceTypes()
	if code != response.CodeSuccess {
		c.JSON(http.StatusInternalServerError, response.Fail(code))
		return
	}
	_ = err
	c.JSON(http.StatusOK, response.Success(data))
}

func (h *AdminHandler) GetSlots(c *gin.Context) {
	var req dto.SlotListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.FailWithMsg(response.CodeParamError, err.Error()))
		return
	}

	data, code, err := service.GetWeekSlots(req.WeekStart)
	if code != response.CodeSuccess {
		c.JSON(mapCodeToHTTP(code), response.Fail(code))
		return
	}
	_ = err
	c.JSON(http.StatusOK, response.Success(data))
}
