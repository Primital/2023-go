package types

type LocationSolution struct {
	Location
	F9            int     `json:"f9"`
	F3            int     `json:"f3"`
	SalesCapacity float64 `json:"salesCapacity"`
	Revenue       float64 `json:"revenue"`
	Earnings      float64 `json:"earnings"`
	LeasingCost   float64 `json:"leasingCost"`
	Co2Saving     float64 `json:"co2Saving"`
}
