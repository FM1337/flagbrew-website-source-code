package helpers

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/FM1337/flagbrew-website-source-code/pkg/models"
	"github.com/getsentry/sentry-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Download Section
	gpssDownloadCounters map[string]*prometheus.CounterVec
	gpssDownloadGauges   map[string]*prometheus.GaugeVec
	gpssDownloadTime     *prometheus.HistogramVec

	// Upload Section
	gpssUploadCounters map[string]*prometheus.CounterVec
	gpssUploadGauges   map[string]*prometheus.GaugeVec
	gpssPatronUploads  prometheus.Counter
	gpssUploadTime     *prometheus.HistogramVec
)

func registerMetric(metricType, name, help string, labels []string) interface{} {
	switch metricType {
	case "counterVec":
		fallthrough
	case "counter":
		counterOpts := prometheus.CounterOpts{
			Name: name,
			Help: help,
		}
		if metricType == "counter" {
			return promauto.NewCounter(counterOpts)
		}
		return promauto.NewCounterVec(counterOpts, labels)
	case "histogramVec":
		return promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name: name,
			Help: help,
		},
			labels,
		)
	case "gaugeVec":
		return promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: name,
			Help: help,
		},
			labels)
	}
	return nil
}

func getGaugeData(svcGPSS models.GPSSService, mapKey, field string, downloads bool) {
	stats, err := svcGPSS.ListCountForFieldStat(context.Background(), field, downloads)
	if err != nil {
		LogToSentry(err)
		os.Exit(1)
	}

	if downloads {
		for label, count := range stats {
			gpssDownloadGauges[mapKey].WithLabelValues(label).Set(count)
		}
	} else {
		for label, count := range stats {
			gpssUploadGauges[mapKey].WithLabelValues(label).Set(count)
		}
	}
}

func increaseData(key, label string, downloads, hasGauge bool) {
	if downloads {
		gpssDownloadCounters[key].WithLabelValues(label).Inc()
		if hasGauge {
			gpssDownloadGauges[key].WithLabelValues(label).Inc()
		}
	} else {
		if key == "patron" {
			gpssPatronUploads.Inc()
			return
		} else {
			gpssUploadCounters[key].WithLabelValues(label).Inc()
		}
		if hasGauge {
			gpssUploadGauges[key].WithLabelValues(label).Inc()
		}
	}
}

func generateMetrics(metricType string, keys, helpName []string, downloads, current bool) {
	helpDownload := "The %s number of GPSS downloads based on %s"
	helpUpload := "The %s number of GPSS uploads based on %s"
	metricFormat := "gpss_%s_%s"
	if !current {
		metricFormat = "gpss_total_%s_%s"
	}
	for i, key := range keys {
		mode := "total"
		if current {
			mode = "current"
		}

		if downloads {
			switch metricType {
			case "counterVec":
				gpssDownloadCounters[key] = registerMetric(metricType, fmt.Sprintf(metricFormat, key, "downloads"), fmt.Sprintf(helpDownload, mode, helpName[i]), []string{key}).(*prometheus.CounterVec)
			case "gaugeVec":
				gpssDownloadGauges[key] = registerMetric(metricType, fmt.Sprintf(metricFormat, key, "downloads"), fmt.Sprintf(helpDownload, mode, helpName[i]), []string{key}).(*prometheus.GaugeVec)
			}
		} else {
			switch metricType {
			case "counterVec":
				gpssUploadCounters[key] = registerMetric(metricType, fmt.Sprintf(metricFormat, key, "uploads"), fmt.Sprintf(helpUpload, mode, helpName[i]), []string{key}).(*prometheus.CounterVec)
			case "gaugeVec":
				gpssUploadGauges[key] = registerMetric(metricType, fmt.Sprintf(metricFormat, key, "uploads"), fmt.Sprintf(helpUpload, mode, helpName[i]), []string{key}).(*prometheus.GaugeVec)
			}
		}
	}
}

// RegisterMetrics inits/registers each custom metric we want to track
func RegisterMetrics(svcGPSS models.GPSSService) {
	gpssDownloadCounters = make(map[string]*prometheus.CounterVec)
	gpssDownloadGauges = make(map[string]*prometheus.GaugeVec)
	gpssUploadCounters = make(map[string]*prometheus.CounterVec)
	gpssUploadGauges = make(map[string]*prometheus.GaugeVec)
	// Downloads
	// Vector Histogram
	gpssDownloadTime = registerMetric("histogramVec", "gpss_download_time", "The amount of time it takes for each kind of GPSS download to finish", []string{"type"}).(*prometheus.HistogramVec)
	// Vector Counters
	generateMetrics("counterVec", []string{"download_type", "species", "generation", "legality", "shiny", "egg", "gender"}, []string{"type", "species", "generations", "legality", "shininess", "egg", "gender"}, true, true)
	// Vector Gauges
	generateMetrics("gaugeVec", []string{"download_type", "species", "generation", "legality", "shiny", "egg", "gender"}, []string{"type", "species", "generations", "legality", "shininess", "egg", "gender"}, true, false)
	getGaugeData(svcGPSS, "generation", "$generation", true)
	getGaugeData(svcGPSS, "species", "$pokemon.species", true)
	getGaugeData(svcGPSS, "legality", "$pokemon.is_legal", true)
	getGaugeData(svcGPSS, "shiny", "$pokemon.is_shiny", true)
	getGaugeData(svcGPSS, "egg", "$pokemon.is_egg", true)
	getGaugeData(svcGPSS, "gender", "$pokemon.gender", true)

	// Uploads
	// Vector Histogram
	gpssUploadTime = registerMetric("histogramVec", "gpss_upload_time", "The amount of time it takes for each kind of GPSS upload to finish", []string{"type"}).(*prometheus.HistogramVec)
	// Vector Counters
	generateMetrics("counterVec", []string{"upload_type", "species", "generation", "legality", "shiny", "egg", "gender"}, []string{"type", "species", "generations", "legality", "shininess", "egg", "gender"}, false, true)
	// Vector Gauges
	generateMetrics("gaugeVec", []string{"upload_type", "species", "generation", "legality", "shiny", "egg", "gender"}, []string{"type", "species", "generations", "legality", "shininess", "egg", "gender"}, false, false)
	getGaugeData(svcGPSS, "generation", "$generation", false)
	getGaugeData(svcGPSS, "species", "$pokemon.species", false)
	getGaugeData(svcGPSS, "legality", "$pokemon.is_legal", false)
	getGaugeData(svcGPSS, "shiny", "$pokemon.is_shiny", false)
	getGaugeData(svcGPSS, "egg", "$pokemon.is_egg", false)
	getGaugeData(svcGPSS, "gender", "$pokemon.gender", false)
	// Counter
	gpssPatronUploads = registerMetric("counter", "gpss_patron_uploads", "The current number of GPSS uploads from patrons", nil).(prometheus.Counter)
}

// IncreaseDownloads increments the prometheus download stats for each download stat I'm tracking.
func IncreaseDownloads(downloadType string, generations, species, gender []string, legality, shiny, egg []bool) {
	gpssDownloadCounters["download_type"].WithLabelValues(downloadType).Inc()
	// loop through our arrays

	// Unless something funky happens, generations and species should always have the same size
	// but just in case
	if len(legality)+len(species)+len(generations)+len(shiny)+len(egg)+len(gender) != len(generations)*6 {
		success, context := GenerateSentryEventLogContext([]string{"download_type", "generations_count", "species_count", "gender_count", "legality_count",
			"shiny_count", "egg_count"}, []interface{}{downloadType, len(generations), len(species), len(gender), len(legality), len(shiny), len(egg)})
		if success {
			LogToSentryWithContext(sentry.LevelError, "Invalid length of arguments provided to IncreaseDownloads", context)
		}
		return
	}
	for i, generation := range generations {
		increaseData("species", species[i], true, true)
		increaseData("generation", generation, true, true)
		increaseData("legality", strconv.FormatBool(legality[i]), true, true)
		increaseData("shiny", strconv.FormatBool(shiny[i]), true, true)
		increaseData("egg", strconv.FormatBool(egg[i]), true, true)
		increaseData("gender", gender[i], true, true)
	}
}

// IncreaseUploads increments the prometheus download stats for each download stat I'm tracking.
func IncreaseUploads(uploadType string, generations, species, gender []string, legality, shiny, egg []bool, patreon bool) {
	gpssUploadCounters["upload_type"].WithLabelValues(uploadType).Inc()
	// loop through our arrays

	// Unless something funky happens, generations and species should always have the same size
	// but just in case
	if len(legality)+len(species)+len(generations)+len(shiny)+len(egg)+len(gender) != len(generations)*6 {
		success, context := GenerateSentryEventLogContext([]string{"upload_type", "generations_count", "species_count", "gender_count", "legality_count",
			"shiny_count", "egg_count", "patreon"}, []interface{}{uploadType, len(generations), len(species), len(gender), len(legality), len(shiny), len(egg), patreon})
		if success {
			LogToSentryWithContext(sentry.LevelError, "Invalid length of arguments provided to IncreaseUploads", context)
		}
		return
	}
	for i, generation := range generations {
		increaseData("species", species[i], false, true)
		increaseData("generation", generation, false, true)
		increaseData("legality", strconv.FormatBool(legality[i]), false, true)
		increaseData("shiny", strconv.FormatBool(shiny[i]), false, true)
		increaseData("egg", strconv.FormatBool(egg[i]), false, true)
		increaseData("gender", gender[i], false, true)
	}
	if patreon {
		gpssPatronUploads.Inc()
	}
}

// MeasureDownloadTime records the time it's taken to process a download.
func MeasureDownloadTime(downloadType string, duration time.Duration) {
	gpssDownloadTime.WithLabelValues(downloadType).Observe(float64(duration.Milliseconds()))
}

// MeasureUploadTime records the time it's taken to process a download.
func MeasureUploadTime(uploadType string, duration time.Duration) {
	gpssUploadTime.WithLabelValues(uploadType).Observe(float64(duration.Milliseconds()))
}
