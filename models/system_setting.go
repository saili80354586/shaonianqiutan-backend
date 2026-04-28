package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

const AdminSystemSettingKey = "admin.system_settings"

// SystemSetting 平台系统设置，按 key 存储 JSON 配置。
type SystemSetting struct {
	Key       string    `json:"key" gorm:"primaryKey;size:100"`
	Value     string    `json:"value" gorm:"type:text;not null"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (SystemSetting) TableName() string {
	return "system_settings"
}

type AdminSystemSettings struct {
	SiteName            string  `json:"siteName"`
	AllowRegistration   bool    `json:"allowRegistration"`
	RequireApproval     bool    `json:"requireApproval"`
	DefaultAnalystPrice float64 `json:"defaultAnalystPrice"`
	CommissionRate      float64 `json:"commissionRate"`
	MaintenanceMode     bool    `json:"maintenanceMode"`
}

type PublicSystemSettings struct {
	SiteName          string `json:"siteName"`
	AllowRegistration bool   `json:"allowRegistration"`
	MaintenanceMode   bool   `json:"maintenanceMode"`
}

func DefaultAdminSystemSettings() AdminSystemSettings {
	return AdminSystemSettings{
		SiteName:            "少年球探",
		AllowRegistration:   true,
		RequireApproval:     true,
		DefaultAnalystPrice: 299,
		CommissionRate:      20,
		MaintenanceMode:     false,
	}
}

func LoadAdminSystemSettings(db *gorm.DB) AdminSystemSettings {
	settings := DefaultAdminSystemSettings()
	if db == nil {
		return settings
	}

	var row SystemSetting
	if err := db.Where("key = ?", AdminSystemSettingKey).First(&row).Error; err == nil && row.Value != "" {
		_ = json.Unmarshal([]byte(row.Value), &settings)
	}
	return settings
}

func (settings AdminSystemSettings) Public() PublicSystemSettings {
	return PublicSystemSettings{
		SiteName:          settings.SiteName,
		AllowRegistration: settings.AllowRegistration,
		MaintenanceMode:   settings.MaintenanceMode,
	}
}
