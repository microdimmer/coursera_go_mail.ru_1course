package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

var (
	ts         = httptest.NewServer(http.HandlerFunc(SearchServer))
	accesToken = "fjrejglkjnvsfjkdhil43095u43SDF"
)

const filePath string = "./dataset.xml"

type DataSetStruct struct {
	XMLName xml.Name `xml:"root"`
	Text    string   `xml:",chardata"`
	Row     []struct {
		Text          string `xml:",chardata"`
		ID            string `xml:"id"`
		GUID          string `xml:"guid"`
		IsActive      string `xml:"isActive"`
		Balance       string `xml:"balance"`
		Picture       string `xml:"picture"`
		Age           string `xml:"age"`
		EyeColor      string `xml:"eyeColor"`
		FirstName     string `xml:"first_name"`
		LastName      string `xml:"last_name"`
		Gender        string `xml:"gender"`
		Company       string `xml:"company"`
		Email         string `xml:"email"`
		Phone         string `xml:"phone"`
		Address       string `xml:"address"`
		About         string `xml:"about"`
		Registered    string `xml:"registered"`
		FavoriteFruit string `xml:"favoriteFruit"`
	} `xml:"row"`
}

type TestCase struct {
	Request *SearchRequest
	Result  *ResultResp
}

type ResultResp struct {
	Resp *SearchResponse
	Err  error
}

func SearchServer(w http.ResponseWriter, request *http.Request) {
	token := request.Header.Get("AccessToken") //request.URL.Query().Get("query")
	if token != accesToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	switch request.URL.Path {
	case "/timeout":
		time.Sleep(2 * time.Second)
		return
	case "/internalerror":
		w.WriteHeader(http.StatusInternalServerError)
		return
	case "/unknownerror":
		w.WriteHeader(http.StatusFound)
		return
	case "/badjson":
		w.Write([]byte("bad json"))
		return
	case "/baderrorjson":
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad json"))
		return
	case "/unknownbadrequest":
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": ""}`))
		return
	}

	orderField := request.FormValue("order_field")
	if orderField != `Id` && orderField != `Age` && orderField != `Name` && orderField != `` { //`Id`, `Age`, `Name`,
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "ErrorBadOrderField"}`))
		return
	}

	orderBy, err := strconv.Atoi(request.FormValue("order_by"))
	if err != nil || orderBy != OrderByAsc && orderBy != OrderByAsIs && orderBy != OrderByDesc { //`Id`, `Age`, `Name`,
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "ErrorBadOrderField"}`))
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	dataSet := DataSetStruct{}
	err = xml.Unmarshal(fileContents, &dataSet)
	if err != nil {
		panic(err)
	}

	query := request.FormValue("query")
	usersSl := []User{}
	for _, record := range dataSet.Row {
		name := fmt.Sprintf("%s %s", record.FirstName, record.LastName)
		if strings.Contains(record.About, query) || strings.Contains(name, query) || query == `` {
			id, _ := strconv.Atoi(record.ID)
			age, _ := strconv.Atoi(record.Age)
			usersSl = append(usersSl, User{id, name, age, record.About, record.Gender})
		}
	}

	if orderBy != OrderByAsIs {
		switch orderField {
		case `Id`:
			sort.SliceStable(usersSl, func(i, j int) bool {
				if orderBy == OrderByAsc {
					return (usersSl[i].ID < usersSl[j].ID)
				}
				return (usersSl[i].ID > usersSl[j].ID)
			})
		case `Age`:
			sort.SliceStable(usersSl, func(i, j int) bool {
				if orderBy == OrderByAsc {
					return (usersSl[i].Age < usersSl[j].Age)
				}
				return (usersSl[i].Age > usersSl[j].Age)
			})
		case `Name`, ``:
			sort.SliceStable(usersSl, func(i, j int) bool {
				if orderBy == OrderByAsc {
					return (usersSl[i].Name < usersSl[j].Name)
				}
				return (usersSl[i].Name > usersSl[j].Name)
			})
		}
	}

	// offset, err := strconv.Atoi(request.FormValue("offset"))
	if offset, err := strconv.Atoi(request.FormValue("offset")); err != nil || offset < 0 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if offset <= len(usersSl) {
		usersSl = usersSl[offset:len(usersSl)]
	} else {
		usersSl = []User{}
	}

	//limit, err := strconv.Atoi(request.FormValue("limit"))
	if limit, err := strconv.Atoi(request.FormValue("limit")); err != nil || limit <= 0 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if limit <= len(usersSl) {
		usersSl = usersSl[0:limit]
	}
	result, err := json.Marshal(usersSl)
	if err != nil {
		panic(err)
	}
	w.Write(result)
}

func TestDataFetch(t *testing.T) {
	testCase := TestCase{
		Request: &SearchRequest{
			Limit:      100,
			Offset:     0,
			Query:      "An",
			OrderField: "",         //`Id`, `Age`, `Name`, default by Name
			OrderBy:    OrderByAsc, //	OrderByAsc  = -1	OrderByAsIs = 0 	OrderByDesc = 1
		},
		Result: &ResultResp{
			Resp: &SearchResponse{
				Users: []User{
					User{
						ID:     16,
						Name:   "Annie Osborn",
						Age:    35,
						About:  "Consequat fugiat veniam commodo nisi nostrud culpa pariatur. Aliquip velit adipisicing dolor et nostrud. Eu nostrud officia velit eiusmod ullamco duis eiusmod ad non do quis.\n",
						Gender: "female",
					},
					User{
						ID:     28,
						Name:   "Cohen Hines",
						Age:    32,
						About:  "Deserunt deserunt dolor ex pariatur dolore sunt labore minim deserunt. Tempor non et officia sint culpa quis consectetur pariatur elit sunt. Anim consequat velit exercitation eiusmod aute elit minim velit. Excepteur nulla excepteur duis eiusmod anim reprehenderit officia est ea aliqua nisi deserunt officia eiusmod. Officia enim adipisicing mollit et enim quis magna ea. Officia velit deserunt minim qui. Commodo culpa pariatur eu aliquip voluptate culpa ullamco sit minim laboris fugiat sit.\n",
						Gender: "male",
					},
					User{
						ID:     24,
						Name:   "Gonzalez Anderson",
						Age:    33,
						About:  "Quis consequat incididunt in ex deserunt minim aliqua ea duis. Culpa nisi excepteur sint est fugiat cupidatat nulla magna do id dolore laboris. Aute cillum eiusmod do amet dolore labore commodo do pariatur sit id. Do irure eiusmod reprehenderit non in duis sunt ex. Labore commodo labore pariatur ex minim qui sit elit.\n",
						Gender: "male",
					},
					User{
						ID:     26,
						Name:   "Sims Cotton",
						Age:    39,
						About:  "Ex cupidatat est velit consequat ad. Tempor non cillum labore non voluptate. Et proident culpa labore deserunt ut aliquip commodo laborum nostrud. Anim minim occaecat est est minim.\n",
						Gender: "male",
					},
				},
				NextPage: false,
			},
			Err: nil,
		},
	}

	s := &SearchClient{
		AccessToken: accesToken,
		URL:         ts.URL,
	}
	result, err := s.FindUsers(*testCase.Request)

	if err != nil {
		t.Errorf("error %v", err.Error())
	}
	if result == nil {
		t.Errorf("error result empty")
	}
	if !reflect.DeepEqual(result, testCase.Result.Resp) && result.NextPage == testCase.Result.Resp.NextPage {
		t.Errorf("results didtn match got:\n %v\n needed:\n %v", result, testCase.Result.Resp)
	}
	// fmt.Printf("result %v\n", result)
	// if err.Error() != testCase.Result.Err.Error() {
	// t.Errorf("error results didtn match got:\n %v\n needed:\n %v", err.Error(), testCase.Result.Err.Error())
	// if result != nil || err.Error() != testCase.Result.Err.Error() {
	// }
}

func TestDataFetchLimit(t *testing.T) {
	testCase := TestCase{
		Request: &SearchRequest{
			Limit:      2,
			Offset:     0,
			Query:      "J",
			OrderField: "Id",        //`Id`, `Age`, `Name`,
			OrderBy:    OrderByDesc, //	OrderByAsc  = -1	OrderByAsIs = 0 	OrderByDesc = 1
		},
		Result: &ResultResp{
			Resp: &SearchResponse{
				Users: []User{
					User{
						ID:     25,
						Name:   "Katheryn Jacobs",
						Age:    32,
						About:  "Magna excepteur anim amet id consequat tempor dolor sunt id enim ipsum ea est ex. In do ea sint qui in minim mollit anim est et minim dolore velit laborum. Officia commodo duis ut proident laboris fugiat commodo do ex duis consequat exercitation. Ad et excepteur ex ea exercitation id fugiat exercitation amet proident adipisicing laboris id deserunt. Commodo proident laborum elit ex aliqua labore culpa ullamco occaecat voluptate voluptate laboris deserunt magna.\n",
						Gender: "female",
					},
					User{
						ID:     21,
						Name:   "Johns Whitney",
						Age:    26,
						About:  "Elit sunt exercitation incididunt est ea quis do ad magna. Commodo laboris nisi aliqua eu incididunt eu irure. Labore ullamco quis deserunt non cupidatat sint aute in incididunt deserunt elit velit. Duis est mollit veniam aliquip. Nulla sunt veniam anim et sint dolore.\n",
						Gender: "male",
					},
				},
				NextPage: true,
			},
			Err: nil,
		},
	}

	s := &SearchClient{
		AccessToken: accesToken,
		URL:         ts.URL,
	}
	result, err := s.FindUsers(*testCase.Request)

	if err != nil {
		t.Errorf("error %v", err.Error())
	}
	if result == nil {
		t.Errorf("error result empty")
	}
	if !reflect.DeepEqual(result, testCase.Result.Resp) && (result.NextPage != testCase.Result.Resp.NextPage) {
		t.Errorf("results didtn match got:\n %v\n needed:\n %v", result, testCase.Result.Resp)
	}
	// fmt.Printf("result %v\n", result)
	// if err.Error() != testCase.Result.Err.Error() {
	// t.Errorf("error results didtn match got:\n %v\n needed:\n %v", err.Error(), testCase.Result.Err.Error())
	// if result != nil || err.Error() != testCase.Result.Err.Error() {
	// }
}

func TestDataFetchOffset(t *testing.T) {
	testCase := TestCase{
		Request: &SearchRequest{
			Limit:      2,
			Offset:     1,
			Query:      "J",
			OrderField: "Id",        //`Id`, `Age`, `Name`,
			OrderBy:    OrderByDesc, //	OrderByAsc  = -1	OrderByAsIs = 0 	OrderByDesc = 1
		},
		Result: &ResultResp{
			Resp: &SearchResponse{
				Users: []User{
					User{
						ID:     21,
						Name:   "Johns Whitney",
						Age:    26,
						About:  "Elit sunt exercitation incididunt est ea quis do ad magna. Commodo laboris nisi aliqua eu incididunt eu irure. Labore ullamco quis deserunt non cupidatat sint aute in incididunt deserunt elit velit. Duis est mollit veniam aliquip. Nulla sunt veniam anim et sint dolore.\n",
						Gender: "male",
					},
					User{
						ID:     8,
						Name:   "Glenn Jordan",
						Age:    29,
						About:  "Duis reprehenderit sit velit exercitation non aliqua magna quis ad excepteur anim. Eu cillum cupidatat sit magna cillum irure occaecat sunt officia officia deserunt irure. Cupidatat dolor cupidatat ipsum minim consequat Lorem adipisicing. Labore fugiat cupidatat nostrud voluptate ea eu pariatur non. Ipsum quis occaecat irure amet esse eu fugiat deserunt incididunt Lorem esse duis occaecat mollit.\n",
						Gender: "male",
					},
				},
				NextPage: true,
			},
			Err: nil,
		},
	}

	s := &SearchClient{
		AccessToken: accesToken,
		URL:         ts.URL,
	}
	result, err := s.FindUsers(*testCase.Request)

	if err != nil {
		t.Errorf("error %v", err.Error())
	}
	if result == nil {
		t.Errorf("error result empty")
	}
	if !reflect.DeepEqual(result, testCase.Result.Resp) && (result.NextPage != testCase.Result.Resp.NextPage) {
		t.Errorf("results didtn match got:\n %v\n needed:\n %v", result, testCase.Result.Resp)
	}
}

func TestDataFetchOrderByAsc(t *testing.T) {
	testCase := TestCase{
		Request: &SearchRequest{
			Limit:      2,
			Offset:     1,
			Query:      "J",
			OrderField: "Id",       //`Id`, `Age`, `Name`,
			OrderBy:    OrderByAsc, //	OrderByAsc  = -1	OrderByAsIs = 0 	OrderByDesc = 1
		},
		Result: &ResultResp{
			Resp: &SearchResponse{
				Users: []User{
					User{
						ID:     8,
						Name:   "Glenn Jordan",
						Age:    29,
						About:  "Duis reprehenderit sit velit exercitation non aliqua magna quis ad excepteur anim. Eu cillum cupidatat sit magna cillum irure occaecat sunt officia officia deserunt irure. Cupidatat dolor cupidatat ipsum minim consequat Lorem adipisicing. Labore fugiat cupidatat nostrud voluptate ea eu pariatur non. Ipsum quis occaecat irure amet esse eu fugiat deserunt incididunt Lorem esse duis occaecat mollit.\n",
						Gender: "male",
					},
					User{
						ID:     21,
						Name:   "Johns Whitney",
						Age:    26,
						About:  "Elit sunt exercitation incididunt est ea quis do ad magna. Commodo laboris nisi aliqua eu incididunt eu irure. Labore ullamco quis deserunt non cupidatat sint aute in incididunt deserunt elit velit. Duis est mollit veniam aliquip. Nulla sunt veniam anim et sint dolore.\n",
						Gender: "male",
					},
				},
				NextPage: true,
			},
			Err: nil,
		},
	}

	s := &SearchClient{
		AccessToken: accesToken,
		URL:         ts.URL,
	}
	result, err := s.FindUsers(*testCase.Request)

	if err != nil {
		t.Errorf("error %v", err.Error())
	}
	if result == nil {
		t.Errorf("error result empty")
	}
	if !reflect.DeepEqual(result, testCase.Result.Resp) && (result.NextPage != testCase.Result.Resp.NextPage) {
		t.Errorf("results didtn match got:\n %v\n needed:\n %v", result, testCase.Result.Resp)
	}
}

func TestDataFetchOrderField(t *testing.T) {
	testCase := TestCase{
		Request: &SearchRequest{
			Limit:      2,
			Offset:     1,
			Query:      "G",
			OrderField: "Age",       //`Id`, `Age`, `Name`,
			OrderBy:    OrderByDesc, //	OrderByAsc  = -1	OrderByAsIs = 0 	OrderByDesc = 1
		},
		Result: &ResultResp{
			Resp: &SearchResponse{
				Users: []User{
					User{
						ID:     24,
						Name:   "Gonzalez Anderson",
						Age:    33,
						About:  "Quis consequat incididunt in ex deserunt minim aliqua ea duis. Culpa nisi excepteur sint est fugiat cupidatat nulla magna do id dolore laboris. Aute cillum eiusmod do amet dolore labore commodo do pariatur sit id. Do irure eiusmod reprehenderit non in duis sunt ex. Labore commodo labore pariatur ex minim qui sit elit.\n",
						Gender: "male",
					},
					User{
						ID:     11,
						Name:   "Gilmore Guerra",
						Age:    32,
						About:  "Labore consectetur do sit et mollit non incididunt. Amet aute voluptate enim et sit Lorem elit. Fugiat proident ullamco ullamco sint pariatur deserunt eu nulla consectetur culpa eiusmod. Veniam irure et deserunt consectetur incididunt ad ipsum sint. Consectetur voluptate adipisicing aute fugiat aliquip culpa qui nisi ut ex esse ex. Sint et anim aliqua pariatur.\n",
						Gender: "male",
					},
				},
				NextPage: true,
			},
			Err: nil,
		},
	}

	s := &SearchClient{
		AccessToken: accesToken,
		URL:         ts.URL,
	}
	result, err := s.FindUsers(*testCase.Request)

	if err != nil {
		t.Errorf("error %v", err.Error())
	}
	if result == nil {
		t.Errorf("error result empty")
	}
	if !reflect.DeepEqual(result, testCase.Result.Resp) && (result.NextPage != testCase.Result.Resp.NextPage) {
		t.Errorf("results didtn match got:\n %v\n needed:\n %v", result, testCase.Result.Resp)
	}
}

func TestNegativeOffset(t *testing.T) {
	testCase := TestCase{
		Request: &SearchRequest{
			Limit:      5,
			Offset:     -10,
			Query:      "",
			OrderField: "",
			OrderBy:    OrderByAsc,
		},
		Result: &ResultResp{
			Resp: nil,
			Err:  fmt.Errorf("offset must be > 0"),
		},
	}

	s := &SearchClient{
		AccessToken: accesToken,
		URL:         ts.URL,
	}
	_, err := s.FindUsers(*testCase.Request)
	if err.Error() != testCase.Result.Err.Error() {
		t.Errorf("error results didtn match got:\n %v\n needed:\n %v", err.Error(), testCase.Result.Err.Error())
	}
}

func TestNegativeLimit(t *testing.T) {
	testCase := TestCase{
		Request: &SearchRequest{
			Limit:      -5,
			Offset:     10,
			Query:      "",
			OrderField: "",
			OrderBy:    OrderByAsc,
		},
		Result: &ResultResp{
			Resp: nil,
			Err:  fmt.Errorf("limit must be > 0"),
		},
	}
	s := &SearchClient{
		AccessToken: accesToken,
		URL:         ts.URL,
	}
	_, err := s.FindUsers(*testCase.Request)
	if err.Error() != testCase.Result.Err.Error() {
		t.Errorf("error results didtn match got:\n %v\n needed:\n %v", err.Error(), testCase.Result.Err.Error())
	}
}

func TestTimeout(t *testing.T) {
	testCase := TestCase{
		Request: &SearchRequest{
			Limit:      10,
			Offset:     10,
			Query:      "",
			OrderField: "",
			OrderBy:    OrderByAsc,
		},
		Result: &ResultResp{
			Resp: nil,
			Err:  fmt.Errorf("timeout for limit=11&offset=10&order_by=-1&order_field=&query="),
		},
	}
	s := &SearchClient{
		AccessToken: accesToken,
		URL:         ts.URL + "/timeout",
	}
	_, err := s.FindUsers(*testCase.Request)
	if err.Error() != testCase.Result.Err.Error() {
		t.Errorf("error results didtn match got:\n %v\n needed:\n %v", err.Error(), testCase.Result.Err.Error())
	}
}

func TestFatalError(t *testing.T) {
	testCase := TestCase{
		Request: &SearchRequest{
			Limit:      10,
			Offset:     10,
			Query:      "",
			OrderField: "",
			OrderBy:    OrderByAsc,
		},
		Result: &ResultResp{
			Resp: nil,
			Err:  fmt.Errorf("SearchServer fatal error"),
		},
	}
	s := &SearchClient{
		AccessToken: accesToken,
		URL:         ts.URL + "/internalerror",
	}
	_, err := s.FindUsers(*testCase.Request)
	if err.Error() != testCase.Result.Err.Error() {
		t.Errorf("error results didtn match got:\n %v\n needed:\n %v", err.Error(), testCase.Result.Err.Error())
	}
}

func TestBadAccesToken(t *testing.T) {
	testCase := TestCase{
		Request: &SearchRequest{
			Limit:      10,
			Offset:     10,
			Query:      "",
			OrderField: "",
			OrderBy:    OrderByAsc,
		},
		Result: &ResultResp{
			Resp: nil,
			Err:  fmt.Errorf("Bad AccessToken"),
		},
	}
	s := &SearchClient{
		AccessToken: accesToken + "asd",
		URL:         ts.URL,
	}
	_, err := s.FindUsers(*testCase.Request)
	if err.Error() != testCase.Result.Err.Error() {
		t.Errorf("error results didtn match got:\n %v\n needed:\n %v", err.Error(), testCase.Result.Err.Error())
	}
}

func TestBadJson(t *testing.T) {
	testCase := TestCase{
		Request: &SearchRequest{
			Limit:      10,
			Offset:     10,
			Query:      "",
			OrderField: "",
			OrderBy:    OrderByAsc,
		},
		Result: &ResultResp{
			Resp: nil,
			Err:  fmt.Errorf("cant unpack result json:"),
		},
	}
	s := &SearchClient{
		AccessToken: accesToken,
		URL:         ts.URL + "/badjson",
	}
	_, err := s.FindUsers(*testCase.Request)
	if !strings.Contains(err.Error(), testCase.Result.Err.Error()) {
		t.Errorf("error results didtn match got:\n %v\n needed:\n %v", err.Error(), testCase.Result.Err.Error())
	}
}

func TestBadErrorJson(t *testing.T) {
	testCase := TestCase{
		Request: &SearchRequest{
			Limit:      10,
			Offset:     10,
			Query:      "",
			OrderField: "",
			OrderBy:    OrderByAsc,
		},
		Result: &ResultResp{
			Resp: nil,
			Err:  fmt.Errorf("cant unpack error json: "),
		},
	}
	s := &SearchClient{
		AccessToken: accesToken,
		URL:         ts.URL + "/baderrorjson",
	}
	_, err := s.FindUsers(*testCase.Request)
	if err != testCase.Result.Err {
		if !strings.Contains(err.Error(), testCase.Result.Err.Error()) {
			t.Errorf("error results didtn match got:\n %v\n needed:\n %v", err, testCase.Result.Err)
		}
	}
}

func TestBadOrderField(t *testing.T) {
	testCase := TestCase{
		Request: &SearchRequest{
			Limit:      10,
			Offset:     10,
			Query:      "",
			OrderField: "asd",
			OrderBy:    OrderByAsc,
		},
		Result: &ResultResp{
			Resp: nil,
			Err:  fmt.Errorf("OrderFeld asd invalid"),
		},
	}
	s := &SearchClient{
		AccessToken: accesToken,
		URL:         ts.URL,
	}
	_, err := s.FindUsers(*testCase.Request)
	if !strings.Contains(err.Error(), testCase.Result.Err.Error()) {
		t.Errorf("error results didtn match got:\n %v\n needed:\n %v", err.Error(), testCase.Result.Err.Error())
	}
}

func TestBadOrderBy(t *testing.T) {
	testCase := TestCase{
		Request: &SearchRequest{
			Limit:      10,
			Offset:     10,
			Query:      "",
			OrderField: "Name",
			OrderBy:    10,
		},
		Result: &ResultResp{
			Resp: nil,
			Err:  fmt.Errorf("OrderFeld Name invalid"),
		},
	}
	s := &SearchClient{
		AccessToken: accesToken,
		URL:         ts.URL,
	}
	_, err := s.FindUsers(*testCase.Request)
	if err != testCase.Result.Err {
		if !strings.Contains(err.Error(), testCase.Result.Err.Error()) {
			t.Errorf("error results didtn match got:\n %v\n needed:\n %v", err, testCase.Result.Err)
		}
	}
}

func TestUnknownBadRequest(t *testing.T) {
	testCase := TestCase{
		Request: &SearchRequest{
			Limit:      10,
			Offset:     10,
			Query:      "",
			OrderField: "",
			OrderBy:    OrderByAsc,
		},
		Result: &ResultResp{
			Resp: nil,
			Err:  fmt.Errorf("unknown bad request error:"),
		},
	}
	s := &SearchClient{
		AccessToken: accesToken,
		URL:         ts.URL + "/unknownbadrequest",
	}
	_, err := s.FindUsers(*testCase.Request)

	if err != testCase.Result.Err {
		if !strings.Contains(err.Error(), testCase.Result.Err.Error()) {
			t.Errorf("error results didtn match got:\n %v\n needed:\n %v", err, testCase.Result.Err)
		}
	}
}

func TestUnknowError(t *testing.T) {
	testCase := TestCase{
		Request: &SearchRequest{
			Limit:      10,
			Offset:     10,
			Query:      "",
			OrderField: "",
			OrderBy:    OrderByAsc,
		},
		Result: &ResultResp{
			Resp: nil,
			Err:  fmt.Errorf("unknown error "),
		},
	}
	s := &SearchClient{
		AccessToken: accesToken,
		URL:         ts.URL + "/unknownerror",
	}
	_, err := s.FindUsers(*testCase.Request)
	if !strings.Contains(err.Error(), testCase.Result.Err.Error()) {
		t.Errorf("error results didtn match got:\n %v\n needed:\n %v", err.Error(), testCase.Result.Err.Error())
	}
}
