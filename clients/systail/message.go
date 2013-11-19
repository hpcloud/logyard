package systail

import (
	"logyard/clients/common"
)

// Message is a struct corresponding to an entry in the
// systail log stream. Generally a subset of the fields in
// `AppLogMessage`.
type Message struct {
	Name string `json:"name"` // Component name (eg: dea)
	common.MessageCommon
}
