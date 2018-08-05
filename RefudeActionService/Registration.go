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
	"log"
)

const RegistrationMediaType = "application/vnd.org.refude.registration"


type ActionDesc struct {
	Name     string
	Comment  string
	IconName string
	Exec     string
}

type RegistrationData struct {
	id    int
	Owner string
	Descs []ActionDesc
}

type Registration struct {
	lib.AbstractResource
	RegistrationData
	expiration time.Time
}

func MakeNewRegistration() *Registration {
	var registration Registration
	registration.id = getId();
	registration.Self = lib.Standardizef("/registration/%d", registration.id)
	registration.Mt = RegistrationMediaType
	registration.Descs = []ActionDesc{}
	registration.expiration = time.Now().Add(time.Minute)
	return &registration
}


func (reg *Registration) DELETE(w http.ResponseWriter, r *http.Request) {
	jm.RemoveAll(reg.Self)
}

func (reg *Registration) PATCH(w http.ResponseWriter, r *http.Request) {
	var regData = reg.RegistrationData
	if data, err := ioutil.ReadAll(r.Body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		log.Println("data:", string(data))
		if err = json.Unmarshal(data, &regData); err != nil {
			lib.ReportUnprocessableEntity(w, err)
		} else if regData.Descs == nil {
			lib.ReportUnprocessableEntity(w, errors.New("'Descs may not be null, use empty array instead'"))
		} else {
			var patchedReg = *reg
			patchedReg.expiration = time.Now().Add(time.Minute)
			patchedReg.RegistrationData = regData
			var newResources = []lib.Resource{&patchedReg}
			var actionPrefix= lib.Standardizef("/actions/%d", patchedReg.id)
			if  regData.Owner != reg.Owner ||
				len(regData.Descs) != len(regData.Descs) ||
				(len(regData.Descs) > 0 && &regData.Descs[0] != &reg.Descs[0]) {
				patchedReg.Relates = make(map[lib.MediaType][]lib.StandardizedPath)
				for i, actionDesc := range regData.Descs {
					var actionPath= lib.Standardizef(fmt.Sprintf("%s/%d", actionPrefix, i))
					var argv= strings.Fields(actionDesc.Exec)
					var executer = func() {
						lib.RunCmd(argv)
					}
					var act= lib.MakeAction(actionPath, actionDesc.Name, actionDesc.Comment, actionDesc.IconName, executer)
					newResources = append(newResources, act)
					lib.Relate(&patchedReg.AbstractResource, &act.AbstractResource)
				}
			}
			fmt.Println("PatchedRegistration:", patchedReg)
			jm.RemoveAndMap([]lib.StandardizedPath{actionPrefix}, newResources)

			w.WriteHeader(http.StatusAccepted)
		}
	}
}

