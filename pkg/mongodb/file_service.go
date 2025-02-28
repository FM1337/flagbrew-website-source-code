package mongo

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/FM1337/flagbrew-website-source-code/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ghSrv satisfies the models.GithubService interface.
type fileSrv struct {
	srv *mongoSrv
}

type file struct {
	Filename string `bson:"filename"`
}

func (s *mongoSrv) NewFileService() *fileSrv {
	return &fileSrv{srv: s}
}

func (s *fileSrv) UploadMysteryGift(ctx context.Context, filename string, data []byte) (size int, err error) {
	fileID, err := s.srv.mysteryGift.getFileID(ctx, bson.M{"filename": filename}, options.GridFSFind().SetLimit(1))
	if err == nil {
		// Drop the existing file
		err = s.srv.mysteryGift.Delete(fileID)
		if err != nil {
			// return if error
			return 0, err
		}
	}
	// Okay now let's upload!
	return s.srv.mysteryGift.upload(filename, data) // options.GridFSUpload().SetMetadata(bson.M{"yeet": "yote"}) for metadata
}

func (s *fileSrv) DownloadMysteryGift(ctx context.Context, filename string) (size int64, file bytes.Buffer, err error) {
	return s.srv.mysteryGift.download(filename)
}

func (s *fileSrv) UploadPatreonBuild(ctx context.Context, filename string, data []byte, hash, app, extension string) (size int, err error) {

	// Set the expiry date field to 1 week after upload
	return s.srv.patreon.upload(filename, data, options.GridFSUpload().SetMetadata(bson.M{"commit_hash": hash, "app": app, "extension": extension, "expire_date": time.Now().Add(time.Hour * 168)}))
}

func (s *fileSrv) PatreonBuildExists(ctx context.Context, app, hash, extension string) bool {
	_, err := s.srv.patreon.getFileID(ctx, bson.M{"metadata.app": app, "metadata.commit_hash": hash, "metadata.extension": extension}, options.GridFSFind().SetLimit(1))
	if err != nil {
		return false
	}

	return true
}

func (s *fileSrv) DownloadPatreonBuild(ctx context.Context, app, hash, extension string) (size int64, file bytes.Buffer, err error) {
	// get the FileID
	fileID, err := s.srv.patreon.getFileID(ctx, bson.M{"metadata.app": app, "metadata.commit_hash": hash, "metadata.extension": extension}, options.GridFSFind().SetLimit(1))

	if err != nil {
		if err != gridfs.ErrFileNotFound {
			return size, file, err
		}
	}

	// fileID shouldn't be nil if there wasn't any error above, but just in case
	if fileID == nil {
		return size, file, fmt.Errorf("FileID is unexpectedly nil, have Allen look into this!\nQuery info: %s, %s, %s", app, hash, extension)
	}

	// Now get the file
	return s.srv.patreon.downloadByID(fileID)
}

func (s *fileSrv) GetLatestAppHash(ctx context.Context, app string) (hash interface{}, err error) {
	return s.srv.patreon.getMetaField(ctx, bson.M{"metadata.app": app}, "commit_hash", options.GridFSFind().SetSort(bson.M{"uploadDate": -1}).SetLimit(1))
}

func (s *fileSrv) CleanOldBuilds(ctx context.Context, app, hash string) (deleted []*models.File, err error) {
	cursor, err := s.srv.patreon.Find(bson.M{"metadata.app": app, "metadata.commit_hash": bson.M{"$ne": hash}, "metadata.expire_date": bson.M{"$lte": time.Now()}})
	if err != nil {
		return nil, err
	}
	files := []*models.File{}
	// Loop through the cursor
	err = cursor.All(context.Background(), &files)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		// Delete each file
		err = s.srv.patreon.Delete(f.ID)
		if err != nil {
			return nil, err
		}
		deleted = append(deleted, f)
	}

	return deleted, err
}
