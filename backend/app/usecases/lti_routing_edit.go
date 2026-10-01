package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
)

type LTIRoutingEditUC struct {
	LTIRoutingQueries *queries.LTIRoutingQueries
}

type LTIRoutingEditInputDTO struct {
	ID             uint   `json:"id"`
	Name           string `json:"name"`
	LTITitle       string `json:"lti_title"`
	LTIDescription string `json:"lti_description"`
	LTITaskID      string `json:"lti_task_id"`
	LTICourseID    string `json:"lti_course_id"`
	LTISubID       string `json:"lti_sub_id"`
	LTIParamsTask  string `json:"lti_params_task"`
	Collaboration  int    `json:"collaboration"`
	LabsType       string `json:"labs_type"`
	LabsPath       string `json:"labs_path"`
	TestPath       string `json:"test_path"`
	ServerID       uint   `json:"server_id"`
	IsDefault      bool   `json:"is_default"`
}

type LTIRoutingEditRequest struct {
	JSONRPC string                 `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string                 `json:"method" default:"lti_routing.upsert" required:"true"`
	Params  LTIRoutingEditInputDTO `json:"params,omitempty"`
	ID      string                 `json:"id,omitempty" default:"1" required:"true"`
}

type LTIRoutingEditOutputDTO struct {
	ID uint `json:"id" required:"true"`
}

type LTIRoutingEditResponse struct {
	JSONRPC string                  `json:"jsonrpc" default:"2.0" required:"true"`
	Result  LTIRoutingEditOutputDTO `json:"result,omitempty"`
	Error   interface{}             `json:"error,omitempty"`
	ID      string                  `json:"id,omitempty" default:"1" required:"true"`
}

func (u *LTIRoutingEditUC) Execute(dto LTIRoutingEditInputDTO) (LTIRoutingEditOutputDTO, error) {
	entity := &models.LTIRouting{}
	entity.ID = dto.ID
	entity.Name = dto.Name
	entity.LTITitle = dto.LTITitle
	entity.LTIDescription = dto.LTIDescription
	entity.LTITaskID = dto.LTITaskID
	entity.LTICourseID = dto.LTICourseID
	entity.LTISubID = dto.LTISubID
	entity.LTIParamsTask = dto.LTIParamsTask
	entity.Collaboration = dto.Collaboration
	entity.LabsType = dto.LabsType
	entity.LabsPath = dto.LabsPath
	entity.TestPath = dto.TestPath
	entity.ServerID = dto.ServerID
	entity.IsDefault = dto.IsDefault
	err := u.LTIRoutingQueries.Upsert(entity)
	return LTIRoutingEditOutputDTO{ID: entity.ID}, err
}
