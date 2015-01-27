package model

import (
	"encoding/json"
	"github.com/petemoore/taskcluster-client-go/utils"
	"io"
	"io/ioutil"
	"net/http"
)

var (
	err error
)

func loadJson(reader io.Reader, schema *string) APIModel {
	var bytes []byte
	bytes, err = ioutil.ReadAll(reader)
	utils.ExitOnFail(err)
	var m APIModel
	switch *schema {
	case "http://schemas.taskcluster.net/base/v1/api-reference.json":
		m = new(API)
	case "http://schemas.taskcluster.net/base/v1/exchanges-reference.json":
		m = new(Exchange)
	}
	err = json.Unmarshal(bytes, m)
	m.postPopulate()
	utils.ExitOnFail(err)
	return m
}

func LoadJsonSchema(url string) *JsonSchemaTopLevel {
	var resp *http.Response
	resp, err = http.Get(url)
	utils.ExitOnFail(err)
	defer resp.Body.Close()
	var bytes []byte
	bytes, err = ioutil.ReadAll(resp.Body)
	utils.ExitOnFail(err)
	m := new(JsonSchemaTopLevel)
	err = json.Unmarshal(bytes, m)
	utils.ExitOnFail(err)
	return m
}

func LoadAPIs(bytes []byte) ([]APIDefinition, map[string]*JsonSchemaTopLevel) {
	err = json.Unmarshal(bytes, &apis)
	utils.ExitOnFail(err)
	for i := range apis {
		var resp *http.Response
		resp, err = http.Get(apis[i].URL)
		utils.ExitOnFail(err)
		defer resp.Body.Close()
		data := loadJson(resp.Body, &apis[i].SchemaURL)
		apis[i].Data = data
	}
	return apis, schemas
}