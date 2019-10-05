package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/DiGregory/tfs-auction/internal/errors"
	"log"
	"strings"
	"encoding/json"
)

type RegUserCase struct {
	Input        string
	Output       error
	OutputStatus int
}

var RegUserTestCases = []RegUserCase{
	{
		Input: `{
			"first_name": "Павел",
				"last_name": "Дуров",
				"birthday":"1905-10-10",
				"email": "durov@telegram7.org",
				"password": "qwerty"
		}`,
		Output: nil, OutputStatus: http.StatusConflict,
	},
	{
		Input: `{
			"first_name": "Павел",
				"last_name": "Дуров",
				"birthday":"1905-10-10",
				 
				"password": "qwerty"
		}`,
		Output: errors.ErrBadReq, OutputStatus: http.StatusBadRequest,
	},
	{
		Input: `{
			"first_name": "",
				"last_name": "Дуров",
				"birthday":"1905-10-10",
				"email": "durov@telegram7.org",
				"password": "qwerty"
		}`,
		Output: errors.ErrBadReq, OutputStatus: http.StatusBadRequest,
	},
	{
		Input: `{
			"first_name": "Дуров",
				"last_name": "",
				"birthday":"1905-10-10",
				"email": "durov@telegram7.org",
				"password": "qwerty"
		}`,
		Output: errors.ErrBadReq, OutputStatus: http.StatusBadRequest,
	},
	{
		Input: `{
			"first_name": "Дуров",
				"last_name": "Дуров",
				"birthday":"",
				"email": "durov@telegram7.org",
				"password": "qwerty"
		}`,
		Output: nil, OutputStatus: http.StatusConflict,
	},
	{
		Input: `{
			"first_name": "Дуров",
				"last_name": "Дуров",
				"birthday":"1995-10-10",
				"email": "",
				"password": "qwerty"
		}`,
		Output: errors.ErrBadReq, OutputStatus: http.StatusBadRequest,
	},
	{
		Input: `{
			"first_name": "Дуров",
				"last_name": "Дуров",
				"birthday":"1995-10-10",
				"email": " durov@telegram7.org",
				"password": ""
		}`,
		Output: errors.ErrBadReq, OutputStatus: http.StatusBadRequest,
	},
}

func TestRegHandler(t *testing.T) {
	DSN := "user=postgres password=1234 dbname=test sslmode=disable"
	a, err := CreateGateWayApp(":5000", DSN, DSN)
	if err != nil {
		log.Fatal("can`t run gateway app")
	}

	t.Run("RegNadlerTest", func(t *testing.T) {
		for i, v := range RegUserTestCases {
			req, err := http.NewRequest(http.MethodPost, "/signup", strings.NewReader(v.Input))
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(a.RegHandler)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != v.OutputStatus {
				t.Errorf("handler returned wrong status code at test [%v]: got %v want %v",
					i, status, v.OutputStatus)
			}
			if v.Output != nil {
				var e = map[string]string{"error": (v.Output.Error())}

				expectResponse, _ := json.Marshal(e)

				if rr.Body.String() != string(expectResponse) {
					t.Errorf("handler returned unexpected body at test [%v]: got %v want %v",
						i, rr.Body.String(), string(expectResponse))
				}
			}
		}
	})

}
