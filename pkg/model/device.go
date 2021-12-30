package model

type Device struct {
	_ struct{} `json:"-"`

	ID          string `json:"id"`
	IsArmed     bool   `json:"isArmed"`
	Description string `json:"description"`
}

func (d Device) MarshalToDB() *DeviceDB {
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

func (d DeviceDB) MarshalToRequest() *Device {
	return &Device{
		ID:      d.ID,
		IsArmed: d.IsArmed,
	}
}
