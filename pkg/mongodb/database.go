package mongo

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type mongoSrv struct {
	db     *mongo.Database
	client *mongo.Client
	log    *log.Logger

	patreon     grid
	mysteryGift grid

	repos      col
	members    col
	users      col
	gpss       col
	bundles    col
	logs       col
	bans       col
	settings   col
	patrons    col
	restricted col
	words      col
}

type col struct {
	*mongo.Collection
}

type grid struct {
	*gridfs.Bucket
}

type gridFileID struct {
	ID primitive.ObjectID `bson:"_id"`
}

type mField struct {
	Value interface{} `bson:"Value"`
}

// // getField gets the requested field and returns it.
// func (c *col) getField(ctx context.Context, filter interface{}, field string) (result *mongo.SingleResult, err error) {
// 	result = c.FindOne(ctx, filter, options.FindOne().SetProjection(bson.M{field: 1}))
// 	return result, result.Err()
// }

// // f (short for find) returns a cursor containing the results of the filter/query provided (along with an error if one occurs)
// func (c *col) f(filter interface{}, opts ...*options.FindOptions) (cursor *mongo.Cursor, err error) {
// 	return c.Find(context.TODO(), filter, opts...)
// }

// // in (short for insert) inserts a new document into the database and returns the result (along with an error if one occurs)
// func (c *col) in(document interface{}, opts ...*options.InsertOneOptions) (result *mongo.InsertOneResult, err error) {
// 	return c.InsertOne(context.TODO(), document, opts...)
// }

// // in (short for insertMany) inserts multiple new document into the database and returns the results (along with an error if one occurs)
// func (c *col) inm(documents []interface{}, opts ...*options.InsertManyOptions) (result *mongo.InsertManyResult, err error) {
// 	return c.InsertMany(context.TODO(), documents, opts...)
// }

// // f (short for findOne) returns a single result based on the query/filter provided (along with an error if one occurs)
// func (c *col) fOne(filter interface{}, opts ...*options.FindOneOptions) (result *mongo.SingleResult) {
// 	return c.FindOne(context.TODO(), filter, opts...)
// }

// // up (short for upsert) inserts/updates a collection in(to) the database and returns the result (along with an error if one occurs)
// func (c *col) up(filter, update interface{}, opts ...*options.UpdateOptions) (result *mongo.UpdateResult, err error) {
// 	return c.UpdateOne(context.TODO(), filter, update, opts...)
// }

// // rm (short for remove) removes documents from the database using the provided filter/query and returns the delete result (along with an error if one occurs)
// func (c *col) rm(filter interface{}, opts ...*options.DeleteOptions) (result *mongo.DeleteResult, err error) {
// 	return c.DeleteMany(context.TODO(), filter, opts...)
// }

// // rm (short for remove one) removes a single document from the database using the provided filter/query and returns the delete result (along with an error if one occurs)
// func (c *col) rm1(filter interface{}, opts ...*options.DeleteOptions) (result *mongo.DeleteResult, err error) {
// 	return c.DeleteOne(context.TODO(), filter, opts...)
// }

// // co (short for count) counts the amount of documents that would be returned based on the filter/query and returns an int64 (along with an error if one occurs)
// func (c *col) co(filter interface{}, opts ...*options.CountOptions) (count int64, err error) {
// 	return c.CountDocuments(context.TODO(), filter, opts...)
// }

// // f1u (short for find one and update) finds a document, updates and then returns that document.
// func (c *col) f1u(filter, update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult {
// 	return c.FindOneAndUpdate(context.TODO(), filter, update, opts...)
// }

// // um (short for update many) updates many documents with the passed data
// func (c *col) um(filter, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
// 	return c.UpdateMany(context.TODO(), filter, update, opts...)
// }

// upload uploads a gridfs file and returns the size of file and an error if one was encountered
func (g *grid) upload(filename string, data []byte, opts ...*options.UploadOptions) (size int, err error) {
	// open the upload stream
	upload, err := g.OpenUploadStream(filename, opts...)
	if err != nil {
		return 0, err
	}

	defer upload.Close()

	size, err = upload.Write(data)
	if err != nil {
		return 0, err
	}

	return size, nil
}

// download downloads a gridfs file by filename and returns the file size, the file itself an an error if one was encountered
func (g *grid) download(filename string, opts ...*options.NameOptions) (size int64, file bytes.Buffer, err error) {
	file = bytes.Buffer{}

	download, err := g.DownloadToStreamByName(filename, &file, opts...)
	if err != nil {
		return 0, file, err
	}

	return download, file, err
}

// downloadByID downloads a gridfs file by file ID and returns the file size, the file itself an an error if one was encountered
func (g *grid) downloadByID(fileID interface{}) (size int64, file bytes.Buffer, err error) {
	file = bytes.Buffer{}

	download, err := g.DownloadToStream(fileID, &file)
	if err != nil {
		return 0, file, err
	}

	return download, file, err
}

// getFileID searches gridfs for the file based on parameters and returns its ID
func (g *grid) getFileID(ctx context.Context, filter interface{}, opts ...*options.GridFSFindOptions) (fileID interface{}, err error) {
	cursor, err := g.Find(filter, opts...)
	if err != nil {
		return nil, err
	}
	if cursor.RemainingBatchLength() == 1 {
		if !cursor.TryNext(ctx) {
			return nil, gridfs.ErrFileNotFound
		}
	}

	fileID, err = cursor.Current.LookupErr("_id")
	return fileID, err
}

// getMetaField searches gridfs for the requested metadata field based on search parameters and returns the field value or an error if one is hit
func (g *grid) getMetaField(ctx context.Context, filter interface{}, field string, opts ...*options.GridFSFindOptions) (metaField interface{}, err error) {
	cursor, err := g.Find(filter, opts...)
	if err != nil {
		return nil, err
	}

	if cursor.RemainingBatchLength() >= 1 {
		if !cursor.TryNext(ctx) {
			return nil, gridfs.ErrFileNotFound
		}
	} else {
		return nil, gridfs.ErrFileNotFound
	}

	mField, err := cursor.Current.LookupErr("metadata", field)
	if err != nil {
		return nil, err
	}
	if result, ok := mField.StringValueOK(); ok {
		metaField = result
	} else {
		err = fmt.Errorf("Could not convert requested metafield %s to string for returning", field)
	}

	return metaField, err
}

// NewSrv returns a new mongoSrv
func NewSrv() *mongoSrv {
	return &mongoSrv{}
}

// initColEntry inits each collection in mongoSrv by setting the functions
func (s *mongoSrv) initColEntry(name string) (c col) {
	c.Collection = s.db.Collection(name)
	return c
}

// initGridEntry inits each gridfs bucket in mongoSrv by setting the functions
func (s *mongoSrv) initGridEntry(name string) (g grid, err error) {
	g.Bucket, err = gridfs.NewBucket(s.db, options.GridFSBucket().SetName(name))
	return g, err
}

// Setup does setup
func (s *mongoSrv) Setup(uri, user, pass string, maxConns int, logger *log.Logger) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	opts := &options.ClientOptions{}
	opts.SetAppName("flagbrew")
	opts.SetConnectTimeout(15 * time.Second)
	opts.SetDirect(true)
	opts.SetMaxConnIdleTime(120 * time.Second)
	opts.SetMinPoolSize(uint64(maxConns))
	opts.SetRetryReads(true)
	opts.SetRetryWrites(true)
	opts.ApplyURI(fmt.Sprintf("mongodb://%s", uri))
	// Authentication Stuff
	opts.SetAuth(options.Credential{
		AuthSource: "admin",
	 	Username:   user,
	 	Password:   pass,
	 })

	s.client, err = mongo.Connect(ctx, opts)

	if err != nil {
		return err
	}

	if err = s.client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("unable to ping mongodb primary: %v", err)
	}

	s.log = logger

	s.db = s.client.Database("flagbrew2")
	s.gpss = s.initColEntry("gpss")
	s.bundles = s.initColEntry("bundles")
	s.members = s.initColEntry("members")
	s.repos = s.initColEntry("repos")
	s.users = s.initColEntry("users")
	s.logs = s.initColEntry("logs")
	s.bans = s.initColEntry("bans")
	s.settings = s.initColEntry("settings")
	s.patrons = s.initColEntry("patrons")
	s.restricted = s.initColEntry("restricted")
	s.words = s.initColEntry("words")

	s.patreon, err = s.initGridEntry("patreon")
	if err != nil {
		return fmt.Errorf("unable to create gridfs bucket for patreon builds: %v", err)
	}

	s.mysteryGift, err = s.initGridEntry("mystery_gift")
	if err != nil {
		return fmt.Errorf("unable to create gridfs bucket for mystery gifts: %v", err)
	}

	return nil
}

func (s *mongoSrv) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s.log.Println("received signal to stop; beginning to close connections")
	if err := s.client.Disconnect(ctx); err != nil {
		err = fmt.Errorf("unable to close mongodb connection: %v", err)
		return err
	}
	return nil
}
