package protocol

import (
	"github.com/JOwlHouse/onionscan/pkg/config"
	"github.com/JOwlHouse/onionscan/pkg/report"
)

type Scanner interface {
	ScanProtocol(hiddenService string, onionscanConfig *config.OnionScanConfig, report *report.OnionScanReport)
}
