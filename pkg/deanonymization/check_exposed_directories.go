package deanonymization

import (
	"net/url"
	"strings"

	"github.com/JOwlHouse/onionscan/pkg/config"
	"github.com/JOwlHouse/onionscan/pkg/report"
)

// CheckExposedDirectories makes note of any directories which display a directory listing
// instead of an expected 403 or equivalent error.
func CheckExposedDirectories(osreport *report.OnionScanReport, report *report.AnonymityReport, osc *config.OnionScanConfig) {
	for key, id := range osreport.Crawls {
		crawlRecord, _ := osc.Database.GetCrawlRecord(id)
		if crawlRecord.Page.Status == 200 && strings.Contains(crawlRecord.Page.Title, "Index of") {
			uri, _ := url.Parse(key)
			report.OpenDirectories = append(report.OpenDirectories, uri.Path)
		}
	}
}
