// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package entity

type Relation string

const (
	Self            = "self"
	Icon            = "icon"
	Related         = "related"
	OrgRefudeAction = "org.refude.action"
	OrgRefudeDelete = "org.refude.delete"
	OrgRefudeMenu   = "org.refude.menu"
)
