package model

type Device struct {
	_ struct{} `json:"-"`

	ID          string `json:"id"`
	IsArmed     bool   `json:"isArmed"`
	Description string `json:"description"`
}

func (d Device) ConvertToDB() *DeviceDB {
	return &DeviceDB{
		ID:      d.ID,
		IsArmed: d.IsArmed,
	}
}

type DeviceResponse struct {
	Device *Device `json:"device"`
}

type DeviceDB struct {
	_ struct{} `dynamodbav:"-"`

	ID          string `dynamodbav:"id"`
	IsArmed     bool   `dynamodbav:"isArmed"`
	Description string `json:"description"`
}

func (d DeviceDB) ConvertToRequest() *Device {
	return &Device{
		ID:      d.ID,
		IsArmed: d.IsArmed,
	}
}

type DevicesResponse struct {
	Devices []*Device `json:"devices"`
}

func ConvertSliceToRequest(devices []DeviceDB) []*Device {
	var resp []*Device
	for _, d := range devices {
		resp = append(resp, d.ConvertToRequest())
	}

	return resp
}
