package resources

import (
	"github.com/gorilla/mux"
	"encoding/json"
	"cooltown/repository"
	"net/http"
	"log"
	"fmt"
	"bytes"
)

func retrieveTrack(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r:= recover(); r != nil {
			fmt.Println("Recovered: ", r)
			w.WriteHeader(500) /* Internal Server Error*/
		}
	}()	
	var sample repository.Sample
	if err := json.NewDecoder(r.Body).Decode(&sample); err != nil {	
		w.WriteHeader(400) /* Bad Request */
		return 
	}
	// Search for track id
	audioBytes, _ := json.Marshal(sample)
	resp, err := http.Post("http://localhost:3001/search", "application/json", bytes.NewBuffer(audioBytes))
	if err != nil {
		panic("retrieveTrack, Error in listening to search microservice :" + err.Error())
	}
	var result repository.Result
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		panic("retrieveTrack, Error in json decoding:" + err.Error())
	}
	tracksResp, tracksErr := http.Get("http://localhost:3000/tracks/" + result.Id)
	if tracksErr != nil {
		panic("retrieveTrack, Error in get request:" + tracksErr.Error())
	}
	var track repository.Track
	if err := json.NewDecoder(tracksResp.Body).Decode(&track); err != nil {	
		panic("retrieveTrack, Error in json decoding:" + err.Error())
	}
	log.Println("Length of track audio: ", len(track.Audio))
	if len(track.Audio) > 0 {
		output := repository.Sample{track.Audio}
		w.WriteHeader(200) /* OK */
		json.NewEncoder(w).Encode(output)
	} else {
		w.WriteHeader(404) /* Not Found */
}

func Router() http.Handler {
	r := mux.NewRouter()
	/* Update */
	r.HandleFunc("/cooltown", retrieveTrack).Methods("POST")
	return r
}
