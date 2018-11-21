package attributes

import (
	eh "github.com/looplab/eventhorizon"
)

const (
	AttributeUpdated                eh.EventType = "AttributeUpdated"
	AttributeUpdateRequest          eh.EventType = "AttributeUpdateRequest"
	AttributeGetCurrentValueRequest eh.EventType = "AttributeGetCurrentValueRequest"
)

func init() {
	eh.RegisterEventData(AttributeUpdated, func() eh.EventData { return &AttributeUpdatedData{} })
	eh.RegisterEventData(AttributeUpdateRequest, func() eh.EventData { return &AttributeUpdateRequestData{} })
	eh.RegisterEventData(AttributeGetCurrentValueRequest, func() eh.EventData { return &AttributeGetCurrentValueRequestData{} })
}

type PrivilegeData struct {
	License        int
	ReadPrivilege  int
	WritePrivilege int
	Readonly       bool
	IsSuppressed   bool
	Private        bool
}

type AttributeUpdatedData struct {
	Privileges PrivilegeData
	ReqID      eh.UUID
	FQDD       string
	Group      string
	Index      string
	Name       string
	Value      interface{}
	Error      string
}

type AttributeUpdateRequestData struct {
	ReqID eh.UUID
	FQDD  string
	Group string
	Index string
	Name  string
	Value interface{}
}

type AttributeGetCurrentValueRequestData struct {
	FQDD  string
	Group string
	Index string
	Name  string
}
