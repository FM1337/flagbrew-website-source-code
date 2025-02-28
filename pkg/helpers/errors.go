package helpers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/FM1337/flagbrew-website-source-code/pkg/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/lrstanley/pt"
)

type ArugmentError struct{ Err error }

func (e ArugmentError) Error() string { return e.Err.Error() }

func IsArugmentError(e error) bool {
	if e == nil {
		return false
	}

	if strings.Contains(e.Error(), "System.Argument") {
		return true
	}

	return false
}

// GetRequiredData loops through the required keys and checks to make sure everything is set, if missing data, an error will be returned to the client
func GetRequiredData(w http.ResponseWriter, r *http.Request, dataType string, required []string, logToSentry bool) (missingData bool, returnData map[string]string) {
	returnData = make(map[string]string)
	missingData = false
	errorString := fmt.Sprintf("The following %s data is missing from the %s %s request", dataType, r.Method, r.URL.Path)
	for _, key := range required {
		val := ""
		switch dataType {
		case "header":
			val = r.Header.Get(key)
			break
		case "route":
			val = chi.URLParam(r, key)
			break
		case "post":
			val = r.PostFormValue(key)
			break
		}
		if val == "" {
			errorString += fmt.Sprintf("\n")
			missingData = true
			errorString += fmt.Sprintf("%s", key)
		}

		if !missingData {
			returnData[key] = val
		}
	}

	if missingData {
		HttpError(w, r, http.StatusBadRequest, fmt.Errorf(errorString), true, logToSentry, false)
		return missingData, nil
	}
	return missingData, returnData
}

func validatorErrorFormat(err error) (string, bool) {
	errs := err.(validator.ValidationErrors)
	var rerr string
	if len(errs) > 1 {
		rerr += "Multiple search parameters have failed validation"
		for _, e := range errs {
			rerr += fmt.Sprintf("\n%s failed the '%s' validation rule", e.Field(), e.Tag())
		}
	} else {
		rerr = fmt.Sprintf("%s failed the '%s' validation rule", errs[0].Field(), errs[0].Tag())
	}
	return rerr, len(errs) > 1
}

func HttpError(w http.ResponseWriter, r *http.Request, statusCode int, err error, multipleErr, logToSentry, simpleEndUserError bool) bool {
	if statusCode == http.StatusNotFound && err == nil {
		err = errors.New("The requested resource was not found")
	}

	if err == nil {
		return false
	}

	if statusCode == http.StatusInternalServerError && models.IsClientError(err) {
		if models.IsNotFound(err) {
			statusCode = http.StatusNotFound
		} else {
			// if it's internal server error, override since we know it's a client error.
			statusCode = http.StatusBadRequest
		}
	}
	if logToSentry {
		LogToSentry(err)
	}
	w.WriteHeader(statusCode)

	if strings.HasPrefix(r.URL.Path, "/api/") {
		if simpleEndUserError {
			pt.JSON(w, r, pt.M{"error": "uh oh, something went wrong, please check your request and try again (or try waiting a bit)", "code": statusCode})
		} else {
			if multipleErr {
				errs := strings.Split(err.Error(), "\n")
				pt.JSON(w, r, pt.M{"error": errs[0], "errors": errs[1:], "type": http.StatusText(statusCode), "code": statusCode})
			} else {
				pt.JSON(w, r, pt.M{"error": err.Error(), "type": http.StatusText(statusCode), "code": statusCode})
			}
		}
	} else {
		if simpleEndUserError {
			http.Error(w, fmt.Sprintf("Something went wrong, please check your request try again, or try waiting a bit before trying again. Status Code: %s", http.StatusText(statusCode)), statusCode)
		} else {
			http.Error(w, http.StatusText(statusCode)+": "+err.Error(), statusCode)
		}
	}
	return true
}

func PanicIfErr(err error) {
	// Should be caught by our recoverer middleware.
	if err != nil {
		panic(err)
	}
}
