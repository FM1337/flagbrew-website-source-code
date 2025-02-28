package helpers

import (
	"time"

	"github.com/FM1337/flagbrew-website-source-code/pkg/models"
)

func GenerateFailedUploadLog(uploaderIP, uploadSource, discordUser, failedReason string, isPatron bool, patronCode, patronDiscord string) *models.GPSSFailedUploadLog {
	return &models.GPSSFailedUploadLog{
		Date:            time.Now(),
		UploaderIP:      uploaderIP,
		UploadSource:    uploadSource,
		UploaderDiscord: discordUser,
		FailedReason:    failedReason,
		Patron:          isPatron,
		PatronCode:      patronCode,
		PatronDiscord:   patronDiscord,
		LogType:         "gpss_failed_upload",
	}
}

func GenerateUploadLog(uploaderIP, uploadSource, discordUser string, isDeleted bool,
	pokemon models.Pokemon, isApproved bool, approvedBy string, isPatron, uploadedInBundle bool, downloadCode, bundleCode string, patronCode, patronDiscord string) *models.GPSSUploadLog {
	return &models.GPSSUploadLog{
		Date:            time.Now(),
		UploaderIP:      uploaderIP,
		UploadSource:    uploadSource,
		UploaderDiscord: discordUser,
		Deleted:         isDeleted,
		PokemonData:     pokemon,
		Approved:        isApproved,
		ApprovedBy:      approvedBy,
		Patron:          isPatron,
		PatronCode:      patronCode,
		PatronDiscord:   patronDiscord,
		BundleUpload:    uploadedInBundle,
		DownloadCode:    downloadCode,
		BundleCode:      bundleCode,
		LogType:         "gpss_upload",
		DBVersion:       2,
	}
}

func GenerateDeletionLog(deleter, reason, entityType, downloadCode string) *models.GPSSDeletionLog {
	return &models.GPSSDeletionLog{
		Date:           time.Now(),
		DeletedBy:      deleter,
		DeletionReason: reason,
		EntityType:     entityType,
		DownloadCode:   downloadCode,
		LogType:        "gpss_deletion",
	}
}

func GenerateUnbanLog(unbanner string, Ban *models.Ban) *models.UnbanLog {
	return &models.UnbanLog{
		Date:       time.Now(),
		Ban:        *Ban,
		UnbannedBy: unbanner,
		LogType:    "unban",
	}
}

func GenerateSettingChangeLog(setting, modifiedBy string, original, new interface{}) *models.SettingChangeLog {
	return &models.SettingChangeLog{
		Date:          time.Now(),
		Setting:       setting,
		OriginalValue: original,
		NewValue:      new,
		ModifiedBy:    modifiedBy,
		LogType:       "setting_change",
	}
}

func GenerateGPSSCleanLog(deleted, reset, failed int64) *models.GPSSCleanLog {
	return &models.GPSSCleanLog{
		Date:    time.Now(),
		Deleted: deleted,
		Reset:   reset,
		Failed:  failed,
		LogType: "gpss_clean",
	}
}

func GenerateBan(ip, reason, banner string) *models.Ban {
	return &models.Ban{
		Date:      time.Now(),
		IP:        ip,
		BanReason: reason,
		BannedBy:  banner,
	}
}

func GeneratePatreonBuildDeleteLog(hash, filename string, expiry time.Time) *models.PatreonBuildDeleteLog {
	return &models.PatreonBuildDeleteLog{
		Date:               time.Now(),
		CommitHash:         hash,
		Filename:           filename,
		OriginalExpiryDate: expiry,
		LogType:            "build_delete",
	}
}

func GenerateUnrestrictLog(user string, original *models.RestrictedUploader) *models.UnrestrictedLog {
	return &models.UnrestrictedLog{
		Date:             time.Now(),
		UnrestrictedBy:   user,
		OriginalRestrict: *original,
		LogType:          "unrestrict",
	}
}

func GenerationBundleUpsertLog(uploaderIP, uploadSource, discordUser string, isPatreon bool,
	pokemons []models.Pokemon, downloadCode, patronCode, patronDiscord string, downloadCodes []string, approved bool) *models.GPSSBundleUploadLog {

	return &models.GPSSBundleUploadLog{
		Date:            time.Now(),
		UploaderIP:      uploaderIP,
		UploadSource:    uploadSource,
		UploaderDiscord: discordUser,
		Deleted:         false,
		Pokemons:        pokemons,
		Patron:          isPatreon,
		PatronCode:      patronCode,
		PatronDiscord:   patronDiscord,
		DownloadCodes:   downloadCodes,
		DownloadCode:    downloadCode,
		LogType:         "gpss_bundle_upload",
		Approved:        approved,
		DBVersion:       2,
	}
}

func GenerateWordDeleteLog(user, original string) *models.WordDeleteLog {
	return &models.WordDeleteLog{
		Date:         time.Now(),
		DeletedBy:    user,
		OriginalWord: original,
		LogType:      "word_delete",
	}
}
