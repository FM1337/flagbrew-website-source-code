package mongo

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/FM1337/flagbrew-website-source-code/pkg/helpers"
	"github.com/FM1337/flagbrew-website-source-code/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// filterSrv satisfies the models.FilterService interface.
type filterSrv struct {
	srv *mongoSrv
}

func (s *mongoSrv) NewFilterService() *filterSrv {
	return &filterSrv{srv: s}
}

func (s *filterSrv) AddWord(ctx context.Context, word string, strict, caseInsensitive bool, createdBy string) (err error) {
	exists, _, _, err := s.CheckWords(ctx, []string{word}, "full", caseInsensitive)
	if exists {
		return fmt.Errorf("word exists already")
	}
	if err != nil {
		return err
	}

	// Word doesn't exist, let's insert it
	insertDoc := &models.WordFilter{
		String:        word,
		Strict:        strict,
		CaseSensitive: !caseInsensitive,
		AddedBy:       createdBy,
		CreatedDate:   time.Now(),
	}

	_, err = s.srv.words.InsertOne(ctx, insertDoc)
	return err
}

func (s *filterSrv) RemoveWord(ctx context.Context, word string, caseInsensitive bool) (err error) {
	exists, _, _, err := s.CheckWords(ctx, []string{word}, "full", caseInsensitive)
	if !exists {
		return fmt.Errorf("word doesn't exist")
	}
	if err != nil {
		return err
	}

	// Word exists, let's remove it
	_, err = s.srv.words.DeleteOne(ctx, bson.M{"string": word})

	return err
}

func (s *filterSrv) CheckWords(ctx context.Context, words []string, mode string, caseInsensitive bool) (match, strict bool, index int, err error) {
	regexMode := ""
	for i, word := range words {
		if word == "" {
			continue
		}
		if caseInsensitive && mode != "full" {
			word = strings.ToLower(word)
			regexMode = "i"
		}
		query := bson.M{}
		switch mode {
		case "start":
			query = bson.M{"string": bson.M{"$regex": primitive.Regex{Pattern: fmt.Sprintf("^%s.*", regexp.QuoteMeta(word)),
				Options: regexMode}}}
		case "end":
			query = bson.M{"string": bson.M{"$regex": primitive.Regex{Pattern: fmt.Sprintf("%s$", regexp.QuoteMeta(word)),
				Options: regexMode}}}
		case "full":
			query = bson.M{"string": word}
		case "any":
			query = bson.M{"string": bson.M{"$regex": primitive.Regex{Pattern: fmt.Sprintf(".*%s.*", regexp.QuoteMeta(word)),
				Options: regexMode}}}
		}
		result := s.srv.words.FindOne(ctx, query)
		if result.Err() != nil {
			if result.Err() != mongo.ErrNoDocuments {
				helpers.LogToSentry(err)
				return false, false, -1, err
			}
			continue
		}
		matchDocument := &models.WordFilter{}
		match = true
		index = i
		err = result.Decode(&matchDocument)
		if err != nil {
			return match, false, index, err
		}
		strict = matchDocument.Strict
		return match, strict, index, err
	}

	return false, false, -1, nil
}

func (s *filterSrv) ListWords(ctx context.Context, query bson.M, page, limit int, sort bson.M) (settings []*models.WordFilter, pages int, total int64, err error) {
	pages = 1
	skip := 0
	count, err := s.srv.settings.CountDocuments(ctx, query)

	if err != nil {
		return settings, 0, 0, errorWrapper(err)
	}

	// If the count is greater than the perPage variable, then we have more than 1 page!
	if count > int64(limit) {
		pages = int(math.Ceil((float64(count) / float64(limit))))
		skip = (page - 1) * limit
	}

	cursor, err := s.srv.words.Find(ctx, query, options.Find().SetSort(sort).SetSkip(int64(skip)).SetLimit(int64(limit)))
	if err != nil {
		return settings, pages, 0, errorWrapper(err)
	}

	err = cursor.All(ctx, &settings)

	return settings, pages, count, err
}

func (s *filterSrv) ListLegality(ctx context.Context, query bson.M, page, limit int, sort bson.M) (settings []*models.LegalityFilter, pages int, total int64, err error) {
	pages = 1
	skip := 0
	count, err := s.srv.settings.CountDocuments(ctx, query)

	if err != nil {
		return settings, 0, 0, errorWrapper(err)
	}

	// If the count is greater than the perPage variable, then we have more than 1 page!
	if count > int64(limit) {
		pages = int(math.Ceil((float64(count) / float64(limit))))
		skip = (page - 1) * limit
	}

	cursor, err := s.srv.words.Find(ctx, query, options.Find().SetSort(sort).SetSkip(int64(skip)).SetLimit(int64(limit)))
	if err != nil {
		return settings, pages, 0, errorWrapper(err)
	}

	err = cursor.All(ctx, &settings)

	return settings, pages, count, err
}
