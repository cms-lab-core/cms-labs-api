package usecases

import (
	"errors"
	"strings"

	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/cms-lab-core/cms-labs-api/backend/pkg/configs"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
)

// ErrLabCatalogDisabled is returned when the catalog is used while
// LAB_CATALOG_ENABLED is not true.
var ErrLabCatalogDisabled = errors.New("lab catalog is disabled")

type LabCatalogUC struct {
	LTIRoutingQueries *queries.LTIRoutingQueries
	LTIAttemptQueries *queries.LTIAttemptQueries
	AttemptCreateUC   *LTIAttemptCreateUC
	user              *cms_client.SSOTokenPublicData
}

type LabCatalogListInputDTO struct {
	Search string `json:"search"`
}

type LabCatalogListRequest struct {
	JSONRPC string                 `json:"jsonrpc" default:"2.0" validate:"required"`
	Method  string                 `json:"method" default:"lab_catalog.list" validate:"required"`
	Params  LabCatalogListInputDTO `json:"params,omitempty"`
	ID      string                 `json:"id,omitempty" default:"1" validate:"required"`
}

type LabCatalogAttemptDTO struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type LabCatalogItemDTO struct {
	ID            uint                  `json:"id"`
	Name          string                `json:"name"`
	Description   string                `json:"description,omitempty"`
	Collaboration int                   `json:"collaboration"`
	Repository    string                `json:"repository"`
	Attempt       *LabCatalogAttemptDTO `json:"attempt,omitempty"`
}

type LabCatalogListOutputDTO struct {
	Labs []LabCatalogItemDTO `json:"labs"`
}

type LabCatalogStartInputDTO struct {
	LabID uint `json:"lab_id" validate:"required"`
}

type LabCatalogStartRequest struct {
	JSONRPC string                  `json:"jsonrpc" default:"2.0" validate:"required"`
	Method  string                  `json:"method" default:"lab_catalog.start" validate:"required"`
	Params  LabCatalogStartInputDTO `json:"params,omitempty"`
	ID      string                  `json:"id,omitempty" default:"1" validate:"required"`
}

type LabCatalogStartOutputDTO struct {
	AttemptID string `json:"attempt_id"`
	Status    string `json:"status"`
	NextURL   string `json:"next_url"`
}

func (u *LabCatalogUC) SetContext(user *cms_client.SSOTokenPublicData) *LabCatalogUC {
	u.user = user
	u.AttemptCreateUC.SetContext(user)
	return u
}

func (u *LabCatalogUC) List(dto LabCatalogListInputDTO) (LabCatalogListOutputDTO, error) {
	if u.user == nil {
		return LabCatalogListOutputDTO{}, errors.New("not logged in")
	}
	if !configs.AppConfig.LabCatalog {
		return LabCatalogListOutputDTO{Labs: []LabCatalogItemDTO{}}, nil
	}
	routes, err := u.LTIRoutingQueries.ListLabs(strings.TrimSpace(dto.Search))
	if err != nil {
		return LabCatalogListOutputDTO{}, err
	}
	routeIDs := make([]uint, 0, len(routes))
	for _, route := range routes {
		routeIDs = append(routeIDs, route.ID)
	}
	attempts, err := u.LTIAttemptQueries.ListActiveByUserID(u.user.UserID(), routeIDs)
	if err != nil {
		return LabCatalogListOutputDTO{}, err
	}
	attemptsByRoute := make(map[uint]LabCatalogAttemptDTO, len(attempts))
	for _, attempt := range attempts {
		if _, exists := attemptsByRoute[attempt.LTIRoutingID]; exists {
			continue
		}
		attemptsByRoute[attempt.LTIRoutingID] = LabCatalogAttemptDTO{
			ID: attempt.AttemptID, Status: attempt.Status,
		}
	}
	items := make([]LabCatalogItemDTO, 0, len(routes))
	for _, route := range routes {
		item := LabCatalogItemDTO{
			ID: route.ID, Name: route.Name, Description: route.LTIDescription,
			Collaboration: route.Collaboration, Repository: route.LabsPath,
		}
		if attempt, exists := attemptsByRoute[route.ID]; exists {
			item.Attempt = &attempt
		}
		items = append(items, item)
	}
	return LabCatalogListOutputDTO{Labs: items}, nil
}

func (u *LabCatalogUC) Start(dto LabCatalogStartInputDTO) (LabCatalogStartOutputDTO, error) {
	if u.user == nil {
		return LabCatalogStartOutputDTO{}, errors.New("not logged in")
	}
	if !configs.AppConfig.LabCatalog {
		return LabCatalogStartOutputDTO{}, ErrLabCatalogDisabled
	}
	attempt, response, err := u.AttemptCreateUC.ExecuteCatalog(dto.LabID)
	if err != nil {
		return LabCatalogStartOutputDTO{}, err
	}
	return LabCatalogStartOutputDTO{AttemptID: attempt.AttemptID, Status: attempt.Status, NextURL: response.NextUrl}, nil
}
