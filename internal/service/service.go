package service

import (
	"appliance-recycle/internal/config"
	"appliance-recycle/internal/dto"
	"appliance-recycle/internal/model"
	"appliance-recycle/internal/pkg/database"
	"appliance-recycle/internal/pkg/jwt"
	"appliance-recycle/internal/pkg/response"
	"appliance-recycle/internal/pkg/upload_helper"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AppointmentService struct {
	storage upload_helper.Storage
	upload  config.UploadConfig
}

func NewAppointmentService(storage upload_helper.Storage, upload config.UploadConfig) *AppointmentService {
	return &AppointmentService{storage: storage, upload: upload}
}

func (s *AppointmentService) ValidateAndSaveImages(files []*multipart.FileHeader) ([]string, int, error) {
	if len(files) < s.upload.MinCount {
		return nil, response.CodeImageTooFew, errors.New("too few images")
	}
	if len(files) > s.upload.MaxCount {
		return nil, response.CodeImageTooMany, errors.New("too many images")
	}

	allowedExt := s.upload.ExtSet()
	for _, fh := range files {
		if fh.Size > s.upload.MaxSize {
			return nil, response.CodeImageTooLarge, errors.New("image too large")
		}
		ext := strings.ToLower(filepath.Ext(fh.Filename))
		if !allowedExt[ext] {
			return nil, response.CodeImageInvalidExt, errors.New("invalid extension")
		}
	}

	imageURLs := make([]string, 0, len(files))
	for _, fh := range files {
		url, err := s.storage.Save(fh)
		if err != nil {
			return nil, response.CodeImageSaveFailed, err
		}
		imageURLs = append(imageURLs, url)
	}
	return imageURLs, response.CodeSuccess, nil
}

func (s *AppointmentService) CreateAppointment(residentID uint64, req *dto.CreateAppointmentRequest, files []*multipart.FileHeader) (*dto.CreateAppointmentResponse, int, error) {
	var applianceType model.ApplianceType
	if err := database.DB.Where("id = ?", req.ApplianceTypeID).First(&applianceType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.CodeApplianceInvalid, err
		}
		return nil, response.CodeDBError, err
	}

	var imageURLs []string
	if len(files) > 0 {
		urls, code, err := s.ValidateAndSaveImages(files)
		if code != response.CodeSuccess {
			return nil, code, err
		}
		imageURLs = urls
	} else if len(req.Images) > 0 {
		imageURLs = req.Images
	}

	tx := database.DB.Begin()

	var slot model.TimeSlot
	if err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("id = ?", req.SlotID).First(&slot).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.CodeSlotNotFound, err
		}
		return nil, response.CodeDBError, err
	}

	slotEndTime, _ := time.ParseInLocation("2006-01-02 15:04:05",
		slot.SlotDate+" "+slot.EndTime+":00", time.Local)
	if slotEndTime.Before(time.Now()) {
		tx.Rollback()
		return nil, response.CodeSlotPast, errors.New("slot past")
	}

	if slot.BookedCount >= slot.Capacity {
		tx.Rollback()
		return nil, response.CodeSlotFull, errors.New("slot full")
	}

	orderNo := generateOrderNo()

	appointment := model.Appointment{
		OrderNo:         orderNo,
		ResidentID:      residentID,
		SlotID:          req.SlotID,
		Phone:           req.Phone,
		Address:         req.Address,
		ApplianceTypeID: req.ApplianceTypeID,
		ApplianceWeight: req.ApplianceWeight,
		Status:          model.AppointmentStatusPending,
		Images:          model.StringArr(imageURLs),
		Remark:          req.Remark,
	}

	if err := tx.Create(&appointment).Error; err != nil {
		tx.Rollback()
		return nil, response.CodeDBError, err
	}

	if err := tx.Model(&slot).Update("booked_count", slot.BookedCount+1).Error; err != nil {
		tx.Rollback()
		return nil, response.CodeDBError, err
	}

	tx.Commit()

	return &dto.CreateAppointmentResponse{
		OrderNo: orderNo,
		ID:      appointment.ID,
	}, response.CodeSuccess, nil
}

func RegisterResident(req *dto.RegisterRequest) (int, error) {
	var count int64
	if err := database.DB.Model(&model.Resident{}).Where("phone = ?", req.Phone).Count(&count).Error; err != nil {
		return response.CodeDBError, err
	}
	if count > 0 {
		return response.CodeUserExist, errors.New("user exist")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return response.CodeServerError, err
	}

	resident := model.Resident{
		Phone:    req.Phone,
		Password: string(hashed),
	}
	if err := database.DB.Create(&resident).Error; err != nil {
		return response.CodeDBError, err
	}

	return response.CodeSuccess, nil
}

func ResidentLogin(req *dto.LoginRequest) (string, int, error) {
	var resident model.Resident
	if err := database.DB.Where("phone = ?", req.Phone).First(&resident).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", response.CodeUserNotFound, err
		}
		return "", response.CodeDBError, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(resident.Password), []byte(req.Password)); err != nil {
		return "", response.CodePasswordWrong, err
	}

	token, err := jwt.GenerateResidentToken(resident.ID)
	if err != nil {
		return "", response.CodeServerError, err
	}

	return token, response.CodeSuccess, nil
}

func AdminLogin(req *dto.AdminLoginRequest) (string, int, error) {
	var admin model.Admin
	if err := database.DB.Where("username = ?", req.Username).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", response.CodeUserNotFound, err
		}
		return "", response.CodeDBError, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		return "", response.CodePasswordWrong, err
	}

	token, err := jwt.GenerateAdminToken(admin.ID)
	if err != nil {
		return "", response.CodeServerError, err
	}

	return token, response.CodeSuccess, nil
}

func EnsureDefaultAdmin() error {
	var count int64
	database.DB.Model(&model.Admin{}).Count(&count)
	if count > 0 {
		return nil
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	admin := model.Admin{
		Username: "admin",
		Password: string(hashed),
	}
	return database.DB.Create(&admin).Error
}

func GetWeekSlots(weekStart string) (*dto.SlotListResponse, int, error) {
	startDate, err := time.Parse("2006-01-02", weekStart)
	if err != nil {
		return nil, response.CodeParamError, err
	}

	for startDate.Weekday() != time.Monday {
		startDate = startDate.AddDate(0, 0, -1)
	}
	endDate := startDate.AddDate(0, 0, 6)

	var slots []model.TimeSlot
	err = database.DB.Where("slot_date BETWEEN ? AND ?",
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02")).
		Order("slot_date ASC, start_time ASC").
		Find(&slots).Error
	if err != nil {
		return nil, response.CodeDBError, err
	}

	if len(slots) == 0 {
		slots = generateWeekSlots(startDate, endDate)
	}

	result := make([]*dto.SlotInfo, 0, len(slots))
	now := time.Now()
	for _, s := range slots {
		available := s.Capacity - s.BookedCount
		slotDateTime, _ := time.ParseInLocation("2006-01-02 15:04:05",
			s.SlotDate+" "+s.EndTime+":00", time.Local)
		isFull := available == 0 || slotDateTime.Before(now)

		result = append(result, &dto.SlotInfo{
			ID:          s.ID,
			SlotDate:    s.SlotDate,
			StartTime:   s.StartTime,
			EndTime:     s.EndTime,
			Capacity:    s.Capacity,
			BookedCount: s.BookedCount,
			Available:   available,
			IsFull:      isFull,
		})
	}

	return &dto.SlotListResponse{
		WeekStart: startDate.Format("2006-01-02"),
		WeekEnd:   endDate.Format("2006-01-02"),
		Slots:     result,
	}, response.CodeSuccess, nil
}

func generateWeekSlots(startDate, endDate time.Time) []model.TimeSlot {
	slots := make([]model.TimeSlot, 0)
	timeRanges := [][2]string{
		{"09:00", "11:00"},
		{"14:00", "16:00"},
		{"16:00", "18:00"},
	}

	tx := database.DB.Begin()
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		for _, tr := range timeRanges {
			slot := model.TimeSlot{
				SlotDate:  dateStr,
				StartTime: tr[0],
				EndTime:   tr[1],
				Capacity:  10,
			}
			tx.Where(model.TimeSlot{
				SlotDate:  dateStr,
				StartTime: tr[0],
				EndTime:   tr[1],
			}).FirstOrCreate(&slot)
			slots = append(slots, slot)
		}
	}
	tx.Commit()
	return slots
}

func generateOrderNo() string {
	now := time.Now().Format("20060102150405")
	n, _ := rand.Int(rand.Reader, big.NewInt(10000))
	return fmt.Sprintf("AP%s%04d", now, n.Int64())
}

func GetResidentAppointments(residentID uint64) (*dto.MyAppointmentListResponse, int, error) {
	var appointments []model.Appointment
	err := database.DB.Preload("ApplianceType").Preload("Slot").
		Where("resident_id = ?", residentID).
		Order("created_at DESC").
		Find(&appointments).Error
	if err != nil {
		return nil, response.CodeDBError, err
	}

	list := make([]*dto.AppointmentItem, 0, len(appointments))
	for _, a := range appointments {
		list = append(list, convertAppointment(&a))
	}

	return &dto.MyAppointmentListResponse{
		Total: int64(len(list)),
		List:  list,
	}, response.CodeSuccess, nil
}

func GetAdminAppointments(req *dto.AdminAppointmentListRequest) (*dto.AdminAppointmentListResponse, int, error) {
	query := database.DB.Model(&model.Appointment{})

	if req.StartDate != "" {
		startTime := req.StartDate + " 00:00:00"
		query = query.Where("created_at >= ?", startTime)
	}
	if req.EndDate != "" {
		endTime := req.EndDate + " 23:59:59"
		query = query.Where("created_at <= ?", endTime)
	}
	if req.Status != nil && *req.Status > 0 {
		query = query.Where("status = ?", *req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, response.CodeDBError, err
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}
	offset := (req.Page - 1) * req.PageSize

	var appointments []model.Appointment
	err := query.Preload("ApplianceType").Preload("Slot").
		Order("created_at DESC").
		Offset(offset).Limit(req.PageSize).
		Find(&appointments).Error
	if err != nil {
		return nil, response.CodeDBError, err
	}

	list := make([]*dto.AppointmentItem, 0, len(appointments))
	for _, a := range appointments {
		list = append(list, convertAppointment(&a))
	}

	return &dto.AdminAppointmentListResponse{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		List:     list,
	}, response.CodeSuccess, nil
}

func UpdateAppointmentStatus(id uint64, req *dto.UpdateStatusRequest) (int, error) {
	var appointment model.Appointment
	if err := database.DB.Where("id = ?", id).First(&appointment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.CodeAppointmentNotFound, err
		}
		return response.CodeDBError, err
	}

	if appointment.Status == model.AppointmentStatusDone {
		return response.CodeAppointmentDone, errors.New("appointment done")
	}
	if appointment.Status == model.AppointmentStatusCanceled {
		return response.CodeAppointmentCanceled, errors.New("appointment canceled")
	}

	if appointment.Status == model.AppointmentStatusPending &&
		req.Status == model.AppointmentStatusCanceled {
		tx := database.DB.Begin()
		if err := tx.Model(&appointment).Updates(map[string]interface{}{
			"status": req.Status,
			"remark": req.Remark,
		}).Error; err != nil {
			tx.Rollback()
			return response.CodeDBError, err
		}
		if err := tx.Model(&model.TimeSlot{}).
			Where("id = ?", appointment.SlotID).
			UpdateColumn("booked_count", gorm.Expr("booked_count - 1")).Error; err != nil {
			tx.Rollback()
			return response.CodeDBError, err
		}
		tx.Commit()
		return response.CodeSuccess, nil
	}

	updates := map[string]interface{}{
		"status": req.Status,
	}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}

	if err := database.DB.Model(&appointment).Updates(updates).Error; err != nil {
		return response.CodeDBError, err
	}

	return response.CodeSuccess, nil
}

func CancelAppointment(residentID uint64, id uint64, req *dto.CancelAppointmentRequest) (int, error) {
	var appointment model.Appointment
	if err := database.DB.Where("id = ? AND resident_id = ?", id, residentID).First(&appointment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.CodeAppointmentNotFound, err
		}
		return response.CodeDBError, err
	}

	if appointment.Status != model.AppointmentStatusPending {
		return response.CodeAppointmentDone, errors.New("cannot cancel")
	}

	tx := database.DB.Begin()
	if err := tx.Model(&appointment).Updates(map[string]interface{}{
		"status": model.AppointmentStatusCanceled,
		"remark": req.Remark,
	}).Error; err != nil {
		tx.Rollback()
		return response.CodeDBError, err
	}
	if err := tx.Model(&model.TimeSlot{}).
		Where("id = ?", appointment.SlotID).
		UpdateColumn("booked_count", gorm.Expr("booked_count - 1")).Error; err != nil {
		tx.Rollback()
		return response.CodeDBError, err
	}
	tx.Commit()

	return response.CodeSuccess, nil
}

func GetMonthlyStatistics() (*dto.StatisticsResponse, int, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	type StatRow struct {
		ApplianceTypeID uint8
		ApplianceType   string
		Count           int64
		TotalWeight     float64
	}

	var rows []StatRow
	err := database.DB.Table("appointments a").
		Select("a.appliance_type_id, at.name as appliance_type, COUNT(*) as count, SUM(a.appliance_weight) as total_weight").
		Joins("LEFT JOIN appliance_types at ON a.appliance_type_id = at.id").
		Where("a.status = ? AND a.created_at BETWEEN ? AND ?",
			model.AppointmentStatusDone, startOfMonth, endOfMonth).
		Group("a.appliance_type_id, at.name").
		Scan(&rows).Error
	if err != nil {
		return nil, response.CodeDBError, err
	}

	list := make([]*dto.ApplianceStatistic, 0, len(rows))
	var totalCount int64
	var totalWeight float64

	for _, r := range rows {
		list = append(list, &dto.ApplianceStatistic{
			ApplianceTypeID: r.ApplianceTypeID,
			ApplianceType:   r.ApplianceType,
			Count:           r.Count,
			TotalWeight:     r.TotalWeight,
		})
		totalCount += r.Count
		totalWeight += r.TotalWeight
	}

	return &dto.StatisticsResponse{
		Month:         startOfMonth.Format("2006-01"),
		TotalWeight:   totalWeight,
		TotalCount:    totalCount,
		ApplianceList: list,
	}, response.CodeSuccess, nil
}

func GetApplianceTypes() ([]model.ApplianceType, int, error) {
	var types []model.ApplianceType
	err := database.DB.Order("id ASC").Find(&types).Error
	if err != nil {
		return nil, response.CodeDBError, err
	}
	return types, response.CodeSuccess, nil
}

func GetAppointmentDetail(id uint64) (*dto.AppointmentItem, int, error) {
	var appointment model.Appointment
	err := database.DB.Preload("ApplianceType").Preload("Slot").
		Where("id = ?", id).First(&appointment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.CodeAppointmentNotFound, err
		}
		return nil, response.CodeDBError, err
	}
	return convertAppointment(&appointment), response.CodeSuccess, nil
}

func convertAppointment(a *model.Appointment) *dto.AppointmentItem {
	statusText := map[uint8]string{
		1: "待上门",
		2: "已完成",
		3: "已取消",
	}[a.Status]

	item := &dto.AppointmentItem{
		ID:              a.ID,
		OrderNo:         a.OrderNo,
		SlotID:          a.SlotID,
		Phone:           a.Phone,
		Address:         a.Address,
		ApplianceTypeID: a.ApplianceTypeID,
		ApplianceType:   a.ApplianceType.Name,
		ApplianceWeight: a.ApplianceWeight,
		Status:          a.Status,
		StatusText:      statusText,
		Images:          []string(a.Images),
		Remark:          a.Remark,
		CreatedAt:       a.CreatedAt.Format("2006-01-02 15:04:05"),
		SlotDate:        a.Slot.SlotDate,
		StartTime:       a.Slot.StartTime,
		EndTime:         a.Slot.EndTime,
	}
	return item
}
