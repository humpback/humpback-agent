package model

type GetNetworkRequest struct {
	NetworkId string `json:"networkId"`
}

type CreateNetworkRequest struct {
	NetworkName string `json:"networkName"`
	Driver      string `json:"driver"`
	Scope       string `json:"scope"`
}
