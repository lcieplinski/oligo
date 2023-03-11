package resources

import (
	"github.com/gorilla/mux"
	"github.com/AudDMusic/audd-go"
	"encoding/json"
	"encoding/base64"
	"search/repository"
	"net/http"
	"log"
	"fmt"
	"os"
)

func searchInAudd(sample string) repository.Metadata {
	data, errD := base64.StdEncoding.DecodeString(sample)
 	if errD != nil {
		panic("searchInAudd, Error in base64 decoding:" + errD.Error())
	}
	if err := os.WriteFile("test.wav", data, 0644); err != nil {
		panic("searchInAudd, Error in writing file:" + err.Error())
	}
	client := audd.NewClient("0257e24289f997380cf3f46b8b53c27a")
	file, err := os.Open("test.wav")
    	if err != nil {
        	panic("searchInAudd, Error in opening wav file:" + err.Error())
    	}
    	result, err := client.Recognize(file, "apple_music,spotify", nil)
    	if err != nil {
        	panic("searchInAudd, Error occurred from Audd.io API:" + err.Error())
    	}
	p := repository.Metadata{Id: "", Title: result.Title, Artist: result.Artist}
	return p
}

func searchTrack(w http.ResponseWriter, r *http.Request) {
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
	result := searchInAudd(sample.Audio)
	trackId, found  := matchMetadata(result)
	if len(trackId) > 0 && found {
		output := repository.Result{Id: trackId}
		w.WriteHeader(200) /* OK */
		json.NewEncoder(w).Encode(output)
	} else {
		w.WriteHeader(404) /* Not Found */
	}
}

func matchMetadata(sampleMetadata repository.Metadata) (string, bool) {
	// get list of files from tracks service and store in array
	resp, err := http.Get("http://localhost:3000/tracks")
	if err != nil {
		panic("matchMetadata, Error in listening to tracks microservice:" + err.Error())
	}
	var list [] string
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		panic("matchMetadata, Error in json decoding:" + err.Error())
	}
	log.Println("List is: ", list)
   	// loop through each ID
	for i, id := range list {
  	// request track from tracks service
		resp, err := http.Get("http://localhost:3000/tracks/" + id)
		if err != nil {
			panic("matchMetadata, Error in get request:" + err.Error())
		}
		var track repository.Track
		if err := json.NewDecoder(resp.Body).Decode(&track); err != nil {
			panic("matchMetadata, Error in json decoding :" + err.Error())
		}
		log.Println("Id is: ", track.Id)
		// call searchInAudd with this track
		result := searchInAudd(track.Audio)
   		// compare returned metadata to sampleMetadata and if they match return id of that track
		if result == sampleMetadata {
			fmt.Printf("Track %d matched\n", i)
			return track.Id, true
		} else {
			fmt.Printf("No match for %d\n", i)
		}		
	}   	
	// if we get to the end and no match is found, then false is returned
	return "", false
}	

func Router() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/search", searchTrack).Methods("POST")
	return r
}
