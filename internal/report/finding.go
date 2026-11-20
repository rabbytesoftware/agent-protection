package report

type Finding struct {
	Rule     string `json:"rule"`
	Location string `json:"location"`
	Snippet  string `json:"snippet"`
}
