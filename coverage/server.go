package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/antchfx/xmlquery"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

var (
	ErrorNotFound      = clientError("Users by request not found")
	ErrorBadOrderField = clientError("OrderField invalid")
	ErrorAccessToken   = clientError("Wrong AccessToken")
	ErrorServer        = clientError("Internal server error")
	ErrorBadRequest    = clientError("request invalid")
	ErrorBadOrderBy    = clientError("OrderBy invalid")

	autenficationtokens = map[string]struct{}{"2a54a886a8bbcc309ae4ffa75241cd6d": {}}
	PatchDataSet        = "data_set.xml"
)

type clientError string

func (c clientError) Error() string {
	return string(c)
}

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

func isTokenValid(r *http.Request) (bool, error) {
	if r == nil {
		return false, fmt.Errorf("*http.Request == nil in isTokenValid. %w", ErrorServer)
	}

	token := r.Header.Get("AccessToken")
	_, ok := autenficationtokens[token]
	return ok, nil
}

func getStringParam(r *http.Request, param string) (string, error) {
	if r == nil {
		return "", fmt.Errorf("*http.Request == nil in getStringParam. %w", ErrorServer)
	}
	return r.URL.Query().Get(param), nil
}

func getIntParam(r *http.Request, param string) (int, error) {
	if r == nil {
		return 0, fmt.Errorf("*http.Request == nil in getIntParam . %w", ErrorServer)
	}

	return strconv.Atoi(r.URL.Query().Get(param))
}

func parseParamFromUrl(sr *SearchRequest, r *http.Request) error {
	if sr == nil {
		return fmt.Errorf("*SearchRequest == nil in parseParamFromUrl. %w", ErrorServer)
	}
	if r == nil {
		return fmt.Errorf("*http.Request == nil in parseParamFromUrl. %w", ErrorServer)
	}

	var err error
	sr.Limit, err = getIntParam(r, "limit")
	if err != nil {
		return fmt.Errorf("error is: %w in parseParamFromUrl", err)
	}
	if sr.Limit < 1 {
		return fmt.Errorf("limit must be > 0")
	}

	sr.Offset, err = getIntParam(r, "offset")
	if err != nil {
		return fmt.Errorf("error is: %w in parseParamFromUrl", err)
	}
	if sr.Offset < 0 {
		return fmt.Errorf("offset must be >= 0")
	}

	sr.OrderBy, err = getIntParam(r, "order_by")
	if err != nil {
		return fmt.Errorf("error is: %w in parseParamFromUrl", err)
	}

	sr.Query, err = getStringParam(r, "query")
	if err != nil {
		return fmt.Errorf("error is: %w in parseParamFromUrl", err)
	}
	sr.OrderField, err = getStringParam(r, "order_field")
	if err != nil {
		return fmt.Errorf("error is: %w in parseParamFromUrl", err)
	}

	return nil
}

func getNodeElementText(node *xmlquery.Node, elem string) (string, error) {
	if node == nil {
		return "", fmt.Errorf("*xmlquery.Node == nil in getNodeElementText. %w", ErrorServer)
	}
	return node.SelectElement(elem).InnerText(), nil
}

func getNodeElementInt(node *xmlquery.Node, elem string) (int, error) {
	if node == nil {
		return 0, fmt.Errorf("*xmlquery.Node == nil in getNodeElementInt. %w", ErrorServer)
	}
	return strconv.Atoi(node.SelectElement(elem).InnerText())
}

func isQueryInXmlNode(node *xmlquery.Node, query string) (bool, error) {
	if node == nil {
		return false, fmt.Errorf("*xmlquery.Node == nil in isQueryInXmlNode. %w", ErrorServer)
	}

	fName, err := getNodeElementText(node, "//first_name")
	if err != nil {
		return false, fmt.Errorf("error is %w in isQueryInXmlNode", err)
	}

	lName, err := getNodeElementText(node, "//last_name")
	if err != nil {
		return false, fmt.Errorf("error is %w in isQueryInXmlNode", err)
	}

	about, err := getNodeElementText(node, "//about")
	if err != nil {
		return false, fmt.Errorf("error is %w in isQueryInXmlNode", err)
	}
	return strings.Contains(fName, query) ||
		strings.Contains(lName, query) ||
		strings.Contains(about, query), nil
}

func createReaderFromFile(patchFile string) (*bytes.Reader, error) {
	data, err := ioutil.ReadFile(patchFile)
	if err != nil {
		return nil, fmt.Errorf("couldn't read file %s in createReaderFromFile. Error is: %+v. %w", PatchDataSet, err, ErrorServer)
	}
	return bytes.NewReader(data), nil
}

func parseUsersFromXml(limit, offset int, query string) (*[]User, error) {
	dataSet, err := createReaderFromFile(PatchDataSet)
	if err != nil {
		return nil, fmt.Errorf("error is: %w in parseUsersFromXml", err)
	}

	doc, err := xmlquery.Parse(dataSet)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse file %s. Error is: %+v. %w", PatchDataSet, err, ErrorServer)
	}

	root := xmlquery.FindOne(doc, "//root")
	if root == nil {
		return nil, fmt.Errorf("//root(node) == nil in parseUsersFromXml. %w", ErrorServer)
	}

	Users := make([]User, 0)
	for _, node := range xmlquery.Find(root, "//row") {
		if limit+offset < 1 {
			break
		}

		bollQueryInXmlNode, err := isQueryInXmlNode(node, query)
		if err != nil {
			return nil, fmt.Errorf("%v == nil in parseUsersFromXml", node)
		}

		if bollQueryInXmlNode {
			id, err := getNodeElementInt(node, "id")
			if err != nil {
				return nil, fmt.Errorf("incorrect id in %s. Error is: %+v", PatchDataSet, err)
			}

			age, err := getNodeElementInt(node, "age")
			if err != nil {
				return nil, fmt.Errorf("incorrect age in %s. Error is: %+v", PatchDataSet, err)
			}

			fName, err := getNodeElementText(node, "//first_name")
			if err != nil {
				return nil, fmt.Errorf("error is %w in parseUsersFromXml", err)
			}

			lName, err := getNodeElementText(node, "//last_name")
			if err != nil {
				return nil, fmt.Errorf("error is %w in parseUsersFromXml", err)
			}

			about, err := getNodeElementText(node, "//about")
			if err != nil {
				return nil, fmt.Errorf("error is %w in parseUsersFromXml", err)
			}

			gender, err := getNodeElementText(node, "//gender")
			if err != nil {
				return nil, fmt.Errorf("error is %w in parseUsersFromXml", err)
			}

			Users = append(Users, User{
				ID:     id,
				Name:   fName + " " + lName,
				Age:    age,
				About:  about,
				Gender: gender,
			})
			limit--
		}
	}

	return &Users, nil
}

func sortSlUser(orderField string, orderBy int, Users *[]User) error {
	switch orderField {
	case "Name", "":
		{
			switch orderBy {
			case OrderByAsc:
				sort.Sort(orderNameAsc(*Users))

			case OrderByDesc:
				sort.Sort(orderNameDesc(*Users))

			case OrderByAsIs:

			default:
				return fmt.Errorf("%w in sortSlUsers", ErrorBadOrderBy)
			}
		}
	case "Id":
		{
			switch orderBy {
			case OrderByAsc:
				sort.Sort(orderIdAsc(*Users))

			case OrderByDesc:
				sort.Sort(orderIdDesc(*Users))

			case OrderByAsIs:

			default:
				return fmt.Errorf("%w in sortSlUsers", ErrorBadOrderBy)
			}
		}
	case "Age":
		{
			switch orderBy {
			case OrderByAsc:
				sort.Sort(orderAgeAsc(*Users))

			case OrderByDesc:
				sort.Sort(orderAgeDesc(*Users))

			case OrderByAsIs:

			default:
				return fmt.Errorf("%w in sortSlUsers", ErrorBadOrderBy)
			}
		}
	default:
		return fmt.Errorf("%w in sortSlUsers", ErrorBadOrderField)
	}

	return nil
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	boolTokenValid, err := isTokenValid(r)
	if err != nil {
		log.Printf("Error is: %+v in SearchServer", err)
		errorWrite(w, ErrorServer.Error(), http.StatusInternalServerError)
		return
	}
	if !boolTokenValid {
		log.Printf("User is not authorized. Token is %+v.", r.Header.Get("AccessToken"))
		errorWrite(w, ErrorAccessToken.Error(), http.StatusUnauthorized)
		return
	}

	sReq := &SearchRequest{}
	err = parseParamFromUrl(sReq, r)
	if err != nil {
		log.Printf("Error is: %+v in SearchServer", err)
		if errors.Is(err, ErrorServer) {
			errorWrite(w, ErrorServer.Error(), http.StatusInternalServerError)
		} else {
			errorWrite(w, ErrorBadRequest.Error(), http.StatusBadRequest)
		}
		return
	}

	var Users *[]User
	Users, err = parseUsersFromXml(sReq.Limit, sReq.Offset, sReq.Query)
	if err != nil || Users == nil {
		log.Printf("Error is: %+v in SearchServer. Users are: %+v", err, Users)
		errorWrite(w, ErrorServer.Error(), http.StatusInternalServerError)
		return
	}

	err = sortSlUser(sReq.OrderField, sReq.OrderBy, Users)
	if err != nil {
		log.Printf("Error is: %+v in SearchServer.", err)
		errorWrite(w, ErrorBadRequest.Error(), http.StatusBadRequest)
		return
	}

	if len(*Users) <= sReq.Offset {
		log.Printf("Operation (*Users)[sReq.Offset:] not posible because len(*Users) <= sReq.Offset")
		errorWrite(w, ErrorNotFound.Error(), 404)
		return
	}
	result, err := json.Marshal((*Users)[sReq.Offset:])
	if err != nil {
		log.Printf("Couldn't marshal result to json in SearchServer. Error is: %+v", err)
		errorWrite(w, ErrorServer.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(result)
	if err != nil {
		log.Printf("Couldn't to write result in http.ResponseWriter in SearchServer. Error is: %+v", err)
		errorWrite(w, ErrorServer.Error(), http.StatusInternalServerError)
		return
	}
}
