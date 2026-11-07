package sony

type MuteStatus string
type PowerStatus string

var (
	NotMuted MuteStatus = "off"
	Muted    MuteStatus = "on"

	Standby PowerStatus = "standby"
	Active  PowerStatus = "active"
)

type GetVolumeInformationParams struct {
	Output string `json:"output"`
}

type GetVolumeInformationResponse struct {
	MaxVolume int        `json:"maxVolume"`
	MinVolume int        `json:"minVolume"`
	Volume    int        `json:"volume"`
	Mute      MuteStatus `json:"mute"`
}

type GetInterfaceInformationResponse struct {
	ModelName string `json:"modelName"`
}

type GetPowerStatusResponse struct {
	Status PowerStatus `json:"status"`
}

type SetVolumeInformationParams struct {
	Volume string `json:"volume"`
	Output string `json:"output"`
}

type SetVolumeInformationResponse int
