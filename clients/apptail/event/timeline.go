package event

type App struct {
	GUID  string `json:"guid"`
	Space string `json:"space_guid"`
	Name  string `json:"name"`
}

type TimelineEvent struct {
	App           App `json:"app"`
	InstanceIndex int `json:"instance_index"`
}
