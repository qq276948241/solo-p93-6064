package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Resident struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Phone     string    `gorm:"size:20;unique;not null" json:"phone"`
	Password  string    `gorm:"size:255;not null" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Resident) TableName() string {
	return "residents"
}

type Admin struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string    `gorm:"size:50;unique;not null" json:"username"`
	Password  string    `gorm:"size:255;not null" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Admin) TableName() string {
	return "admins"
}

type ApplianceType struct {
	ID        uint8     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:50;unique;not null" json:"name"`
	CreatedAt time.Time `json:"-"`
}

func (ApplianceType) TableName() string {
	return "appliance_types"
}

type TimeSlot struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	SlotDate    string    `gorm:"type:date;not null" json:"slot_date"`
	StartTime   string    `gorm:"type:time;not null" json:"start_time"`
	EndTime     string    `gorm:"type:time;not null" json:"end_time"`
	Capacity    uint      `gorm:"not null;default:10" json:"capacity"`
	BookedCount uint      `gorm:"not null;default:0" json:"booked_count"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}

func (TimeSlot) TableName() string {
	return "time_slots"
}

const (
	AppointmentStatusPending  = 1
	AppointmentStatusDone     = 2
	AppointmentStatusCanceled = 3
)

type Appointment struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderNo         string    `gorm:"size:32;unique;not null" json:"order_no"`
	ResidentID      uint64    `gorm:"not null" json:"resident_id"`
	SlotID          uint64    `gorm:"not null" json:"slot_id"`
	Phone           string    `gorm:"size:20;not null" json:"phone"`
	Address         string    `gorm:"size:255;not null" json:"address"`
	ApplianceTypeID uint8     `gorm:"not null" json:"appliance_type_id"`
	ApplianceWeight float64   `gorm:"type:decimal(8,2);not null" json:"appliance_weight"`
	Status          uint8     `gorm:"not null;default:1" json:"status"`
	Images          StringArr `gorm:"type:text" json:"images,omitempty"`
	Remark          string    `gorm:"size:255;default:null" json:"remark,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	ApplianceType ApplianceType `gorm:"foreignKey:ApplianceTypeID" json:"appliance_type,omitempty"`
	Slot          TimeSlot      `gorm:"foreignKey:SlotID" json:"slot,omitempty"`
}

func (Appointment) TableName() string {
	return "appointments"
}

type StringArr []string

func (a StringArr) Value() (driver.Value, error) {
	if len(a) == 0 {
		return nil, nil
	}
	b, err := json.Marshal(a)
	return string(b), err
}

func (a *StringArr) Scan(input interface{}) error {
	if input == nil {
		*a = nil
		return nil
	}
	var bytes []byte
	switch v := input.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("invalid type for StringArr")
	}
	return json.Unmarshal(bytes, a)
}
