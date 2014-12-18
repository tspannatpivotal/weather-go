package weather

import (
	"fmt"
	"github.com/benschw/opin-go/rest"
	"github.com/benschw/weather-go/weather/api"
	"github.com/benschw/weather-go/weather/openweather"
	"github.com/jinzhu/gorm"
	"log"
	"net/http"
)

var _ = log.Print

type LocationResource struct {
	Db         gorm.DB
	CondClient openweather.WeatherClient
}

func (r *LocationResource) includeConditions(loc *api.Location) error {
	cond, err := r.CondClient.FindForLocation(loc.City, loc.State)
	if err == nil {
		loc.Conditions = cond
	}
	return err
}

func (r *LocationResource) Add(res http.ResponseWriter, req *http.Request) {
	var location api.Location

	if err := rest.Bind(req, &location); err != nil {
		rest.SetBadRequestResponse(res)
		return
	}

	var found api.Location
	if location.Id != 0 && !r.Db.First(&found, location.Id).RecordNotFound() {
		rest.SetConflictResponse(res)
		return
	}
	location.Id = 0

	r.Db.Save(&location)

	if err := r.includeConditions(&location); err != nil {
		rest.SetInternalServerErrorResponse(res, err)
		return
	}
	if err := rest.SetCreatedResponse(res, location, fmt.Sprintf("location/%d", location.Id)); err != nil {
		rest.SetInternalServerErrorResponse(res, err)
		return
	}
}

func (r *LocationResource) FindAll(res http.ResponseWriter, req *http.Request) {
	var locations []api.Location

	r.Db.Find(&locations)
	for i, _ := range locations {
		r.includeConditions(&locations[i])
	}

	if err := rest.SetOKResponse(res, locations); err != nil {
		rest.SetInternalServerErrorResponse(res, err)
		return
	}
}

func (r *LocationResource) Find(res http.ResponseWriter, req *http.Request) {
	id, err := rest.PathInt(req, "id")
	if err != nil {
		rest.SetBadRequestResponse(res)
		return
	}
	var location api.Location

	if r.Db.First(&location, id).RecordNotFound() {
		rest.SetNotFoundResponse(res)
		return
	}

	if err = r.includeConditions(&location); err != nil {
		rest.SetInternalServerErrorResponse(res, err)
		return
	}
	if err := rest.SetOKResponse(res, location); err != nil {
		rest.SetInternalServerErrorResponse(res, err)
		return
	}
}

func (r *LocationResource) Save(res http.ResponseWriter, req *http.Request) {
	var location api.Location

	id, err := rest.PathInt(req, "id")
	if err != nil {
		rest.SetBadRequestResponse(res)
		return
	}
	if err := rest.Bind(req, &location); err != nil {
		rest.SetBadRequestResponse(res)
		return
	}
	if location.Id != 0 && location.Id != id {
		rest.SetBadRequestResponse(res)
		return
	}
	location.Id = id

	var found api.Location
	if r.Db.First(&found, id).RecordNotFound() {
		rest.SetNotFoundResponse(res)
		return
	}

	r.Db.Save(&location)

	if err = r.includeConditions(&location); err != nil {
		rest.SetInternalServerErrorResponse(res, err)
		return
	}
	if err := rest.SetOKResponse(res, location); err != nil {
		rest.SetInternalServerErrorResponse(res, err)
		return
	}
}

func (r *LocationResource) Delete(res http.ResponseWriter, req *http.Request) {
	id, err := rest.PathInt(req, "id")
	if err != nil {
		rest.SetBadRequestResponse(res)
		return
	}
	var location api.Location

	if r.Db.First(&location, id).RecordNotFound() {
		rest.SetNotFoundResponse(res)
		return
	}

	r.Db.Delete(&location)

	if err = r.includeConditions(&location); err != nil {
		rest.SetInternalServerErrorResponse(res, err)
		return
	}
	if err := rest.SetNoContentResponse(res); err != nil {
		rest.SetInternalServerErrorResponse(res, err)
		return
	}
}
