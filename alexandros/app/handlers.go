package app

import (
	"encoding/json"
	"github.com/kpango/glg"
	"github.com/odysseia-greek/plato/aristoteles/configs"
	"github.com/odysseia-greek/plato/helpers"
	"github.com/odysseia-greek/plato/middleware"
	"github.com/odysseia-greek/plato/models"
	"net/http"
)

type AlexandrosHandler struct {
	Config *configs.AlexandrosConfig
}

// PingPong pongs the ping
func (a *AlexandrosHandler) pingPong(w http.ResponseWriter, req *http.Request) {
	pingPong := models.ResultModel{Result: "pong"}
	middleware.ResponseWithJson(w, pingPong)
}

// returns the health of the api
func (a *AlexandrosHandler) health(w http.ResponseWriter, req *http.Request) {
	health := helpers.GetHealthOfApp(a.Config.Elastic)
	if !health.Healthy {
		middleware.ResponseWithCustomCode(w, 502, health)
		return
	}

	middleware.ResponseWithJson(w, health)
}

// Search a word based on part of that word
func (a *AlexandrosHandler) searchWord(w http.ResponseWriter, req *http.Request) {
	queryWord := req.URL.Query().Get("word")

	var searchResults []models.Meros

	if queryWord == "" {
		e := models.ValidationError{
			ErrorModel: models.ErrorModel{UniqueCode: middleware.CreateGUID()},
			Messages: []models.ValidationMessages{
				{
					Field:   "word",
					Message: "cannot be empty",
				},
			},
		}
		middleware.ResponseWithJson(w, e)
		return
	}

	glg.Debugf("looking for %s", queryWord)

	query := a.Config.Elastic.Builder().MultiMatchWithGram(queryWord)
	response, err := a.Config.Elastic.Query().Match(a.Config.Index, query)

	if err != nil {
		e := models.ElasticSearchError{
			ErrorModel: models.ErrorModel{UniqueCode: middleware.CreateGUID()},
			Message: models.ElasticErrorMessage{
				ElasticError: err.Error(),
			},
		}
		middleware.ResponseWithJson(w, e)
		return
	}

	if len(response.Hits.Hits) == 0 {
		e := models.NotFoundError{
			ErrorModel: models.ErrorModel{UniqueCode: middleware.CreateGUID()},
			Message: models.NotFoundMessage{
				Type:   queryWord,
				Reason: "produced 0 results",
			},
		}
		middleware.ResponseWithJson(w, e)
		return
	}

	for _, hit := range response.Hits.Hits {
		jsonHit, _ := json.Marshal(hit.Source)
		meros, _ := models.UnmarshalMeros(jsonHit)
		if meros.Original != "" {
			meros.Greek = meros.Original
			meros.Original = ""
		}
		searchResults = append(searchResults, meros)
	}

	middleware.ResponseWithJson(w, searchResults)
}
