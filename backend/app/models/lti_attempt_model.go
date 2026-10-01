package models

import (
	"time"

	"gorm.io/datatypes"
)

const (
	// AttemptStatusPending Попытка создана в базе, но еще не назначена на конкретный сервер или сервер еще не подтвердил прием.
	AttemptStatusPending = "pending"
	// AttemptStatusActive Сервер подтвердил получение задачи и начал работу. Попытка считается «живой».
	AttemptStatusActive = "active"
	// AttemptStatusTerminating Пользователь или система отправили запрос на удаление/завершение, но сервер еще не отчитался об очистке ресурсов.
	AttemptStatusTerminating = "terminating"
	// AttemptStatusCompleted Сервер подтвердил успешное удаление/завершение всех процессов, связанных с этой попыткой. Конечная точка.
	AttemptStatusCompleted = "completed"
)

type LTIAttemptResult struct {
	// Числовые оценки
	MaxScore     float64 `json:"max_score"`     // Максимально возможный балл (например, 10 или 100)
	CurrentScore float64 `json:"current_score"` // Текущий полученный балл
	// Текстовый результат или фидбек
	// Называем универсально Comment или Feedback, так как там может быть и текст, и URL
	ResultDisplay string           `json:"result_display"`
	CheckID       string           `json:"check_id,omitempty"`
	Report        string           `json:"report,omitempty"`
	Logs          string           `json:"logs,omitempty"`
	Tasks         []LTICheckerTask `json:"tasks,omitempty"`
}

// LTICheckerLog is a structured, student-visible message produced by one
// laboratory check. Raw process output is stored separately in Logs.
type LTICheckerLog struct {
	Node      string `json:"node,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Message   string `json:"message"`
}

// LTICheckerTask describes one independently reported laboratory check.
type LTICheckerTask struct {
	Title       string          `json:"title"`
	Description string          `json:"description,omitempty"`
	Logs        []LTICheckerLog `json:"logs,omitempty"`
	Complete    bool            `json:"complete"`
}

// LTIAttemptBase struct to describe LTIAttempt object.
type LTIAttemptBase struct {
	AttemptID      string                                `gorm:"type:varchar(255)" json:"attempt_id" validate:"required"`
	Status         string                                `gorm:"type:varchar(255)" json:"status" validate:"required"`
	Result         *datatypes.JSONType[LTIAttemptResult] `gorm:"type:json" json:"result" swaggertype:"object"`
	UserID         uint                                  `gorm:"type:int" json:"user_id"`
	ServerID       *uint                                 `gorm:"type:int" json:"server_id"`
	LTIRoutingID   uint                                  `gorm:"type:int" json:"lti_routing_id"`
	SynchronizedAt *time.Time                            `gorm:"type:datetime(3)" json:"synchronized_at"`
}

func (a *LTIAttemptBase) SetStatus(status string) {
	// Порядок статусов (чем больше число, тем "старше" статус)
	statusOrder := map[string]int{
		AttemptStatusPending:     1,
		AttemptStatusActive:      2,
		AttemptStatusTerminating: 3,
		AttemptStatusCompleted:   4,
	}

	currentOrder, currentExists := statusOrder[a.Status]
	newOrder, newExists := statusOrder[status]

	// Если какой-то статус неизвестен — не обновляем
	if !currentExists || !newExists {
		return
	}

	// Разрешаем переход только если новый статус "больше" текущего
	if newOrder > currentOrder {
		a.Status = status
	}
}

type LTIAttemptSecret struct {
	RoomID *uint `gorm:"type:int" json:"room_id"`
}

type LTIAttemptListItem struct {
	Base
	LTIAttemptBase
	UserEmail      string `json:"user_email"`
	UserName       string `json:"user_name"`
	ServerName     string `json:"server_name"`
	LTIRoutingName string `json:"lti_routing_name"`
	Repository     string `json:"repository"`
	TestPath       string `json:"test_path"`
}

// TableName переопределяет название таблицы для LTIAttemptListItem на `lti_attempts`
func (LTIAttemptListItem) TableName() string {
	return "lti_attempts"
}

type LTIAttempt struct {
	Base
	LTIAttemptBase
	LTIAttemptSecret
}
