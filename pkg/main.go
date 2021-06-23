package main

import (
	"os"

	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

func main() {
	// Start listening to requests send from Grafana. This call is blocking so
	// it won't finish until Grafana shutsdown the process or the plugin choose
	// to exit close down by itself
	if err := datasource.Manage("teamviewer-datasource", NewWebMonitoringDatasource, datasource.ManageOpts{}); err != nil {
		log.DefaultLogger.Error(err.Error())
		os.Exit(1)
	}
}
