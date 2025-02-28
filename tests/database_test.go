package tests

type database struct {
	URI      string
	MaxConns int
}

var dbInfo database

// Init sets up the test
func init() {
	dbInfo.URI = "localhost/"
	dbInfo.MaxConns = 128
}

// Tests to see if we can successfully connect to our mongodb database,
// func TestMongoCanConnectSuccessfully(t *testing.T) {
// 	mongoSrv := mongo.NewSrv()
// 	if err := mongoSrv.Setup(dbInfo.URI, dbInfo.MaxConns, nil); err != nil {
// 		t.Errorf("Expected no error, got error of %s instead", err.Error())
// 	}
// }
