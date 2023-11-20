package logs

import (
	"net/http"

	"github.com/bb-consent/api/internal/actionlog"
	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
)

// ConfigPurgeOrgLogs
func ConfigPurgeOrgLogs(w http.ResponseWriter, r *http.Request) {
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	// Repository
	actionLogRepo := actionlog.ActionLogRepository{}
	actionLogRepo.Init(organisationId)

	// Count logs
	count, err := actionLogRepo.CountLogs()
	if err != nil {
		m := "Failed to count logs"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	// Purge all logs except last 100 logs
	if count > 100 {
		log, err := actionLogRepo.GetLogOfIndexHundread()
		if err != nil {
			m := "Failed to fetch logs"
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}

		err = actionLogRepo.DeleteLogsLessThanTimestamp(log.Timestamp)
		if err != nil {
			m := "Failed to purge logs"
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}
	}

	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

}
