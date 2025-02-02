package tasks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-co-op/gocron"
	"slashbase.com/backend/internal/config"
	"slashbase.com/backend/internal/dao"
	"slashbase.com/backend/internal/models"
)

func InitCron() {
	scheduler := gocron.NewScheduler(time.UTC)
	clearOldLogs(scheduler)
	telemetryPings(scheduler)
	scheduler.StartAsync()
}

func clearOldLogs(s *gocron.Scheduler) {
	s.Every(1).Day().Do(func() {
		setting, err := dao.Setting.GetSingleSetting(models.SETTING_NAME_LOGS_EXPIRE)
		if err != nil {
			return
		}
		days := setting.Int()
		err = dao.DBQueryLog.ClearOldLogs(days)
		if err != nil {
			return
		}
	})
}

func telemetryPings(s *gocron.Scheduler) {
	if !config.IsLive() {
		return
	}
	s.Every(1).Day().Do(func() {
		setting, err := dao.Setting.GetSingleSetting(models.SETTING_NAME_TELEMETRY_ENABLED)
		if err != nil {
			return
		}
		if !setting.Bool() {
			return
		}
		setting, err = dao.Setting.GetSingleSetting(models.SETTING_NAME_APP_ID)
		if err != nil {
			return
		}
		values := map[string]interface{}{
			"api_key": "phc_XSWvMvnTUEH9pLJDVmYfaKaKH8QZtK5fJO8NIiFoNwv",
			"event":   "Telemetry Ping",
			"properties": map[string]string{
				"distinct_id": setting.UUID().String(),
				"version":     config.VERSION,
			},
		}
		json_data, _ := json.Marshal(values)
		http.Post("https://app.posthog.com/capture/", "application/json",
			bytes.NewBuffer(json_data))
	})
}
