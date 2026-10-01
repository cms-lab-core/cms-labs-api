package routes

import (
	"testing"

	json "github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/maintainer64/cms-labs-api/backend/app/models"
	"github.com/maintainer64/cms-labs-api/backend/pkg/configs"
)

func TestLabCatalogListAndIdempotentStart(t *testing.T) {
	f := NewTestHTTP()
	defer f.Close()
	configs.AppConfig.LabCatalog = true
	defer func() { configs.AppConfig.LabCatalog = false }()

	server := models.Server{
		ServerBase:   models.ServerBase{Name: "Catalog test", Url: "http://clabgate:8080", Type: models.ServerTypeKubernetes, IsActive: true},
		ServerSecret: models.ServerSecret{ClientID: uuid.NewString(), Token: uuid.NewString()},
	}
	if err := f.DB.Create(&server).Error; err != nil {
		t.Fatal(err)
	}
	defer f.DB.Delete(&server)

	route := models.LTIRouting{
		LTIRoutingBase: models.LTIRoutingBase{Name: "Catalog route test " + uuid.NewString()},
		LTIRoutingSecret: models.LTIRoutingSecret{
			Collaboration: 1, LabsType: "default",
			LabsPath: "https://git.example.test/course/lab.git#main", ServerID: server.ID,
		},
	}
	if err := f.DB.Create(&route).Error; err != nil {
		t.Fatal(err)
	}
	defer f.DB.Delete(&route)

	authorization := f.AuthorizationUser(0, nil)
	status, body := f.Rpc(&TestRpcRequest{Method: "lab_catalog.list", Params: map[string]any{}, Authorization: authorization})
	if status != 200 {
		t.Fatalf("list status=%d body=%s", status, body)
	}
	var listResponse struct {
		Result struct {
			Labs []struct {
				ID         uint   `json:"id"`
				Name       string `json:"name"`
				Repository string `json:"repository"`
			} `json:"labs"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(body), &listResponse); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, lab := range listResponse.Result.Labs {
		if lab.ID != route.ID {
			continue
		}
		found = true
		if lab.Name != route.Name || lab.Repository != route.LabsPath {
			t.Fatalf("unexpected catalog item: %#v", lab)
		}
	}
	if !found {
		t.Fatalf("catalog route %q is missing: %s", route.Name, body)
	}

	start := func() string {
		status, body = f.Rpc(&TestRpcRequest{Method: "lab_catalog.start", Params: map[string]any{"lab_id": route.ID}, Authorization: authorization})
		if status != 200 {
			t.Fatalf("start status=%d body=%s", status, body)
		}
		var response struct {
			Result struct {
				AttemptID string `json:"attempt_id"`
				NextURL   string `json:"next_url"`
			} `json:"result"`
		}
		if err := json.Unmarshal([]byte(body), &response); err != nil {
			t.Fatal(err)
		}
		if response.Result.AttemptID == "" || response.Result.NextURL != "/session/"+response.Result.AttemptID {
			t.Fatalf("unexpected start response: %s", body)
		}
		return response.Result.AttemptID
	}
	firstAttemptID := start()
	if secondAttemptID := start(); secondAttemptID != firstAttemptID {
		t.Fatalf("start is not idempotent: %q != %q", firstAttemptID, secondAttemptID)
	}
	defer f.DB.Where("attempt_id = ?", firstAttemptID).Delete(&models.LTIAttempt{})
}

// With LAB_CATALOG_ENABLED off the API serves an empty catalog, so the frontend
// can hide the section, and starting a laboratory is rejected.
func TestLabCatalogDisabledServesNothing(t *testing.T) {
	f := NewTestHTTP()
	defer f.Close()

	configs.AppConfig.LabCatalog = false
	authorization := f.AuthorizationUser(0, nil)
	status, body := f.Rpc(&TestRpcRequest{Method: "lab_catalog.list", Params: map[string]any{}, Authorization: authorization})
	if status != 200 {
		t.Fatalf("list status=%d body=%s", status, body)
	}
	var listResponse struct {
		Result struct {
			Labs []map[string]any `json:"labs"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(body), &listResponse); err != nil {
		t.Fatal(err)
	}
	if len(listResponse.Result.Labs) != 0 {
		t.Fatalf("disabled catalog returned laboratories: %s", body)
	}
	status, body = f.Rpc(&TestRpcRequest{Method: "lab_catalog.start", Params: map[string]any{"lab_id": 1}, Authorization: authorization})
	if status == 200 {
		t.Fatalf("disabled catalog started a laboratory: %s", body)
	}
}
