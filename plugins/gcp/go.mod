module github.com/scttfrdmn/snoozebot/plugins/gcp

go 1.18

require (
	cloud.google.com/go/compute v1.22.0
	github.com/hashicorp/go-hclog v0.16.2
	github.com/hashicorp/go-plugin v1.6.3
	github.com/scttfrdmn/snoozebot v0.1.0
	google.golang.org/api v0.147.0
)

replace github.com/scttfrdmn/snoozebot => ../../