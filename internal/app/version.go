package app

import "time"

var buildVersion = "dev"

func init() {
	if buildVersion == "" || buildVersion == "dev" {
		buildVersion = "dev " + time.Now().UTC().Format(time.RFC3339)
	}
}
