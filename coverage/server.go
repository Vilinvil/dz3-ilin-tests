package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/antchfx/xmlquery"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

const (
	ErrorBadLimit   = `Limit invalid`
	ErrorBadOffset  = `Offset invalid`
	ErrorBadOrderBy = `OrderBy invalid`
)

var PatchDataSet = "dataset.xml"

// аналогична http.Error() только не добавляет /n в конце

func ErrorWrite(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	data := SearchErrorResponse{Error: error}
	res, err := json.Marshal(&data) // можно не покрывать
	if err != nil {
		log.Printf("couldn't json.Marshall %v. Error is %v", data, err)
	}

	_, err = w.Write(res) // можно не покрывать
	if err != nil {
		log.Printf("couldn't w.Write %v. Error is %v", res, err)
	}

}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var Users []User
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		ErrorWrite(w, ErrorBadLimit, 400)
		return
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		ErrorWrite(w, ErrorBadOffset, 400)
		return
	}

	orderBy, err := strconv.Atoi(r.URL.Query().Get("order_by"))
	if err != nil {
		ErrorWrite(w, ErrorBadOrderBy, 400)
		return
	}

	query := r.URL.Query().Get("query")
	orderField := r.URL.Query().Get("order_field")
	dataSet, err := ioutil.ReadFile(PatchDataSet) //можно не покрывать
	if err != nil {
		ErrorWrite(w, fmt.Sprintf("couldn't read file %s. Error is: %v", PatchDataSet, err), 500)
		return
	}

	doc, err := xmlquery.Parse(bytes.NewReader(dataSet))
	if err != nil {
		ErrorWrite(w, fmt.Sprintf("couldn't parse file %s. Error is: %v", PatchDataSet, err), 500)
		return
	}

	root := xmlquery.FindOne(doc, "//root")
	for _, n := range xmlquery.Find(root, "//row") {
		if limit+offset < 1 {
			break
		}

		// Сортировка найденных юзеров. В if-е поиск подстроки  query в нужных полях
		if strings.Contains(n.SelectElement("//first_name").InnerText(), query) ||
			strings.Contains(n.SelectElement("//last_name").InnerText(), query) ||
			strings.Contains(n.SelectElement("//about").InnerText(), query) {

			// Добавление найденного юзера в слайс юзеров
			id, err := strconv.Atoi(n.SelectElement("//id").InnerText())
			if err != nil {
				ErrorWrite(w, fmt.Sprintf("in %s incorrect id. Error is: %v", PatchDataSet, err), 500)
				return
			}

			age, err := strconv.Atoi(n.SelectElement("//age").InnerText())
			if err != nil {
				ErrorWrite(w, fmt.Sprintf("in %s incorrect age Error is: %v", PatchDataSet, err), 500)
				return
			}

			name := n.SelectElement("//first_name").InnerText() + " " + n.SelectElement("//last_name").InnerText()
			about := n.SelectElement("//about").InnerText()
			gender := n.SelectElement("//gender").InnerText()

			tmpUser := User{
				ID:     id,
				Name:   name,
				Age:    age,
				About:  about,
				Gender: gender,
			}
			Users = append(Users, tmpUser)
			limit--
		}
	}

	switch {
	case orderField == "Name" || orderField == "":
		{
			switch orderBy {
			case OrderByAsc:
				sort.Sort(orderNameAsc(Users))

			case OrderByDesc:
				sort.Sort(orderNameDesc(Users))

			case OrderByAsIs:

			default:
				ErrorWrite(w, ErrorBadOrderBy, 400)
				return
			}
		}
	case orderField == "Id":
		{
			switch orderBy {
			case OrderByAsc:
				sort.Sort(orderIdAsc(Users))

			case OrderByDesc:
				sort.Sort(orderIdDesc(Users))

			case OrderByAsIs:

			default:
				ErrorWrite(w, ErrorBadOrderBy, 400)
				return
			}
		}
	case orderField == "Age":
		{
			switch orderBy {
			case OrderByAsc:
				sort.Sort(orderAgeAsc(Users))

			case OrderByDesc:
				sort.Sort(orderAgeDesc(Users))

			case OrderByAsIs:

			default:
				{
					ErrorWrite(w, ErrorBadOrderBy, 400)
					return
				}
			}
		}
	default:
		ErrorWrite(w, ErrorBadOrderField, 400)
		return
	}

	result, err := json.Marshal(Users[offset:]) //можно не покрывать
	if err != nil {
		ErrorWrite(w, "couldn't marshal result to json", 500)
		return
	}

	_, err = w.Write(result) //можно не покрывать
	if err != nil {
		ErrorWrite(w, "couldn't to write result in http.ResponseWriter", 500)
		return
	}
}

func main() {

}
