package utils

import "github.com/hhstu/prometheus-proxy/log"

func init() {
	log.Logger.Infow("gin_template", "Version", Version, "BuildDate", BuildDate, "Commit", Commit)
}

var (
	Version   = "-"
	BuildDate = ""
	Commit    = ""
)
