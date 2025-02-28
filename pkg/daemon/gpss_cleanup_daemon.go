package daemon

import (
	"context"
	"fmt"
	"time"

	"github.com/FM1337/flagbrew-website-source-code/pkg/helpers"
	"github.com/FM1337/flagbrew-website-source-code/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
)

type GPSSCleanupDaemon struct {
	settings   map[string]*models.Setting
	runtime    time.Duration // Originally wanted this to be a month, moving it to 1 week as the last_reset field now acts as a safety net to only reset downloads if the last_reset date was 1 month or more ago and the same for deletions
	enabled    bool
	logSrv     models.LogService
	settingSrv models.SettingService
	gpssSrv    models.GPSSService
}

func NewGPSSCleanupDaemon(settings map[string]*models.Setting, gpssSrv *models.GPSSService, logSrv *models.LogService, settingSrv *models.SettingService) *GPSSCleanupDaemon {
	return &GPSSCleanupDaemon{settings: settings, runtime: 10080 * time.Minute, enabled: false, gpssSrv: *gpssSrv, logSrv: *logSrv, settingSrv: *settingSrv}
}

func (d *GPSSCleanupDaemon) Start() {
	d.enabled = true
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if !d.enabled {
				break
			}
			mutex.Lock()
			if d.settings["gpss_clean_enabled"].Value.(bool) {
				d.cleanGPSS()
			}
			mutex.Unlock()
			time.Sleep(d.runtime)
		}
	}()
}

func (d *GPSSCleanupDaemon) cleanGPSS() {
	var log interface{}
	// First disable downloads
	setting := d.settings["gpss_download_enabled"]
	originalValue := setting.Value.(bool)
	if originalValue {
		setting.Value = false
		setting.ModifiedDate = time.Now()
		d.settings[setting.MapKey] = setting
		// Upsert this in the database and log
		d.settingSrv.UpdateSetting(context.Background(), setting.MapKey, false)
		log = helpers.GenerateSettingChangeLog(setting.MapKey, "System", true, false)
		d.logSrv.UpsertLog(context.Background(), &log)

	}
	// Now clean GPSS
	// We want to first remove any Pokemon with Current Download counts of 0 AND where last_reset is a month or greater
	pkmn, _, _, err := d.gpssSrv.ListPokemons(context.Background(), bson.M{"current_downloads": 0, "last_reset": bson.M{"$lt": time.Now().Add(-43800 * time.Minute)}, "deleted": false}, 1, -1, bson.M{}, false)
	deleted := 0
	failed := 0
	if err != nil {
		helpers.LogToSentry(err)
		// Disable cleaning to be safe
		cleanSetting := *d.settings["gpss_clean_enabled"]
		cleanSetting.Value = false
		cleanSetting.ModifiedDate = time.Now()
		d.settings[cleanSetting.MapKey] = &cleanSetting
		// Upsert this in the database and log
		d.settingSrv.UpdateSetting(context.Background(), cleanSetting.MapKey, false)
		log = helpers.GenerateSettingChangeLog(cleanSetting.MapKey, "System", true, false)
		err = d.logSrv.UpsertLog(context.Background(), &log)
		if err != nil {
			helpers.LogToSentry(err)
		}
	} else {
		for _, pk := range pkmn {
			// Delete the pokemon
			err = d.gpssSrv.RemovePokemon(context.Background(), pk.DownloadCode, false, true)
			if err != nil {
				helpers.LogToSentry(err)
				continue
			}
			// Log the deletion
			log = helpers.GenerateDeletionLog("System", fmt.Sprintf("Cleaned by GPSS Cleaning Daemon %d lifetime downloads at the time of deletion", pk.LifetimeDownloads), "Pokemon", pk.DownloadCode)
			err = d.logSrv.UpsertLog(context.Background(), &log)
			if err != nil {
				helpers.LogToSentry(err)
				failed++
				continue
			}
			deleted++
			d.logSrv.UpdateLog(context.Background(), bson.M{"log_type": "gpss_upload", "download_code": pk.DownloadCode}, bson.M{"$set": bson.M{"deleted": true}})
		}

		// Now call the reset downloads method
		amount, err := d.gpssSrv.ResetOldPokemonDownloads(context.Background())
		if err != nil {
			helpers.LogToSentry(err)
		} else {
			log = helpers.GenerateGPSSCleanLog(int64(deleted), amount, int64(failed))
			err = d.logSrv.UpsertLog(context.Background(), &log)
			if err != nil {
				helpers.LogToSentry(err)
			}
		}
	}

	// When done re-enable downloads (only if the value was true to begin with)
	if originalValue {
		setting.Value = true
		setting.ModifiedDate = time.Now()
		d.settings[setting.MapKey] = setting
		d.settingSrv.UpdateSetting(context.Background(), setting.MapKey, true)
		log = helpers.GenerateSettingChangeLog(setting.MapKey, "System", false, true)
		d.logSrv.UpsertLog(context.Background(), &log)
	}
}

func (d *GPSSCleanupDaemon) Stop() {
	d.enabled = false
}
