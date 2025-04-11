package model

type GetImageRequest struct {
	ImageId string `json:"imageId"`
}

type QueryImageRequest struct{}
type PushImageRequest struct{}
type PullImageRequest struct {
	Image         string `json:"image"`
	All           bool   `json:"all"`
	Platform      string `json:"platform"`
	ServerAddress string
	UserName      string
	Password      string
}
type DeleteImageRequest struct{}
