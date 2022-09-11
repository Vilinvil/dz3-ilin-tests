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
	ErrorBadLimit    = `Limit invalid`
	ErrorBadOffset   = `Offset invalid`
	ErrorBadOrderBy  = `OrderBy invalid`
	ErrorAccessToken = "wrong AccessToken"
)

var (
	autenficationtokens = map[string]struct{}{"2a54a886a8bbcc309ae4ffa75241cd6d": {}}
	PatchDataSet        = "data_set.xml"
)

func errorWrite(w http.ResponseWriter, errStr string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	data := SearchErrorResponse{Error: errStr}
	res, err := json.Marshal(&data)
	if err != nil {
		log.Printf("couldn't json.Marshall %v. Error is %v", data, err)
	}

	_, err = w.Write(res)
	if err != nil {
		log.Printf("couldn't w.Write %v. Error is %v", res, err)
	}
}

func isTokenValid(r *http.Request) bool {
	token := r.Header.Get("AccessToken")
	_, ok := autenficationtokens[token]
	return ok
}

func getStringParam(r *http.Request, param string) string {
	return r.URL.Query().Get(param)
}

func getIntParam(r *http.Request, param string) (int, error) {
	res, err := strconv.Atoi(r.URL.Query().Get(param))
	return res, err
}

func parseParamFromUrl(sr *SearchRequest, r *http.Request) error {
	var err error
	sr.Limit, err = getIntParam(r, "limit")
	if err != nil {
		return fmt.Errorf("%v. Error is: %v", ErrorBadLimit, err)
	}

	sr.Offset, err = getIntParam(r, "offset")
	if err != nil {
		return fmt.Errorf("%v. Error is: %v", ErrorBadOffset, err)

	}

	sr.OrderBy, err = getIntParam(r, "order_by")
	if err != nil {
		return fmt.Errorf("%v. Error is: %v", ErrorBadOrderBy, err)
	}

	sr.Query = getStringParam(r, "query")
	sr.OrderField = getStringParam(r, "order_field")
	return nil
}

func isQueryInXmlNode(n *xmlquery.Node, query string) bool {
	return strings.Contains(n.SelectElement("//first_name").InnerText(), query) ||
		strings.Contains(n.SelectElement("//last_name").InnerText(), query) ||
		strings.Contains(n.SelectElement("//about").InnerText(), query)
}

func parseUsersFromXml(sReq *SearchRequest, Users *[]User) error {
	dataSet, err := ioutil.ReadFile(PatchDataSet)
	if err != nil {
		return fmt.Errorf("couldn't read file %s. Error is: %v", PatchDataSet, err)
	}

	doc, err := xmlquery.Parse(bytes.NewReader(dataSet))
	if err != nil {
		return fmt.Errorf("couldn't parse file %s. Error is: %v", PatchDataSet, err)
	}

	root := xmlquery.FindOne(doc, "//root")
	for _, n := range xmlquery.Find(root, "//row") {
		if sReq.Limit+sReq.Offset < 1 {
			break
		}
		//Поиск нужных юзеров в xml, и добавление их в слайс
		if isQueryInXmlNode(n, sReq.Query) {
			id, err := strconv.Atoi(n.SelectElement("//id").InnerText())
			if err != nil {
				return fmt.Errorf("in %s incorrect id. Error is: %v", PatchDataSet, err)
			}

			age, err := strconv.Atoi(n.SelectElement("//age").InnerText())
			if err != nil {
				return fmt.Errorf("in %s incorrect age Error is: %v", PatchDataSet, err)
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
			*Users = append(*Users, tmpUser)
			sReq.Limit--
		}
	}
	return nil
}

func sortSlUser(sReq *SearchRequest, Users *[]User) error {
	switch sReq.OrderField {
	case "Name", "":
		{
			switch sReq.OrderBy {
			case OrderByAsc:
				sort.Sort(orderNameAsc(*Users))

			case OrderByDesc:
				sort.Sort(orderNameDesc(*Users))

			case OrderByAsIs:

			default:
				return fmt.Errorf(ErrorBadOrderBy)
			}
		}
	case "Id":
		{
			switch sReq.OrderBy {
			case OrderByAsc:
				sort.Sort(orderIdAsc(*Users))

			case OrderByDesc:
				sort.Sort(orderIdDesc(*Users))

			case OrderByAsIs:

			default:
				return fmt.Errorf(ErrorBadOrderBy)
			}
		}
	case "Age":
		{
			switch sReq.OrderBy {
			case OrderByAsc:
				sort.Sort(orderAgeAsc(*Users))

			case OrderByDesc:
				sort.Sort(orderAgeDesc(*Users))

			case OrderByAsIs:

			default:
				{
					return fmt.Errorf(ErrorBadOrderBy)
				}
			}
		}
	default:
		return fmt.Errorf(ErrorBadOrderField)
	}
	return nil
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if !isTokenValid(r) {
		errorWrite(w, ErrorAccessToken, http.StatusUnauthorized)
		return
	}

	sReq := &SearchRequest{}
	err := parseParamFromUrl(sReq, r)
	if err != nil {
		errorWrite(w, err.Error(), 400)
		return
	}

	var Users []User
	err = parseUsersFromXml(sReq, &Users)
	if err != nil {
		errorWrite(w, err.Error(), 500)
		return
	}

	err = sortSlUser(sReq, &Users)
	if err != nil {
		errorWrite(w, err.Error(), 400)
		return
	}

	result, err := json.Marshal(Users[sReq.Offset:])
	if err != nil {
		errorWrite(w, "couldn't marshal result to json", 500)
		return
	}

	_, err = w.Write(result)
	if err != nil {
		errorWrite(w, "couldn't to write result in http.ResponseWriter", 500)
		return
	}
}
