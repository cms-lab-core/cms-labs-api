package queries

import (
	"errors"

	"github.com/cms-lab-core/cms-labs-api/pnetlabaddon/app/models"
	"gorm.io/gorm"
)

type LabSessionQuery struct {
	DB *gorm.DB
}

var (
	LabSessionNotFoundError = errors.New("Lab Session not found")
)

func (q *LabSessionQuery) tableName(object interface{}) string {
	stmt := &gorm.Statement{DB: q.DB}
	_ = stmt.Parse(object)
	return stmt.Schema.Table
}

func (q *LabSessionQuery) GetByAttemptId(attemptId string) (models.LabSession, error) {
	entityDB := models.LabSession{}
	q.DB.Where("lab_session_lid = ?", attemptId).Find(&entityDB)
	if entityDB.LabSessionID != 0 {
		return entityDB, nil
	}
	return entityDB, LabSessionNotFoundError
}

// LabSessionRunningLabs модель для активных сессий лабораторных
type LabSessionRunningLabs struct {
	LabSessionID   int    `gorm:"column:lab_session_id" json:"lab_session_id"`
	LabSessionLID  string `gorm:"column:lab_session_lid" json:"lab_session_lid"`
	LabSessionPath string `gorm:"column:lab_session_path" json:"lab_session_path"`
	Name           string `gorm:"column:name" json:"name"`
	Email          string `gorm:"column:email" json:"email"`
}

func (q *LabSessionQuery) GetRunningLabs() ([]LabSessionRunningLabs, error) {
	var entities []LabSessionRunningLabs
	query := q.DB.Table(
		q.tableName(&models.LabSession{})+" AS lab_sessions",
	).Select(
		"lab_sessions.lab_session_id, lab_sessions.lab_session_path, users.name, lab_sessions.lab_session_lid, users.email",
	).Joins(
		"left join "+q.tableName(&models.User{})+" users on users.pod = lab_sessions.lab_session_pod",
	).Where(
		"lab_sessions.lab_session_lid != ?",
		"",
	)
	query = query.Offset(0).Order("lab_sessions.lab_session_id desc")
	if err := query.Scan(&entities).Error; err != nil {
		return entities, err
	}
	return entities, nil
}
