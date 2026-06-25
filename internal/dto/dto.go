package dto

type RegisterRequest struct {
	Phone    string `json:"phone" binding:"required,len=11"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type SlotListRequest struct {
	WeekStart string `form:"week_start" binding:"required"`
}

type SlotInfo struct {
	ID          uint64 `json:"id"`
	SlotDate    string `json:"slot_date"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Capacity    uint   `json:"capacity"`
	BookedCount uint   `json:"booked_count"`
	Available   uint   `json:"available"`
	IsFull      bool   `json:"is_full"`
}

type SlotListResponse struct {
	WeekStart string      `json:"week_start"`
	WeekEnd   string      `json:"week_end"`
	Slots     []*SlotInfo `json:"slots"`
}

type CreateAppointmentRequest struct {
	SlotID          uint64   `json:"slot_id" binding:"required,min=1"`
	Phone           string   `json:"phone" binding:"required,len=11"`
	Address         string   `json:"address" binding:"required,min=5,max=255"`
	ApplianceTypeID uint8    `json:"appliance_type_id" binding:"required,min=1"`
	ApplianceWeight float64  `json:"appliance_weight" binding:"required,gt=0"`
	Remark          string   `json:"remark"`
	Images          []string `json:"images,omitempty"`
}

type CreateAppointmentForm struct {
	SlotID          uint64  `form:"slot_id" binding:"required,min=1"`
	Phone           string  `form:"phone" binding:"required,len=11"`
	Address         string  `form:"address" binding:"required,min=5,max=255"`
	ApplianceTypeID uint8   `form:"appliance_type_id" binding:"required,min=1"`
	ApplianceWeight float64 `form:"appliance_weight" binding:"required,gt=0"`
	Remark          string  `form:"remark"`
}

type CreateAppointmentResponse struct {
	OrderNo string `json:"order_no"`
	ID      uint64 `json:"id"`
}

type MyAppointmentListResponse struct {
	Total int64              `json:"total"`
	List  []*AppointmentItem `json:"list"`
}

type AppointmentItem struct {
	ID              uint64   `json:"id"`
	OrderNo         string   `json:"order_no"`
	SlotID          uint64   `json:"slot_id"`
	SlotDate        string   `json:"slot_date"`
	StartTime       string   `json:"start_time"`
	EndTime         string   `json:"end_time"`
	Phone           string   `json:"phone"`
	Address         string   `json:"address"`
	ApplianceTypeID uint8    `json:"appliance_type_id"`
	ApplianceType   string   `json:"appliance_type"`
	ApplianceWeight float64  `json:"appliance_weight"`
	Status          uint8    `json:"status"`
	StatusText      string   `json:"status_text"`
	Images          []string `json:"images,omitempty"`
	Remark          string   `json:"remark,omitempty"`
	CreatedAt       string   `json:"created_at"`
}

type AdminAppointmentListRequest struct {
	Page      int    `form:"page,default=1"`
	PageSize  int    `form:"page_size,default=20"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Status    *uint8 `form:"status"`
}

type AdminAppointmentListResponse struct {
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	List     []*AppointmentItem `json:"list"`
}

type UpdateStatusRequest struct {
	Status uint8  `json:"status" binding:"required,oneof=1 2 3"`
	Remark string `json:"remark"`
}

type StatisticsResponse struct {
	Month         string                `json:"month"`
	TotalWeight   float64               `json:"total_weight"`
	TotalCount    int64                 `json:"total_count"`
	ApplianceList []*ApplianceStatistic `json:"appliance_list"`
}

type ApplianceStatistic struct {
	ApplianceTypeID uint8   `json:"appliance_type_id"`
	ApplianceType   string  `json:"appliance_type"`
	Count           int64   `json:"count"`
	TotalWeight     float64 `json:"total_weight"`
}

type CancelAppointmentRequest struct {
	Remark string `json:"remark"`
}
