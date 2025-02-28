package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/FM1337/flagbrew-website-source-code/pkg/helpers"
	"github.com/FM1337/flagbrew-website-source-code/pkg/models"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/lrstanley/pt"
)

type UploadJSONResponse struct {
	Uploaded bool   `json:"uploaded"`
	Approved bool   `json:"approved"`
	Code     string `json:"code"`
	Error    string `json:"error"`
}

type IndividualDownloadJSONResponse struct {
	Filename   string `json:"filename"`
	Generation string `json:"generation"`
	Pokemon    string `json:"pokemon"`
}

type BundleDownloadJSONResponse struct {
	Pokemons []struct {
		Generation   string `json:"generation"`
		Pokemon      string `json:"pokemon"`
		Filename     string `json:"filename"`
		Legal        bool   `json:"legal"`
		DownloadCode string `json:"download_code"`
	} `json:"pokemons"`
	Count int `json:"count"`
}

type BundleSearchJSONResponse struct {
	Bundles []struct {
		DownloadCode  string   `json:"download_code"`
		DownloadCodes []string `json:"download_codes"`
		Pokemons      []struct {
			Base64     string `json:"base_64"`
			Generation string `json:"generation"`
			Legality   bool   `json:"legality"`
		} `json:"pokemons"`
		MinGen  string `json:"min_gen"`
		MaxGen  string `json:"max_gen"`
		IsLegal bool   `json:"is_legal"`
	} `json:"bundles"`
	Page  int `json:"page"`
	Pages int `json:"pages"`
	Total int `json:"total"`
}

// type BundleSearchJSONResponse struct {
// 	Pages   int `json:"pages"`
// 	Results []struct {
// 		Pokemon []struct {
// 			Base64     string `json:"base64"`
// 			Generation string `json:"generation"`
// 			Legal      bool   `json:"legal"`
// 			Code       string `json:"code"`
// 		}
// 		MinGen string `json:"min_gen"`
// 		MaxGen string `json:"max_gen"`
// 		Code   string `json:"code"`
// 		Legal  bool   `json:"legal"`
// 	}
// }

func setupLegacyClient(r *http.Request, url string, headers http.Header, sendBody []byte) (*http.Request, error) {
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		// details to log to sentry
		success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
		if success {
			helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: could not read request body", context)
		}
		return nil, err
	}
	r.Body = ioutil.NopCloser(bytes.NewReader(body))

	if sendBody != nil {
		body = sendBody
	}

	proxyReq, err := http.NewRequest(r.Method, url, bytes.NewReader(body))
	if err != nil {
		// details to log to sentry
		success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
		if success {
			helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: an error was encountered calling NewRequest", context)
		}
		// error to return to end user
		return nil, err
	}
	if headers == nil {
		headers = make(http.Header)
	}
	headers.Add("Flagbrew-V0-Legacy", legacyKey)
	headers.Add("Flagbrew-V0-Legacy-Ip", helpers.GetIP(r, legacyKey))
	headers.Add("source", "Legacy Proxy")
	headers.Add("Content-Type", r.Header.Get("Content-Type"))

	proxyReq.Header = headers

	return proxyReq, nil
}

/* this file is to be removed when legacy support is no longer needed */
/* most code is re-used from the old source, any chances are to clean up or simplify stuff */
func registerPKSMLegacyRoutes(r chi.Router) {
	r.With(httprate.LimitByIP(5, 10*time.Second)).Route("/legality", func(r chi.Router) {
		r.Post("/check", func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, 300000)
			missingHeaders, headers := helpers.GetRequiredData(w, r, "header", []string{"Generation"}, true)
			if missingHeaders {
				return
			}

			// Our script is case sensitive.
			headers["generation"] = headers["Generation"]
			delete(headers, "Generation")

			file, _, err := r.FormFile("pkmn")
			if err != nil {
				// details to log to sentry
				success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header})
				if success {
					helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: could not read pkmn from r.FormFile", context)
				}
				// error to return to end user
				helpers.HttpError(w, r, http.StatusBadRequest, fmt.Errorf("could not read pokemon from request"), false, false, false)
				return
			}

			var buf bytes.Buffer
			_, err = io.Copy(&buf, file)
			if err != nil {
				helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, true)
				return
			}
			legality, err := apiGPSS.GetLegalityInfo(buf.Bytes(), headers)
			if err != nil {
				helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, true)
				return
			}
			w.Write([]byte(strings.Join(legality.Report, "\n")))
		})
	})
}

func registerGPSSLegacyRoutes(r chi.Router) {
	r.Use(gpssDisabled)

	r.With(httprate.LimitByIP(10, 15*time.Second)).Post("/share", func(w http.ResponseWriter, r *http.Request) {

		r.Body = http.MaxBytesReader(w, r.Body, 40000)
		body, err := ioutil.ReadAll(r.Body)

		if err != nil {
			// details to log to sentry
			success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
			if success {
				helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: could not read request body", context)
			}
			// error to return to end user
			helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("could not read request body"), false, false, true)
			return
		}

		url := ""

		r.Body = ioutil.NopCloser(bytes.NewReader(body))
		isBundle := false
		if r.Header.Get("bundle") == "yes" || r.Header.Get("bundle") == "Yes" || r.Header.Get("bundle") == "True" || r.Header.Get("bundle") == "true" {
			url = fmt.Sprintf("http://127.0.0.1:%s/api/v2/gpss/upload/bundle", legacyPort)
			isBundle = true
		} else {
			url = fmt.Sprintf("http://127.0.0.1:%s/api/v2/gpss/upload/pokemon", legacyPort)
		}

		header := make(http.Header)

		if isBundle {
			header.Add("count", r.Header.Get("amount"))
			header.Add("generations", r.Header.Get("Generations"))
		} else {
			header.Add("generation", r.Header.Get("Generation"))
		}

		header.Add("patreon", r.Header.Get("PC"))

		proxyReq, err := setupLegacyClient(r, url, header, nil)
		if err != nil {
			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, false, true)
			return
		}

		httpClient := http.Client{}
		resp, err := httpClient.Do(proxyReq)
		if err != nil {
			// details to log to sentry
			success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
			if success {
				helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: something went wrong when making proxy legacy request", context)
			}
			// error to return to end user
			helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("dunno"), false, false, true)
			return
		}
		body, err = ioutil.ReadAll(resp.Body)

		if err != nil {
			// details to log to sentry
			success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
			if success {
				helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: could not read new request body", context)
			}
			// error to return to end user
			helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("could not read request body"), false, false, true)
			return
		}

		// if the status is 200, then we parse the data we want, otherwise we just return the body
		if resp.StatusCode == 200 {
			respData := UploadJSONResponse{}

			err = json.Unmarshal(body, &respData)
			if err != nil {
				// details to log to sentry
				success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
				if success {
					helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: something went wrong decoding response", context)
				}
				// error to return to end user
				helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("could not read request body"), false, false, true)
				return
			}

			if respData.Uploaded {
				w.WriteHeader(http.StatusCreated)
			}

			if respData.Code != "" {
				w.Write([]byte(respData.Code))

				return
			}

		}
		w.Write(body)
	})
	r.With(httprate.LimitByIP(10, 15*time.Second)).Get("/download/{code}", func(w http.ResponseWriter, r *http.Request) {
		downloadCode := chi.URLParam(r, "code")
		if len(downloadCode) != 10 {
			// details to log to sentry
			success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header})
			if success {
				helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY:invalid download code", context)
			}
			// error to return to end user
			helpers.HttpError(w, r, http.StatusBadRequest, fmt.Errorf("invalid downlaod code"), false, false, false)
			return
		}

		proxyReq, err := setupLegacyClient(r, fmt.Sprintf("http://127.0.0.1:%s/api/v2/gpss/download/pokemon/%s", legacyPort, downloadCode), nil, nil)
		if err != nil {
			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, false, true)
			return
		}

		httpClient := http.Client{}
		resp, err := httpClient.Do(proxyReq)
		if err != nil {
			// details to log to sentry
			success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
			if success {
				helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: something went wrong when making proxy legacy request", context)
			}
			// error to return to end user
			helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("dunno"), false, false, true)
			return
		}
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			// details to log to sentry
			success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
			if success {
				helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: could not read new request body", context)
			}
			// error to return to end user
			helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("could not read request body"), false, false, true)
			return
		}
		// if the status is 200, then we parse the data we want, otherwise we just return the body
		if resp.StatusCode == 200 {
			respData := IndividualDownloadJSONResponse{}

			err = json.Unmarshal(body, &respData)
			if err != nil {
				// details to log to sentry
				success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
				if success {
					helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: something went wrong decoding response", context)
				}
				// error to return to end user
				helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("could not read request body"), false, false, true)
				return
			}

			w.Header().Add("Generation", respData.Generation)
			// _, err := base64.StdEncoding.DecodeString(respData.Pokemon)
			// if err != nil {
			// 	// details to log to sentry
			// 	success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
			// 	if success {
			// 		helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: something went wrong decoding base64 for pokemon", context)
			// 	}
			// 	// error to return to end user
			// 	helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("could not decode base64"), false, false, true)
			// 	return
			// }
			w.Write([]byte(respData.Pokemon))
			return
		}

		w.Write(body)
	})

	r.With(httprate.LimitByIP(10, 15*time.Second)).Get("/download/bundle/{code}", func(w http.ResponseWriter, r *http.Request) {
		downloadCode := chi.URLParam(r, "code")
		if len(downloadCode) != 10 {
			// details to log to sentry
			success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header})
			if success {
				helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY:invalid download code", context)
			}
			// error to return to end user
			helpers.HttpError(w, r, http.StatusBadRequest, fmt.Errorf("invalid downlaod code"), false, false, false)
			return
		}

		proxyReq, err := setupLegacyClient(r, fmt.Sprintf("http://127.0.0.1:%s/api/v2/gpss/download/bundle/%s", legacyPort, downloadCode), nil, nil)
		if err != nil {
			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, false, true)
			return
		}

		httpClient := http.Client{}
		resp, err := httpClient.Do(proxyReq)
		if err != nil {
			// details to log to sentry
			success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
			if success {
				helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: something went wrong when making proxy legacy request", context)
			}
			// error to return to end user
			helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("dunno"), false, false, true)
			return
		}
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			// details to log to sentry
			success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
			if success {
				helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: could not read new request body", context)
			}
			// error to return to end user
			helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("could not read request body"), false, false, true)
			return
		}
		// if the status is 200, then we parse the data we want, otherwise we just return the body
		if resp.StatusCode == 200 {
			respData := BundleDownloadJSONResponse{}

			err = json.Unmarshal(body, &respData)
			if err != nil {
				// details to log to sentry
				success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
				if success {
					helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: something went wrong decoding response", context)
				}
				// error to return to end user
				helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("could not read request body"), false, false, true)
				return
			}

			returnSlice := []pt.M{}

			for _, pk := range respData.Pokemons {
				returnSlice = append(returnSlice, pt.M{"base64": pk.Pokemon, "legal": pk.Legal, "generation": pk.Generation, "code": pk.DownloadCode})
			}

			pt.JSON(w, r, pt.M{"pokemon": returnSlice})
			return
		}

		w.Write(body)
	})
}

func registerAPIV1LegacyRoutes(r chi.Router) {
	r.Route("/bot", func(r chi.Router) {
		r.With(httprate.LimitByIP(5, 10*time.Second)).Post("/auto_legality", func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, 40000)
			body, err := ioutil.ReadAll(r.Body)

			if err != nil {
				// details to log to sentry
				success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
				if success {
					helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: could not read request body", context)
				}
				// error to return to end user
				helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("could not read request body"), false, false, true)
				return
			}

			r.Body = ioutil.NopCloser(bytes.NewReader(body))
			url := fmt.Sprintf("http://127.0.0.1:%s/api/v2/pksm/legalize", legacyPort)

			header := make(http.Header)

			header.Add("version", r.Header.Get("version"))
			header.Add("generation", r.Header.Get("generation"))

			proxyReq, err := setupLegacyClient(r, url, header, nil)
			if err != nil {
				helpers.HttpError(w, r, http.StatusInternalServerError, err, false, false, true)
				return
			}

			httpClient := http.Client{}
			resp, err := httpClient.Do(proxyReq)
			status := resp.StatusCode
			resp.Close = true
			if err != nil {
				// details to log to sentry
				success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err.Error()})
				if success {
					helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: something went wrong when making proxy legacy request", context)
				}
				// error to return to end user
				helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("dunno"), false, false, true)
				return
			}
			body, err = ioutil.ReadAll(resp.Body)

			if err != nil {
				// details to log to sentry
				success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
				if success {
					helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: could not read new request body", context)
				}
				// error to return to end user
				helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("could not read request body"), false, false, true)
				return
			}

			// if status is 200, json is trustable I think
			if status == 200 {
				legality := models.AutoLegalize{}

				err = json.Unmarshal(body, &legality)
				if err != nil {
					// details to log to sentry
					success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
					if success {
						helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: something went wrong decoding response", context)
					}
					// error to return to end user
					helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("could not read request body"), false, false, true)
					return
				}

				pt.JSON(w, r, pt.M{"Pokemon": legality.Pokemon, "Success": legality.Success, "Report": legality.Report, "Ran": legality.Ran})
				return
			}

			w.Write(body)
		})
	})
	r.Route("/gpss", func(r chi.Router) {
		r.With(httprate.LimitByIP(30, 10*time.Second)).Get("/all", func(w http.ResponseWriter, r *http.Request) {
			// get the sort by method
			sort := r.URL.Query().Get("sort")
			if sort == "" {
				sort = "latest"
			}

			// get the direction to sort by
			direction := r.URL.Query().Get("dir")

			sortBool := true
			if direction == "ascend" {
				sortBool = false
			}

			// get the page
			page := r.URL.Query().Get("page")

			// get the option to show legal pokemon only
			legalOnly := r.URL.Query().Get("legal_only")

			legal := false
			if legalOnly == "yes" || legalOnly == "True" || legalOnly == "Yes" || legalOnly == "true" {
				legal = true
			}

			minGen := r.URL.Query().Get("min_gen")
			maxGen := r.URL.Query().Get("max_gen")
			gens := []string{}
			if minGen == "" {
				minGen = "1"
			}

			if maxGen == "" {
				maxGen = "8"
			}

			if test, err := strconv.Atoi(minGen); err != nil || test <= 0 {
				// details to log to sentry
				success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
				if success {
					helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: bad min generation", context)
				}
				// error to return to end user
				helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("invalid min generation provided"), false, false, false)
				return
			}

			if test, err := strconv.Atoi(maxGen); err != nil || test <= 0 {
				// details to log to sentry
				success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
				if success {
					helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: bad max generation", context)
				}
				// error to return to end user
				helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("invalid max generation provided"), false, false, false)
				return
			}

			i, _ := strconv.Atoi(minGen)
			max, _ := strconv.Atoi(maxGen)
			for {
				gens = append(gens, strconv.Itoa(i))

				if i+1 > max {
					break
				}
				i++
			}

			if r.URL.Query().Get("lgpe") == "yes" {
				gens = append(gens, "LGPE")
			}

			// now comes the tough part, translating all this into data that the new searching can use >_<

			operators := []pt.M{
				{"operator": "IN", "field": "generations"},
				{"operator": "=", "field": "legal"},
			}

			postJSON := pt.M{"sort_direction": sortBool, "sort_field": sort, "mode": "and", "generations": gens, "legal": legal, "operators": operators}

			url := fmt.Sprintf("http://127.0.0.1:%s/api/v2/gpss/search/pokemon?page=%s", legacyPort, page)

			headers := make(http.Header)
			// force pksm mode
			headers.Add("pksm-mode", "yes")

			postJSONBytes, err := json.Marshal(postJSON)
			if err != nil {
				// details to log to sentry
				success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
				if success {
					helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: could not marshal json for translationing request", context)
				}
				// error to return to end user
				helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("yeet"), false, false, true)
				return
			}

			// have to override :(
			r.Method = "POST"
			r.Header.Set("Content-Type", "application/json;")
			proxyReq, err := setupLegacyClient(r, url, headers, postJSONBytes)
			if err != nil {
				helpers.HttpError(w, r, http.StatusInternalServerError, err, false, false, true)
				return
			}

			httpClient := http.Client{}
			resp, err := httpClient.Do(proxyReq)
			if err != nil {
				// details to log to sentry
				success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
				if success {
					helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: something went wrong when making proxy legacy request", context)
				}
				// error to return to end user
				helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("dunno"), false, false, true)
				return
			}
			body, err := ioutil.ReadAll(resp.Body)

			if err != nil {
				// details to log to sentry
				success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
				if success {
					helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: could not read new request body", context)
				}
				// error to return to end user
				helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("could not read request body"), false, false, true)
				return
			}
			w.Header().Set("Content-Type", "application/json;")
			w.Write(body)
		})

		r.With(httprate.LimitByIP(30, 10*time.Second)).Get("/bundles/all", func(w http.ResponseWriter, r *http.Request) {
			// get the sort by method
			sort := r.URL.Query().Get("sort")
			if sort == "" {
				sort = "latest"
			}

			// get the direction to sort by
			direction := r.URL.Query().Get("dir")

			sortBool := true
			if direction == "ascend" {
				sortBool = false
			}
			// get the page
			page := r.URL.Query().Get("page")

			// get the option to show legal pokemon only
			legalOnly := r.URL.Query().Get("legal_only")

			legal := false
			if legalOnly == "yes" || legalOnly == "True" || legalOnly == "Yes" || legalOnly == "true" {
				legal = true
			}

			minGen := r.URL.Query().Get("min_gen")
			maxGen := r.URL.Query().Get("max_gen")
			gens := []string{}
			if minGen == "" {
				minGen = "1"
			}

			if maxGen == "" {
				maxGen = "8"
			}

			if test, err := strconv.Atoi(minGen); err != nil || test <= 0 {
				// details to log to sentry
				success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
				if success {
					helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: bad min generation", context)
				}
				// error to return to end user
				helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("invalid min generation provided"), false, false, false)
				return
			}

			if test, err := strconv.Atoi(maxGen); err != nil || test <= 0 {
				// details to log to sentry
				success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
				if success {
					helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: bad max generation", context)
				}
				// error to return to end user
				helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("invalid max generation provided"), false, false, false)
				return
			}

			i, _ := strconv.Atoi(minGen)
			max, _ := strconv.Atoi(maxGen)
			for {
				gens = append(gens, strconv.Itoa(i))

				if i+1 > max {
					break
				}
				i++
			}

			if r.URL.Query().Get("lgpe") == "yes" {
				gens = append(gens, "LGPE")
			}

			// now comes the tough part, translating all this into data that the new searching can use >_<

			operators := []pt.M{
				{"operator": "IN", "field": "generations"},
				{"operator": "=", "field": "legal"},
			}

			postJSON := pt.M{"sort_direction": sortBool, "sort_field": sort, "mode": "and", "generations": gens, "legal": legal, "operators": operators}

			url := fmt.Sprintf("http://127.0.0.1:%s/api/v2/gpss/search/bundles?page=%s", legacyPort, page)

			headers := make(http.Header)
			// force pksm mode
			headers.Add("pksm-mode", "yes")

			postJSONBytes, err := json.Marshal(postJSON)
			if err != nil {
				// details to log to sentry
				success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
				if success {
					helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: could not marshal json for translationing request", context)
				}
				// error to return to end user
				helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("yeet"), false, false, true)
				return
			}

			// have to override :(
			r.Method = "POST"
			r.Header.Set("Content-Type", "application/json;")
			proxyReq, err := setupLegacyClient(r, url, headers, postJSONBytes)
			if err != nil {
				helpers.HttpError(w, r, http.StatusInternalServerError, err, false, false, true)
				return
			}

			httpClient := http.Client{}
			resp, err := httpClient.Do(proxyReq)
			if err != nil {
				// details to log to sentry
				success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
				if success {
					helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: something went wrong when making proxy legacy request", context)
				}
				// error to return to end user
				helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("dunno"), false, false, true)
				return
			}
			body, err := ioutil.ReadAll(resp.Body)

			if err != nil {
				// details to log to sentry
				success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
				if success {
					helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: could not read new request body", context)
				}
				// error to return to end user
				helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("could not read request body"), false, false, true)
				return
			}

			// if the status is 200, then we parse the data we want, otherwise we just return the body
			if resp.StatusCode == 200 {
				respData := BundleSearchJSONResponse{}

				err = json.Unmarshal(body, &respData)
				if err != nil {
					// details to log to sentry
					success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
					if success {
						helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: something went wrong decoding response", context)
					}
					// error to return to end user
					helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("could not read request body"), false, false, true)
					return
				}
				returnSlice := []pt.M{}

				for _, bundle := range respData.Bundles {
					tmpSlice := []pt.M{}
					for i, pk := range bundle.Pokemons {
						tmpSlice = append(tmpSlice, pt.M{"base64": pk.Base64, "generation": pk.Generation, "legal": pk.Legality, "code": bundle.DownloadCodes[i]})
					}
					returnSlice = append(returnSlice, pt.M{"code": bundle.DownloadCode, "pokemon": tmpSlice, "min_gen": bundle.MinGen, "max_gen": bundle.MaxGen, "legal": bundle.IsLegal})
				}

				pt.JSON(w, r, pt.M{"pages": respData.Pages, "total_bundles": respData.Total, "results": returnSlice})
				return
			}
			w.Write(body)
		})
	})
}

func registerMysteryGiftLegacyRoute(r chi.Router) {
	r.With(rateLimitByIp(18, 1*time.Minute)).Get("/{filename}", func(w http.ResponseWriter, r *http.Request) {
		filename := chi.URLParam(r, "filename")
		if filename == "" {
			// details to log to sentry
			success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header})
			if success {
				helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: missing mystery gift filename", context)
			}
			// error to return to end user
			helpers.HttpError(w, r, http.StatusNotFound, fmt.Errorf("could not read request body"), false, false, true)
			return
		}
		url := fmt.Sprintf("http://127.0.0.1:%s/api/v2/files/download/mystery-gift/%s", legacyPort, filename)
		proxyReq, err := setupLegacyClient(r, url, nil, nil)
		if err != nil {
			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, false, true)
			return
		}

		httpClient := http.Client{}
		resp, err := httpClient.Do(proxyReq)
		if err != nil {
			// details to log to sentry
			success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
			if success {
				helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: something went wrong when making proxy legacy request", context)
			}
			// error to return to end user
			helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("dunno"), false, false, true)
			return
		}
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			// details to log to sentry
			success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{r.RemoteAddr, r.URL.Path, r.Header, err})
			if success {
				helpers.LogToSentryWithContext(sentry.LevelError, "LEGACY: could not read new request body", context)
			}
			// error to return to end user
			helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("could not read request body"), false, false, true)
			return
		}

		// set the headers from the new request
		w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))
		w.Header().Set("Content-Disposition", resp.Header.Get("Content-Disposition"))
		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))

		w.Write(body)
	})
}
