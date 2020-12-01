package tidal

import (
	"fmt"
	"net/http"
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

func (t *Tidal) HandleAuthCallback() {

	go func() {
		http.HandleFunc("/", t.listenerFunc)
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			fmt.Println("listing on port 8080 failed", err)
		}
	}()
}

func (t *Tidal) listenerFunc(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query()["code"]
	if len(code) == 1 && code[0] != "" {
		code := code[0]
		fmt.Println("have code:", code)

		err := t.session.LoginWithOauth2Code(code)

		if err == nil {
			fmt.Fprint(w, "you can close this tab")
			t.CredentialsChan <- true
		} else {
			fmt.Fprint(w, err)
		}
	} else {
		//TODO exit terminal
		fmt.Fprint(w, "sorry, something went wrong")
	}

}
