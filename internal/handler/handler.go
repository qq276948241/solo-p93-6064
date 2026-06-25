package handler

import (
	"appliance-recycle/internal/dto"
	"appliance-recycle/internal/pkg/middleware"
	"appliance-recycle/internal/pkg/response"
	"appliance-recycle/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ResidentHandler struct{}

func NewResidentHandler() *ResidentHandler {
	return &ResidentHandler{}
}

func (h *ResidentHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.FailWithMsg(response.CodeParamError, err.Error()))
		return
	}

	code, err := service.RegisterResident(&req)
	if code != response.CodeSuccess {
		httpStatus := http.StatusBadRequest
		if code == response.CodeDBError || code == response.CodeServerError {
			httpStatus = http.StatusInternalServerError
		}
		c.JSON(httpStatus, response.Fail(code))
		return
	}
	_ = err
	c.JSON(http.StatusOK, response.Success(nil))
}

func (h *ResidentHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.FailWithMsg(response.CodeParamError, err.Error()))
		return
	}

	token, code, err := service.ResidentLogin(&req)
	if code != response.CodeSuccess {
		httpStatus := http.StatusBadRequest
		if code == response.CodeDBError || code == response.CodeServerError {
			httpStatus = http.StatusInternalServerError
		}
		if code == response.CodeUserNotFound {
			httpStatus = http.StatusNotFound
		}
		c.JSON(httpStatus, response.Fail(code))
		return
	}
	_ = err
	c.JSON(http.StatusOK, response.Success(dto.LoginResponse{Token: token}))
}

func (h *ResidentHandler) GetSlots(c *gin.Context) {
	var req dto.SlotListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.FailWithMsg(response.CodeParamError, err.Error()))
		return
	}

	data, code, err := service.GetWeekSlots(req.WeekStart)
	if code != response.CodeSuccess {
		httpStatus := http.StatusInternalServerError
		if code == response.CodeParamError {
			httpStatus = http.StatusBadRequest
		}
		c.JSON(httpStatus, response.Fail(code))
		return
	}
	_ = err
	c.JSON(http.StatusOK, response.Success(data))
}

func (h *ResidentHandler) CreateAppointment(c *gin.Context) {
	var req dto.CreateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.FailWithMsg(response.CodeParamError, err.Error()))
		return
	}

	residentID := middleware.GetUserID(c)
	data, code, err := service.CreateAppointment(residentID, &req)
	if code != response.CodeSuccess {
		httpStatus := http.StatusBadRequest
		if code == response.CodeDBError || code == response.CodeServerError {
			httpStatus = http.StatusInternalServerError
		}
		if code == response.CodeSlotNotFound || code == response.CodeApplianceInvalid {
			httpStatus = http.StatusNotFound
		}
		c.JSON(httpStatus, response.Fail(code))
		return
	}
	_ = err
	c.JSON(http.StatusCreated, response.Success(data))
}

func (h *ResidentHandler) MyAppointments(c *gin.Context) {
	residentID := middleware.GetUserID(c)
	data, code, err := service.GetResidentAppointments(residentID)
	if code != response.CodeSuccess {
		c.JSON(http.StatusInternalServerError, response.Fail(code))
		return
	}
	_ = err
	c.JSON(http.StatusOK, response.Success(data))
}

func (h *ResidentHandler) CancelAppointment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, response.Fail(response.CodeParamError))
		return
	}

	var req dto.CancelAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Remark = ""
	}

	residentID := middleware.GetUserID(c)
	code, err := service.CancelAppointment(residentID, id, &req)
	if code != response.CodeSuccess {
		httpStatus := http.StatusBadRequest
		if code == response.CodeDBError {
			httpStatus = http.StatusInternalServerError
		}
		if code == response.CodeAppointmentNotFound {
			httpStatus = http.StatusNotFound
		}
		c.JSON(httpStatus, response.Fail(code))
		return
	}
	_ = err
	c.JSON(http.StatusOK, response.Success(nil))
}

func (h *ResidentHandler) GetApplianceTypes(c *gin.Context) {
	data, code, err := service.GetApplianceTypes()
	if code != response.CodeSuccess {
		c.JSON(http.StatusInternalServerError, response.Fail(code))
		return
	}
	_ = err
	c.JSON(http.StatusOK, response.Success(data))
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
		httpStatus := http.StatusBadRequest
		if code == response.CodeDBError || code == response.CodeServerError {
			httpStatus = http.StatusInternalServerError
		}
		if code == response.CodeUserNotFound {
			httpStatus = http.StatusNotFound
		}
		c.JSON(httpStatus, response.Fail(code))
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
		httpStatus := http.StatusInternalServerError
		if code == response.CodeParamError {
			httpStatus = http.StatusBadRequest
		}
		c.JSON(httpStatus, response.Fail(code))
		return
	}
	_ = err
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
		httpStatus := http.StatusBadRequest
		if code == response.CodeDBError {
			httpStatus = http.StatusInternalServerError
		}
		if code == response.CodeAppointmentNotFound {
			httpStatus = http.StatusNotFound
		}
		c.JSON(httpStatus, response.Fail(code))
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
		httpStatus := http.StatusInternalServerError
		if code == response.CodeParamError {
			httpStatus = http.StatusBadRequest
		}
		c.JSON(httpStatus, response.Fail(code))
		return
	}
	_ = err
	c.JSON(http.StatusOK, response.Success(data))
}
