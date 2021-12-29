package model

const DeviceIDTag = "device_id"

type Capture struct {
	_ struct{} `json:"-"`

	DeviceID    string `json:"deviceID"`
	CaptureDate int64  `json:"captureDate"`
	S3ObjectKey string `json:"s3ObjectKey"`
}

func (c Capture) MarshalToCaptureDB() *CaptureDB {
	return &CaptureDB{
		DeviceID:    c.DeviceID,
		CaptureDate: c.CaptureDate,
		S3ObjectKey: c.S3ObjectKey,
	}
}

type CaptureResponse struct {
	Capture *Capture `json:"captures"`
}

type CapturesResponse struct {
	Captures []Capture `json:"captures"`
}

type CaptureDB struct {
	_ struct{} `dynamodbav:"-"`

	DeviceID    string `dynamodbav:"deviceID"`
	CaptureDate int64  `dynamodbav:"captureDate"`
	S3ObjectKey string `dynamodbav:"s3ObjectKey"`
}

func (c CaptureDB) MarshalToCapture() *Capture {
	return &Capture{
		DeviceID:    c.DeviceID,
		CaptureDate: c.CaptureDate,
		S3ObjectKey: c.S3ObjectKey,
	}
}
