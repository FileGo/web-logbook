package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vsimakhin/web-logbook/internal/models"
)

// HandlerAirportByID returns airport record by ID (ICAO or IATA)
func (app *application) HandlerAirportByID(w http.ResponseWriter, r *http.Request) {
	uuid := chi.URLParam(r, "id")

	airport, err := app.db.GetAirportByID(uuid)
	if err != nil {
		app.errorLog.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if airport.IATA == "" && airport.ICAO == "" {
		// looks like there is no such ID in airport database
		app.warningLog.Printf("cannot find %s in the airport database\n", uuid)
	}

	err = app.writeJSON(w, http.StatusOK, airport)
	if err != nil {
		app.errorLog.Println(err)
		return
	}
}

// HandlerAirportUpdate updates the Airports DB
func (app *application) HandlerAirportUpdate(w http.ResponseWriter, r *http.Request) {

	var airportsDB map[string]interface{}
	var airports []models.Airport
	var response models.JSONResponse

	// download the json db from the repo
	resp, err := http.Get("https://github.com/vsimakhin/Airports/raw/master/airports.json")
	if err != nil {
		app.errorLog.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		app.errorLog.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// parse the json
	err = json.Unmarshal(body, &airportsDB)
	if err != nil {
		app.errorLog.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, item := range airportsDB {
		airportItem := item.(map[string]interface{})

		var airport models.Airport

		if value, ok := airportItem["icao"].(string); ok {
			airport.ICAO = value
		}

		if value, ok := airportItem["iata"].(string); ok {
			airport.IATA = value
		}

		if value, ok := airportItem["name"].(string); ok {
			airport.Name = value
		}

		if value, ok := airportItem["city"].(string); ok {
			airport.City = value
		}

		if value, ok := airportItem["country"].(string); ok {
			airport.Country = value
		}

		if value, ok := airportItem["elevation"].(float64); ok {
			airport.Elevation = int(value)
		}

		if value, ok := airportItem["lat"].(float64); ok {
			airport.Lat = value
		}

		if value, ok := airportItem["lon"].(float64); ok {
			airport.Lon = value
		}

		airports = append(airports, airport)
	}

	records, err := app.db.UpdateAirportDB(airports)
	if err != nil {
		app.errorLog.Println(err)
		response.OK = false
		response.Message = err.Error()
	} else {
		response.OK = true
		response.Message = fmt.Sprintf("%d", records)

	}

	err = app.writeJSON(w, http.StatusOK, response)
	if err != nil {
		app.errorLog.Println(err)
		return
	}
}

// HandlerAirportAddCustom adds a new custom airport
func (app *application) HandlerAirportAddCustom(w http.ResponseWriter, r *http.Request) {

	var airport models.Airport
	var response models.JSONResponse

	err := json.NewDecoder(r.Body).Decode(&airport)
	if err != nil {
		app.errorLog.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = app.db.AddCustomAirport(airport)
	if err != nil {
		app.errorLog.Println(fmt.Errorf("error adding a new custom airport - %s", err))
		response.OK = false
		response.Message = err.Error()
	} else {
		response.OK = true
		response.Message = fmt.Sprintf("Custom airport %s has been added", airport.Name)
	}

	err = app.writeJSON(w, http.StatusOK, response)
	if err != nil {
		app.errorLog.Println(err)
		return
	}
}

// HandlerAirportDeleteCustom removes a custom airport
func (app *application) HandlerAirportDeleteCustom(w http.ResponseWriter, r *http.Request) {

	var airport models.Airport
	var response models.JSONResponse

	err := json.NewDecoder(r.Body).Decode(&airport)
	if err != nil {
		app.errorLog.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = app.db.RemoveCustomAirport(airport.Name)
	if err != nil {
		app.errorLog.Println(fmt.Errorf("cannot remove custom airport - %s", err))
		response.OK = false
		response.Message = err.Error()
	} else {
		response.OK = true
		response.Message = fmt.Sprintf("Airport %s has been removed", airport.Name)
	}

	err = app.writeJSON(w, http.StatusOK, response)
	if err != nil {
		app.errorLog.Println(err)
		return
	}
}

// HandlerAirportCustomData generates data for the custom airports table
func (app *application) HandlerAirportCustomData(w http.ResponseWriter, r *http.Request) {

	type TableData struct {
		Data [][]string `json:"data"`
	}

	var tableData TableData

	airports, err := app.db.GetCustomAirports()
	if err != nil {
		app.errorLog.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, item := range airports {
		tableRow := []string{item.Name, item.City, item.Country,
			fmt.Sprintf("%d", item.Elevation), fmt.Sprintf("%f", item.Lat), fmt.Sprintf("%f", item.Lon), "",
		}

		tableData.Data = append(tableData.Data, tableRow)
	}

	err = app.writeJSON(w, http.StatusOK, tableData)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

}

// HandlerAirportDBData generates data for the standard airports table
func (app *application) HandlerAirportDBData(w http.ResponseWriter, r *http.Request) {

	type TableData struct {
		Data [][]string `json:"data"`
	}

	var tableData TableData

	airports, err := app.db.GetStandardAirports()
	if err != nil {
		app.errorLog.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, item := range airports {
		tableRow := []string{item.ICAO, item.IATA, item.Name, item.City, item.Country,
			fmt.Sprintf("%d", item.Elevation), fmt.Sprintf("%f", item.Lat), fmt.Sprintf("%f", item.Lon), "",
		}

		tableData.Data = append(tableData.Data, tableRow)
	}

	err = app.writeJSON(w, http.StatusOK, tableData)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

}
