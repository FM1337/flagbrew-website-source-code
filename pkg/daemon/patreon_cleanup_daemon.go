package daemon

import (
	"context"
	"time"

	"github.com/FM1337/flagbrew-website-source-code/pkg/helpers"
	"github.com/FM1337/flagbrew-website-source-code/pkg/models"
)

type PatreonCleanupDaemon struct {
	fileSrv models.FileService
	logSrv  models.LogService
	runtime time.Duration
	enabled bool
}

func NewPatreonCleanupDaemon(fileSrv *models.FileService, logSrv *models.LogService) *PatreonCleanupDaemon {
	return &PatreonCleanupDaemon{fileSrv: *fileSrv, logSrv: *logSrv, runtime: time.Hour*24 + (time.Minute * 3), enabled: false}
}

func (d *PatreonCleanupDaemon) Start() {
	d.enabled = true
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if !d.enabled {
				break
			}
			mutex.Lock()
			d.cleanPatron()
			mutex.Unlock()
			time.Sleep(d.runtime)
		}
	}()
}

func (d *PatreonCleanupDaemon) cleanPatron() {
	var log interface{}
	// define available apps
	apps := []string{"PKSM", "Checkpoint"}
	// Loop through each app
	for _, app := range apps {
		// Get the latest hash
		hash, err := d.fileSrv.GetLatestAppHash(context.Background(), app)
		if err != nil {
			continue
		}

		deleted, err := d.fileSrv.CleanOldBuilds(context.Background(), app, hash.(string))
		if err != nil {
			helpers.LogToSentry(err)
			continue
		}
		// log each deleted build
		for _, build := range deleted {
			log = helpers.GeneratePatreonBuildDeleteLog(build.Metadata.CommitHash, build.Filename, build.Metadata.ExpireDate)
			err = d.logSrv.UpsertLog(context.Background(), &log)
			if err != nil {
				helpers.LogToSentry(err)
			}
		}
	}
}

func (d *PatreonCleanupDaemon) Stop() {
	d.enabled = false
}
