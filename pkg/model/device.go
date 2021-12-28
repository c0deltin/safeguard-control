package model

type Device struct {
	_ struct{} `json:"-"`

	ID      string `json:"id"`
	IsArmed bool   `json:"isArmed"`
}

func (d Device) MarshalToDeviceDB() *DeviceDB {
	return &DeviceDB{
		ID:      d.ID,
		IsArmed: d.IsArmed,
	}
}

type DeviceDB struct {
	_ struct{} `dynamodbav:"-"`

	ID      string `dynamodbav:"id"`
	IsArmed bool   `dynamodbav:"isArmed"`
}

func (d DeviceDB) MarshalToDeviceDB() *Device {
	return &Device{
		ID:      d.ID,
		IsArmed: d.IsArmed,
	}
}
