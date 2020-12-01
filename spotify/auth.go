package spotify

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

/*{
  "access_token": "BQAxVGjohv16P2_rP_vvvAemoNhECqYE7efIuaxPVR5PfcOV3Okat0cdgJEyQ_RUgOIX9EvcQPTkvXF5244sm2miYEbyIalEm6SEIDamEbTB_vn_IpREwyeLr_IBxXounU-nvVNLUZnceq_hN5GelXICs21xcfIQDr7vRA",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "AQB5rNJv_zieu6nQQ-wPkZ_j4cYBKMs8gUzqVMoEcPJlKTsYV_-7jL8h04abU40qmGgwpc37CBfpaAR4TaGxMle43LgN9049T1-3610aTk7tnBTyVGCKK5F2XCsm5xNM778",
  "scope": ""
}*/
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

func (s *Spotify) HandleAuthCallback() {

	go func() {
		http.HandleFunc("/", s.listenerFunc)
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			fmt.Println("listing on port 8080 failed", err)
		}
	}()
}

func (s *Spotify) listenerFunc(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query()["code"]
	if len(code) == 1 && code[0] != "" {
		code := code[0]
		fmt.Println("have code:", code)
		data := url.Values{}
		data.Set("grant_type", "authorization_code")
		data.Set("code", code)
		data.Set("redirect_uri", "http://localhost:8080/")
		data.Set("client_id", s.clientID)
		data.Set("client_secret", s.clientSecret)

		req, _ := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := http.DefaultClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			tokenResponse := new(TokenResponse)
			json.NewDecoder(resp.Body).Decode(tokenResponse) //TODO check error
			s.Credentials = *tokenResponse
			fmt.Printf("token: %s\n", tokenResponse.AccessToken)
			fmt.Fprint(w, "you can close this tab")
			s.CredentialsChan <- true
		} else {
			fmt.Println("error exchanging code: ", err)
			fmt.Fprintf(w, "<div>status: %s</div>", resp.Status)
			fmt.Fprint(w, "<div>")
			fmt.Fprint(w, err)
			fmt.Fprint(w, "</div>")
		}
	} else {
		//TODO exit terminal
		fmt.Fprint(w, "sorry, something went wrong")
	}

}
