package financialinstruments

type financialInstrument struct {
	UUID                   string                 `json:"uuid"`
	PrefLabel              string                 `json:"prefLabel"`
	AlternativeIdentifiers alternativeIdentifiers `json:"alternativeIdentifiers"`
	IssuedBy               string 		      `json:"issuedBy,omitempty"`
}

type alternativeIdentifiers struct {
	UUIDS             []string `json:"uuids"`
	FactsetIdentifier string   `json:"factsetIdentifier"`
	FIGICode          string   `json:"figiCode"`
}

const (
	uppIdentifierLabel = "UPPIdentifier"
	factsetIdentifierLabel = "FactsetIdentifier"
	figiIdentifierLabel = "FIGIIdentifier"
)