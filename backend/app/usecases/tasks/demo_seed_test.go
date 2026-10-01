package tasks

import (
	"os"
	"path/filepath"
	"testing"
)

func writeDemoCatalog(t *testing.T, content string) string {
	t.Helper()
	filename := filepath.Join(t.TempDir(), "labs.json")
	if err := os.WriteFile(filename, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return filename
}

func TestDefaultDemoCatalogPassesValidation(t *testing.T) {
	catalog := DefaultDemoCatalog()
	if err := validateDemoCatalog(&catalog); err != nil {
		t.Fatal(err)
	}
	lab := catalog.Labs[0]
	if lab.ID != DemoDefaultRoutingID || lab.Name == "" || lab.LabsPath == "" {
		t.Fatalf("unexpected built-in lab: %#v", lab)
	}
}

func TestLoadDemoCatalogKeepsRefInsideGitLink(t *testing.T) {
	catalog, err := loadDemoCatalog(writeDemoCatalog(t,
		`{"labs":[{"id":7,"name":" Lab ","description":" About ","labs_path":" https://git.example.test/lab.git#release ","test_path":" sdn_lab_5 "}]}`))
	if err != nil {
		t.Fatal(err)
	}
	lab := catalog.Labs[0]
	if lab.ID != 7 || lab.Name != "Lab" || lab.Description != "About" ||
		lab.LabsPath != "https://git.example.test/lab.git#release" || lab.TestPath != "sdn_lab_5" {
		t.Fatalf("unexpected catalog lab: %#v", lab)
	}
}

func TestLoadDemoCatalogRejectsInvalidEntries(t *testing.T) {
	for name, content := range map[string]string{
		"version":     `{"version":1,"labs":[{"id":1,"name":"Lab","labs_path":"https://git.example.test/lab.git"}]}`,
		"extra field": `{"labs":[{"id":1,"name":"Lab","labs_path":"https://git.example.test/lab.git","labs_type":"default"}]}`,
		"no id":       `{"labs":[{"name":"Lab","labs_path":"https://git.example.test/lab.git"}]}`,
		"zero id":     `{"labs":[{"id":0,"name":"Lab","labs_path":"https://git.example.test/lab.git"}]}`,
		"no name":     `{"labs":[{"id":1,"labs_path":"https://git.example.test/lab.git"}]}`,
		"no link":     `{"labs":[{"id":1,"name":"Lab","labs_path":"task"}]}`,
		"credentials": `{"labs":[{"id":1,"name":"Lab","labs_path":"https://user:pass@git.example.test/lab.git"}]}`,
		"query":       `{"labs":[{"id":1,"name":"Lab","labs_path":"https://git.example.test/lab.git?token=1"}]}`,
		"backslashes": `{"labs":[{"id":1,"name":"Lab","labs_path":"https://git.example.test\\lab.git"}]}`,
		"empty":       `{"labs":[]}`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := loadDemoCatalog(writeDemoCatalog(t, content)); err == nil {
				t.Fatal("invalid catalog was accepted")
			}
		})
	}
}

func TestLoadDemoCatalogRejectsDuplicateIDs(t *testing.T) {
	_, err := loadDemoCatalog(writeDemoCatalog(t,
		`{"labs":[{"id":1,"name":"A","labs_path":"https://git.example.test/a.git"},{"id":1,"name":"B","labs_path":"https://git.example.test/b.git"}]}`))
	if err == nil {
		t.Fatal("duplicate lab ids were accepted")
	}
}
