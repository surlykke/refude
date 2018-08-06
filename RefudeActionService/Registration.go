package main

import (
	"github.com/surlykke/RefudeServices/lib"
	"time"
	"fmt"
	"strings"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"errors"
)

const RegistrationMediaType = "application/vnd.org.refude.registration+json"

type ActionDesc struct {
	Name     string
	Comment  string
	IconName string
	Exec     string
}

type RegistrationData struct {
	Owner string
	Descs []ActionDesc
}

type Registration struct {
	lib.AbstractResource
	RegistrationData
	id      int
	Expires time.Time
}

func MakeNewRegistration() *Registration {
	var registration Registration
	registration.id = getId();
	registration.Self = lib.Standardizef("/registration/%d", registration.id)
	registration.Mt = RegistrationMediaType
	registration.Descs = []ActionDesc{}
	registration.Expires = time.Now().Add(time.Minute)
	return &registration
}

func (reg *Registration) DELETE(w http.ResponseWriter, r *http.Request) {
	jm.RemoveAll(reg.Self)
}

func (reg *Registration) PATCH(w http.ResponseWriter, r *http.Request) {
	if bytes, err := ioutil.ReadAll(r.Body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		var patchedRegistration = *reg
		if err = json.Unmarshal(bytes, &patchedRegistration.RegistrationData); err != nil {
			lib.ReportUnprocessableEntity(w, err)
		} else if patchedRegistration.Descs == nil {
			lib.ReportUnprocessableEntity(w, errors.New("'Descs may not be null, use empty array instead'"))
		} else {
			patchedRegistration.Expires = time.Now().Add(time.Minute)
			var newResources = []lib.Resource{&patchedRegistration}
			var prefixesToRemove []lib.StandardizedPath
			if reg.dataDifferent(&patchedRegistration.RegistrationData) {
				patchedRegistration.Relates = make(map[lib.MediaType][]lib.StandardizedPath)
				var actionPrefix = lib.Standardizef("/actions/%d", patchedRegistration.id)
				prefixesToRemove = []lib.StandardizedPath{actionPrefix}
				for i, actionDesc := range patchedRegistration.Descs {
					var actionPath = lib.Standardizef(fmt.Sprintf("%s/%d", actionPrefix, i))
					var argv = strings.Fields(actionDesc.Exec)
					var executer = func() {
						lib.RunCmd(argv)
					}
					var act = lib.MakeAction(actionPath, actionDesc.Name, actionDesc.Comment, actionDesc.IconName, executer)
					newResources = append(newResources, act)
					lib.Relate(&patchedRegistration.AbstractResource, &act.AbstractResource)
				}
			}

			jm.RemoveAndMap(prefixesToRemove, newResources)
			w.WriteHeader(http.StatusAccepted)
		}
	}
}

func (reg *Registration) dataDifferent(other *RegistrationData) bool {
	return !(other.Owner == reg.Owner &&
		len(other.Descs) == len(reg.Descs) &&
		(len(other.Descs) == 0 || &other.Descs[0] != &reg.Descs[0]))
}
