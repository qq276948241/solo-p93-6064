package handler

import (
	"appliance-recycle/internal/dto"
	"appliance-recycle/internal/pkg/middleware"
	"appliance-recycle/internal/pkg/response"
	"appliance-recycle/internal/service"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ResidentHandler struct {
	AppointmentSvc *service.AppointmentService
}

func NewResidentHandler(svc *service.AppointmentService) *ResidentHandler {
	return &ResidentHandler{AppointmentSvc: svc}
}

func (h *ResidentHandler) CreateAppointment(c *gin.Context) {
	req, files, errCode := parseCreateAppointmentRequest(c)
	if errCode != 0 {
		c.JSON(http.StatusBadRequest, response.Fail(errCode))
		return
	}

	residentID := middleware.GetUserID(c)
	data, code, err := h.AppointmentSvc.CreateAppointment(residentID, req, files)
	if code != response.CodeSuccess {
		httpStatus := mapCodeToHTTP(code)
		c.JSON(httpStatus, response.Fail(code))
		_ = err
		return
	}
	_ = err
	c.JSON(http.StatusCreated, response.Success(data))
}

func parseCreateAppointmentRequest(c *gin.Context) (*dto.CreateAppointmentRequest, []*multipart.FileHeader, int) {
	contentType := c.ContentType()
	var req dto.CreateAppointmentRequest
	var files []*multipart.FileHeader

	if contentType == "application/json" || contentType == "" {
		if err := c.ShouldBindJSON(&req); err != nil {
			return nil, nil, response.CodeParamError
		}
		return &req, nil, 0
	}

	var form dto.CreateAppointmentForm
	if err := c.ShouldBind(&form); err != nil {
		return nil, nil, response.CodeParamError
	}
	req = dto.CreateAppointmentRequest{
		SlotID:          form.SlotID,
		Phone:           form.Phone,
		Address:         form.Address,
		ApplianceTypeID: form.ApplianceTypeID,
		ApplianceWeight: form.ApplianceWeight,
		Remark:          form.Remark,
	}

	formFiles, err := c.MultipartForm()
	if err == nil && formFiles != nil && len(formFiles.File["images"]) > 0 {
		files = formFiles.File["images"]
	}
	if len(files) == 0 {
		_, header, err := c.Request.FormFile("images")
		if err == nil && header != nil {
			files = []*multipart.FileHeader{header}
		}
	}

	return &req, files, 0
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
		httpStatus := mapCodeToHTTP(code)
		c.JSON(httpStatus, response.Fail(code))
		_ = err
		return
	}
	_ = err
	c.JSON(http.StatusOK, response.Success(nil))
}

func (h *ResidentHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.FailWithMsg(response.CodeParamError, err.Error()))
		return
	}

	code, err := service.RegisterResident(&req)
	if code != response.CodeSuccess {
		c.JSON(mapCodeToHTTP(code), response.Fail(code))
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
		c.JSON(mapCodeToHTTP(code), response.Fail(code))
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
		c.JSON(mapCodeToHTTP(code), response.Fail(code))
		return
	}
	_ = err
	c.JSON(http.StatusOK, response.Success(data))
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
