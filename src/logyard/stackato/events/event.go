package events

type Event struct {
	Type     string // what type of event?
	Desc     string // description of this event to be shown as-is to humans
	Severity string
	Info     map[string]interface{} // event-specific information as json
	Process  string                 // which process generated this event?
	UnixTime int64
	NodeID   string // from which node did this event appear?
}
