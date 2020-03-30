package main

import (
	"bufio"
	json "encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

type usersStruct struct {
	Browsers []string `json:"browsers"`
	// Company  string   `json:"company"`
	// Country  string   `json:"country"`
	Email string `json:"email"`
	// Job      string   `json:"job"`
	Name string `json:"name"`
	// Phone    string   `json:"phone"`
}

//FastSearch ...
func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var userPool *sync.Pool
	userPool = &sync.Pool{
		New: func() interface{} {
			return new(usersStruct)
		},
	}

	seenBrowsers2 := make(map[string]bool, 115)
	foundUsers := ""
	i := -1
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		i++
		fileLine := scanner.Text()
		if !strings.Contains(fileLine, "Android") && !strings.Contains(fileLine, "MSIE") {
			continue
		}

		user := userPool.Get().(*usersStruct)
		userPool.Put(user)
		err1 := user.UnmarshalJSON(scanner.Bytes())
		// err1 := user.UnmarshalJSON(scanner.Bytes())
		if err != nil {
			panic(err1)
		}

		isAndroid := false
		isMSIE := false

		for _, browser := range user.Browsers {
			if strings.Contains(browser, "Android") {
				isAndroid = true
				seenBrowsers2[browser] = true
				if isMSIE {
					break
				}
			}
			if strings.Contains(browser, "MSIE") {
				isMSIE = true
				seenBrowsers2[browser] = true
				if isAndroid {
					break
				}
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		// log.Println("Android and MSIE user:", user["name"], user["email"])
		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user.Name, strings.Replace(user.Email, "@", " [at] ", 1))
	}

	fmt.Fprintln(out, "found users:\n"+foundUsers)
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers2))
}

//-------------easyJSON--------------
func easyjsonD02638feDecodeCourseraHw3Bench(in *jlexer.Lexer, out *usersStruct) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "browsers":
			if in.IsNull() {
				in.Skip()
				out.Browsers = nil
			} else {
				in.Delim('[')
				if out.Browsers == nil {
					if !in.IsDelim(']') {
						out.Browsers = make([]string, 0, 4)
					} else {
						out.Browsers = []string{}
					}
				} else {
					out.Browsers = (out.Browsers)[:0]
				}
				for !in.IsDelim(']') {
					var v1 string
					v1 = string(in.String())
					out.Browsers = append(out.Browsers, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		// case "company":
		// 	out.Company = string(in.String())
		// case "country":
		// 	out.Country = string(in.String())
		case "email":
			out.Email = string(in.String())
		// case "job":
		// 	out.Job = string(in.String())
		case "name":
			out.Name = string(in.String())
		// case "phone":
		// 	out.Phone = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *usersStruct) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonD02638feDecodeCourseraHw3Bench(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *usersStruct) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonD02638feDecodeCourseraHw3Bench(l, v)
}
