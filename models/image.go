package models

// Image - define image info struct
type Image struct {
	Image string `json:'Image'`
}

// ImageTags - define image tags info struct
type ImageTags struct {
	ImageName string   `json:"ImageName"`
	Tags      []string `json:"Tags"`
}
