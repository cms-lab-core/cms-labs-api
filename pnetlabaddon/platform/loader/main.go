package main

import (
	"fmt"
	"io"
	"os"

	"github.com/cms-lab-core/cms-labs-api/pnetlabaddon/app/models"

	"ariga.io/atlas-provider-gorm/gormschema"
)

func main() {
	stmts, err := gormschema.New("mysql").Load(
		models.Wireshark{},
		models.User{},
		models.UserRole{},
		models.UserPermission{},
		models.ProcessDevice{},
		models.Process{},
		models.NodeSession{},
		models.LabSession{},
		models.IfSession{},
		models.HTML5{},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load gorm schema: %v\n", err)
		os.Exit(1)
	}
	_, _ = io.WriteString(os.Stdout, stmts)
}
