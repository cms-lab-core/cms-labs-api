package models

// LTIRoutingBase struct to describe LTIRouting object.
type LTIRoutingBase struct {
	Name string `gorm:"type:varchar(255);column:name" json:"name"`
}

type LTIRoutingSecret struct {
	// LTI Params
	LTITitle       string `gorm:"type:varchar(255);column:lti_title" json:"lti_title"`
	LTIDescription string `gorm:"type:varchar(255);column:lti_description" json:"lti_description"`
	LTIParamsTask  string `gorm:"type:varchar(255);column:lti_params_task" json:"lti_params_task"`
	// Параметры
	Collaboration int `gorm:"type:int;column:collaboration" json:"collaboration"`
	// The type of PNETLabsType, cms_client.PNETLabsTypeDefault
	// enum: default,curl,sso
	LabsType string `gorm:"type:varchar(255);column:labs_type" json:"labs_type"`
	// LabsPath holds the Git link of the laboratory with an optional `#ref`
	// fragment, for example https://git.example.org/course/lab.git#main.
	// The manifests are read from the repository root.
	LabsPath  string `gorm:"type:varchar(255);column:labs_path" json:"labs_path"`
	TestPath  string `gorm:"type:varchar(255);column:test_path" json:"test_path"`
	ServerID  uint   `gorm:"type:int;column:server_id" json:"server_id"`
	IsDefault bool   `gorm:"type:bool;column:is_default" json:"is_default"`
	// Автоматические
	LTITaskID   string `gorm:"type:varchar(255);column:lti_task_id" json:"lti_task_id"`
	LTICourseID string `gorm:"type:varchar(255);column:lti_course_id" json:"lti_course_id"`
	LTISubID    string `gorm:"type:varchar(255);column:lti_sub_id" json:"lti_sub_id"`
}

type LTIRoutingListItem struct {
	Base
	LTIRoutingBase
}

// TableName переопределяет название таблицы для LTIRoutingListItem на `lti_routings`
func (LTIRoutingListItem) TableName() string {
	return "lti_routings"
}

type LTIRouting struct {
	Base
	LTIRoutingBase
	LTIRoutingSecret
}
