package mongo

import (
	"fmt"

	"github.com/FM1337/flagbrew-website-source-code/pkg/models"
	"go.mongodb.org/mongo-driver/mongo"
)

// errorWrapper is a useful wrapper for returning customized errors from mongo.
func errorWrapper(err error) error {
	switch err {
	case mongo.ErrNoDocuments:
		return &models.ErrClientError{Err: fmt.Errorf("The requested data could not be found")} // &models.ErrNotFound{Err: err}
	default:
		return err
	}
}
