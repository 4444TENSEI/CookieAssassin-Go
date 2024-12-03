package model

import (
	"encoding/json"

	"gorm.io/gorm"
)

// Cookie基础数据表
type CkInfo struct {
	ID         uint   `json:"ID" comment:"自增序号"`
	Url        string `gorm:"type:varchar(255);default:null;" json:"url" comment:"网址"`
	UrlTitle   string `gorm:"type:varchar(255);default:null;" json:"url_title" comment:"网址标题"`
	Ip         string `gorm:"type:varchar(255);default:null;" json:"ip" comment:"IP地址"`
	ChaoxingId string `gorm:"type:varchar(255);default:null;index;" json:"chaoxing_name" comment:"超星用户名"`
	Type       string `gorm:"type:varchar(255);default:null;index;" json:"type" comment:"分类"`
	Tag        string `gorm:"type:varchar(255);default:null;" json:"tag" comment:"标签"`
	gorm.Model
}

// Cookie独立数据表
type CkData struct {
	ID           uint            `json:"ID" comment:"对应ck_info表的id"`
	Data         json.RawMessage `gorm:"type:json;default:null;" json:"data" comment:"数据"`
	Type         string          `gorm:"type:varchar(255);default:null;index;" json:"type" comment:"分类"`
	UpdateStatus string          `gorm:"type:varchar(255);default:null;index;" json:"update_status" comment:"续期状态"`
	gorm.Model
}
