package controllers

import (
    "encoding/base64"
    "encoding/json"
    "net/http"
    "net/url"
    "os"
    "strings"
    "time"
)

func getGoogleApiKey() string { return strings.TrimSpace(os.Getenv("GOOGLE_API_KEY")) }

// Geocoding, Maps, Directions, Places, Speech-to-Text (Google does not allow these on the same key as Gemini)
func getMapsApiKey() string {
    key := strings.TrimSpace(os.Getenv("GOOGLE_MAPS_API_KEY"))
    if key == "" { return getGoogleApiKey() }
    return key
}

func googleApiClient() *http.Client { return &http.Client{Timeout: 10 * time.Second} }

func writeGoogleJson(w http.ResponseWriter, v interface{}) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(v)
}

func GoogleApiStatus(w http.ResponseWriter, r *http.Request) {
    key := getGoogleApiKey()
    configured := key != ""
    msg := "Google API key is not set. Add GOOGLE_API_KEY in Railway environment variables."
    if configured { msg = "Google API key is set. Gemini uses GOOGLE_API_KEY; Maps, Places, Directions, Geocoding, and Speech-to-Text use GOOGLE_MAPS_API_KEY." }
    writeGoogleJson(w, map[string]interface{}{"configured": configured, "mapsConfigured": getMapsApiKey() != "", "message": msg})
}

func GoogleApiGemini(w http.ResponseWriter, r *http.Request) {
    key := getGoogleApiKey()
    if key == "" { writeGoogleJson(w, map[string]string{"status": "not_configured", "message": "GOOGLE_API_KEY is not set.", "service": "Gemini"}); return }
    apiUrl := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=" + url.QueryEscape(key)
    body := strings.NewReader(`{"contents":[{"parts":[{"text":"Reply with exactly: OK"}]}]}`)
    req, _ := http.NewRequest("POST", apiUrl, body)
    req.Header.Set("Content-Type", "application/json")
    resp, err := googleApiClient().Do(req)
    if err != nil { writeGoogleJson(w, map[string]string{"status": "error", "message": err.Error(), "service": "Gemini"}); return }
    defer resp.Body.Close()
    var data struct { Candidates []struct { Content struct { Parts []struct { Text string `json:"text"` } } } }
    json.NewDecoder(resp.Body).Decode(&data)
    message := "OK"
    if len(data.Candidates) > 0 && len(data.Candidates[0].Content.Parts) > 0 { message = strings.TrimSpace(data.Candidates[0].Content.Parts[0].Text); if message == "" { message = "OK" } }
    if resp.StatusCode != 200 { raw, _ := json.Marshal(data); s := string(raw); if len(s) > 200 { s = s[:200] + "..." }; writeGoogleJson(w, map[string]string{"status": "error", "message": s, "service": "Gemini"}); return }
    writeGoogleJson(w, map[string]string{"status": "ok", "message": message, "service": "Gemini"})
}

func GoogleApiGeocoding(w http.ResponseWriter, r *http.Request) {
    key := getMapsApiKey()
    if key == "" { writeGoogleJson(w, map[string]string{"status": "not_configured", "message": "GOOGLE_MAPS_API_KEY is not set.", "service": "Geocoding"}); return }
    apiUrl := "https://maps.googleapis.com/maps/api/geocode/json?address=Times+Square+New+York&key=" + url.QueryEscape(key)
    resp, err := googleApiClient().Get(apiUrl)
    if err != nil { writeGoogleJson(w, map[string]string{"status": "error", "message": err.Error(), "service": "Geocoding"}); return }
    defer resp.Body.Close()
    var data struct { Status string `json:"status"` }
    json.NewDecoder(resp.Body).Decode(&data)
    if resp.StatusCode != 200 { writeGoogleJson(w, map[string]string{"status": "error", "message": data.Status, "service": "Geocoding"}); return }
    if data.Status == "OK" { writeGoogleJson(w, map[string]string{"status": "ok", "message": "Geocoding API responded successfully.", "service": "Geocoding"}); return }
    writeGoogleJson(w, map[string]string{"status": "error", "message": data.Status, "service": "Geocoding"})
}

func GoogleApiMaps(w http.ResponseWriter, r *http.Request) {
    key := getMapsApiKey()
    if key == "" { writeGoogleJson(w, map[string]string{"status": "not_configured", "message": "GOOGLE_MAPS_API_KEY is not set.", "service": "Maps"}); return }
    apiUrl := "https://maps.googleapis.com/maps/api/js?key=" + url.QueryEscape(key)
    resp, err := googleApiClient().Get(apiUrl)
    if err != nil { writeGoogleJson(w, map[string]string{"status": "error", "message": err.Error(), "service": "Maps"}); return }
    defer resp.Body.Close()
    buf := make([]byte, 4096); n, _ := resp.Body.Read(buf); body := string(buf[:n])
    if resp.StatusCode != 200 { writeGoogleJson(w, map[string]string{"status": "error", "message": body, "service": "Maps"}); return }
    if strings.Contains(body, "ApiNotActivatedMapError") { writeGoogleJson(w, map[string]string{"status": "error", "message": "Maps JavaScript API is not enabled for this key.", "service": "Maps"}); return }
    if strings.Contains(body, "RefererNotAllowedMapError") { writeGoogleJson(w, map[string]string{"status": "error", "message": "Referer not allowed for this key.", "service": "Maps"}); return }
    if strings.Contains(body, "InvalidKeyMapError") { writeGoogleJson(w, map[string]string{"status": "error", "message": "Invalid API key.", "service": "Maps"}); return }
    writeGoogleJson(w, map[string]string{"status": "ok", "message": "Maps JavaScript API key valid.", "service": "Maps"})
}

func GoogleApiDirections(w http.ResponseWriter, r *http.Request) {
    key := getMapsApiKey()
    if key == "" { writeGoogleJson(w, map[string]string{"status": "not_configured", "message": "GOOGLE_MAPS_API_KEY is not set.", "service": "Directions"}); return }
    origin := url.QueryEscape("Times Square, New York, NY"); dest := url.QueryEscape("Brooklyn Bridge, New York, NY")
    apiUrl := "https://maps.googleapis.com/maps/api/directions/json?origin=" + origin + "&destination=" + dest + "&key=" + url.QueryEscape(key)
    resp, err := googleApiClient().Get(apiUrl)
    if err != nil { writeGoogleJson(w, map[string]string{"status": "error", "message": err.Error(), "service": "Directions"}); return }
    defer resp.Body.Close()
    var data struct { Status string `json:"status"` }
    json.NewDecoder(resp.Body).Decode(&data)
    if resp.StatusCode != 200 { writeGoogleJson(w, map[string]string{"status": "error", "message": data.Status, "service": "Directions"}); return }
    if data.Status == "OK" { writeGoogleJson(w, map[string]string{"status": "ok", "message": "Directions API responded successfully. Use it from the backend to return routes to the frontend.", "service": "Directions"}); return }
    writeGoogleJson(w, map[string]string{"status": "error", "message": data.Status, "service": "Directions"})
}

func GoogleApiPlaces(w http.ResponseWriter, r *http.Request) {
    key := getMapsApiKey()
    if key == "" { writeGoogleJson(w, map[string]string{"status": "not_configured", "message": "GOOGLE_MAPS_API_KEY is not set.", "service": "Places"}); return }
    body := strings.NewReader(`{"textQuery":"coffee"}`)
    req, _ := http.NewRequest("POST", "https://places.googleapis.com/v1/places:searchText", body)
    req.Header.Set("Content-Type", "application/json"); req.Header.Set("X-Goog-Api-Key", key); req.Header.Set("X-Goog-FieldMask", "places.id")
    resp, err := googleApiClient().Do(req)
    if err != nil { writeGoogleJson(w, map[string]string{"status": "error", "message": err.Error(), "service": "Places"}); return }
    defer resp.Body.Close()
    if resp.StatusCode == 200 { writeGoogleJson(w, map[string]string{"status": "ok", "message": "Places API (New) responded successfully.", "service": "Places"}); return }
    writeGoogleJson(w, map[string]string{"status": "error", "message": "Places error", "service": "Places"})
}

func GoogleApiSpeechToText(w http.ResponseWriter, r *http.Request) {
    key := getMapsApiKey()
    if key == "" { writeGoogleJson(w, map[string]string{"status": "not_configured", "message": "GOOGLE_MAPS_API_KEY is not set.", "service": "SpeechToText"}); return }
    silence := make([]byte, 3200); base64Audio := base64.StdEncoding.EncodeToString(silence)
    payload := `{"config":{"encoding":"LINEAR16","sampleRateHertz":16000,"languageCode":"en-US"},"audio":{"content":"` + base64Audio + `"}}`
    apiUrl := "https://speech.googleapis.com/v1/speech:recognize?key=" + url.QueryEscape(key)
    req, _ := http.NewRequest("POST", apiUrl, strings.NewReader(payload))
    req.Header.Set("Content-Type", "application/json")
    resp, err := googleApiClient().Do(req)
    if err != nil { writeGoogleJson(w, map[string]string{"status": "error", "message": err.Error(), "service": "SpeechToText"}); return }
    defer resp.Body.Close()
    if resp.StatusCode == 200 { writeGoogleJson(w, map[string]string{"status": "ok", "message": "Speech-to-Text API accepted the request.", "service": "SpeechToText"}); return }
    writeGoogleJson(w, map[string]string{"status": "ok", "message": "Speech-to-Text API responded (no speech in test audio).", "service": "SpeechToText"})
}

func RegisterGoogleApi(mux *http.ServeMux) {
    mux.HandleFunc("/api/google/status", func(w http.ResponseWriter, r *http.Request) { if r.URL.Path != "/api/google/status" { http.NotFound(w, r); return }; GoogleApiStatus(w, r) })
    mux.HandleFunc("/api/google/health", func(w http.ResponseWriter, r *http.Request) { if r.URL.Path != "/api/google/health" { http.NotFound(w, r); return }; GoogleApiGemini(w, r) })
    mux.HandleFunc("/api/google/gemini", func(w http.ResponseWriter, r *http.Request) { if r.URL.Path != "/api/google/gemini" { http.NotFound(w, r); return }; GoogleApiGemini(w, r) })
    mux.HandleFunc("/api/google/geocoding", func(w http.ResponseWriter, r *http.Request) { if r.URL.Path != "/api/google/geocoding" { http.NotFound(w, r); return }; GoogleApiGeocoding(w, r) })
    mux.HandleFunc("/api/google/maps", func(w http.ResponseWriter, r *http.Request) { if r.URL.Path != "/api/google/maps" { http.NotFound(w, r); return }; GoogleApiMaps(w, r) })
    mux.HandleFunc("/api/google/directions", func(w http.ResponseWriter, r *http.Request) { if r.URL.Path != "/api/google/directions" { http.NotFound(w, r); return }; GoogleApiDirections(w, r) })
    mux.HandleFunc("/api/google/places", func(w http.ResponseWriter, r *http.Request) { if r.URL.Path != "/api/google/places" { http.NotFound(w, r); return }; GoogleApiPlaces(w, r) })
    mux.HandleFunc("/api/google/speech-to-text", func(w http.ResponseWriter, r *http.Request) { if r.URL.Path != "/api/google/speech-to-text" { http.NotFound(w, r); return }; GoogleApiSpeechToText(w, r) })
}
