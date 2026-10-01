package tasks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// DemoSeedUC creates deterministic records used by the local full-stack demo.
type DemoSeedUC struct {
	DB     *gorm.DB
	Logger *zerolog.Logger
}

const DemoAttemptID = "00000000-0000-0000-0000-000000000000"

// DemoDefaultRoutingID is the routing id of the built-in catalog entry, so a
// demo run without a catalog file never collides with imported laboratories.
const DemoDefaultRoutingID = 10000

// DemoCatalog is the laboratory list accepted by `apiserver --demo <file.json>`.
// Every entry is matched against lti_routings by its id and is created or
// overwritten in place; entries absent from the file are left untouched.
type DemoCatalog struct {
	Labs []DemoCatalogLab `json:"labs"`
}

// DemoCatalogLab describes one routing. ID is the routing id and the match key,
// so it must be unique inside the file. LabsPath carries the whole Git link and
// may pin a branch or commit through a `#ref` fragment; the manifests are read
// from the repository root.
type DemoCatalogLab struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	LabsPath    string `json:"labs_path"`
	TestPath    string `json:"test_path"`
}

// DefaultDemoCatalog is the catalog `apiserver --demo` deploys when no file is
// passed. It travels through the same seed path as an imported catalog.
func DefaultDemoCatalog() DemoCatalog {
	return DemoCatalog{Labs: []DemoCatalogLab{{
		ID:          DemoDefaultRoutingID,
		Name:        "Simple Task Demo",
		Description: "Пример задания по автоматизации SSH и мониторингу SNMP",
		LabsPath:    "https://github.com/cms-lab-core/cms-labs-simple-task.git#main",
		TestPath:    "sdn_lab_5",
	}}}
}

func (u *DemoSeedUC) Execute(catalogPath string) error {
	catalog, builtin := DefaultDemoCatalog(), true
	if strings.TrimSpace(catalogPath) != "" {
		loaded, err := loadDemoCatalog(catalogPath)
		if err != nil {
			return err
		}
		catalog, builtin = loaded, false
	}
	if err := validateDemoCatalog(&catalog); err != nil {
		return err
	}
	if err := u.seedCatalog(catalog); err != nil {
		return err
	}
	if !builtin {
		return nil
	}
	return u.seedDemoAttempt()
}

// seedDemoAttempt attaches the stable demo attempt to the built-in routing.
func (u *DemoSeedUC) seedDemoAttempt() error {
	var user models.User
	if err := u.DB.Where("email = ?", "admin@admin.com").First(&user).Error; err != nil {
		return err
	}
	var server models.Server
	if err := u.DB.Where("client_id = ?", "demo-kubernetes").First(&server).Error; err != nil {
		return err
	}
	var route models.LTIRouting
	if err := u.DB.First(&route, DemoDefaultRoutingID).Error; err != nil {
		return err
	}
	var attempt models.LTIAttempt
	err := u.DB.Where("attempt_id = ?", DemoAttemptID).First(&attempt).Error
	if err == gorm.ErrRecordNotFound {
		attempt = models.LTIAttempt{LTIAttemptBase: models.LTIAttemptBase{AttemptID: DemoAttemptID, Status: models.AttemptStatusPending, UserID: user.ID, LTIRoutingID: route.ID}, LTIAttemptSecret: models.LTIAttemptSecret{}}
		attempt.ServerID = &server.ID
		if err := u.DB.Create(&attempt).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	u.Logger.Info().Str("attempt_id", DemoAttemptID).Msg("demo records ready")
	return nil
}

func loadDemoCatalog(filename string) (DemoCatalog, error) {
	file, err := os.Open(filename)
	if err != nil {
		return DemoCatalog{}, fmt.Errorf("open demo catalog: %w", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	catalog := DemoCatalog{}
	if err := decoder.Decode(&catalog); err != nil {
		return DemoCatalog{}, fmt.Errorf("decode demo catalog: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return DemoCatalog{}, fmt.Errorf("decode demo catalog: multiple JSON documents are not allowed")
		}
		return DemoCatalog{}, fmt.Errorf("decode demo catalog: %w", err)
	}
	if err := validateDemoCatalog(&catalog); err != nil {
		return DemoCatalog{}, err
	}
	return catalog, nil
}

// validateDemoCatalog trims and checks every entry, including the built-in one,
// so both seeding paths share the same rules.
func validateDemoCatalog(catalog *DemoCatalog) error {
	if len(catalog.Labs) == 0 {
		return fmt.Errorf("demo catalog contains no labs")
	}
	seen := make(map[uint]struct{}, len(catalog.Labs))
	for index := range catalog.Labs {
		if err := validateDemoCatalogLab(&catalog.Labs[index]); err != nil {
			return fmt.Errorf("lab %d: %w", index+1, err)
		}
		if _, exists := seen[catalog.Labs[index].ID]; exists {
			return fmt.Errorf("duplicate lab id %d", catalog.Labs[index].ID)
		}
		seen[catalog.Labs[index].ID] = struct{}{}
	}
	return nil
}

func validateDemoCatalogLab(lab *DemoCatalogLab) error {
	lab.Name = strings.TrimSpace(lab.Name)
	lab.Description = strings.TrimSpace(lab.Description)
	lab.LabsPath = strings.TrimSpace(lab.LabsPath)
	lab.TestPath = strings.TrimSpace(lab.TestPath)
	if lab.ID == 0 {
		return fmt.Errorf("id is required and must be a positive routing id")
	}
	if len(lab.Description) > 255 || len(lab.TestPath) > 255 {
		return fmt.Errorf("one or more catalog fields exceed database limits")
	}
	if lab.Name == "" || len(lab.Name) > 255 {
		return fmt.Errorf("name is required and must not exceed 255 characters")
	}
	if lab.LabsPath == "" {
		return fmt.Errorf("labs_path is required and must hold the Git link")
	}
	if len(lab.LabsPath) > 255 {
		return fmt.Errorf("labs_path exceeds 255 characters")
	}
	if strings.Contains(lab.LabsPath, "\\") {
		return fmt.Errorf("labs_path must not contain backslashes")
	}
	parsed, err := url.Parse(strings.SplitN(lab.LabsPath, "#", 2)[0])
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("labs_path must start with a full HTTP or HTTPS Git link")
	}
	if parsed.User != nil || parsed.RawQuery != "" {
		return fmt.Errorf("labs_path must not contain credentials or a query")
	}
	return nil
}

func (u *DemoSeedUC) seedCatalog(catalog DemoCatalog) error {
	return u.DB.Transaction(func(tx *gorm.DB) error {
		server := models.Server{}
		if err := tx.Where("client_id = ?", "demo-kubernetes").First(&server).Error; err == gorm.ErrRecordNotFound {
			server = models.Server{ServerBase: models.ServerBase{Name: "Demo Kubernetes", Url: "http://clabgate:8080", Type: models.ServerTypeKubernetes, IsActive: true}, ServerSecret: models.ServerSecret{ClientID: "demo-kubernetes", Token: "demo-secret"}}
			if err := tx.Create(&server).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		for _, lab := range catalog.Labs {
			route := models.LTIRouting{}
			result := tx.First(&route, lab.ID)
			if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
				return result.Error
			}
			route.ID = lab.ID
			route.LTIRoutingBase.Name = lab.Name
			route.LTIDescription = lab.Description
			route.LabsPath = lab.LabsPath
			route.TestPath = lab.TestPath
			route.Collaboration = 1
			route.LabsType = "default"
			route.ServerID = server.ID
			route.IsDefault = lab.ID == DemoDefaultRoutingID
			if result.Error == gorm.ErrRecordNotFound {
				if err := tx.Create(&route).Error; err != nil {
					return err
				}
			} else if err := tx.Save(&route).Error; err != nil {
				return err
			}
		}
		u.Logger.Info().Int("labs", len(catalog.Labs)).Msg("demo catalog imported")
		return nil
	})
}
