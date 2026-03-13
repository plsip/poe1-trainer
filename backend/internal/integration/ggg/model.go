package ggg

// Surowe typy odpowiedzi GGG API.
// Używane tylko wewnątrz pakietu ggg — nie wyciekają do pozostałej aplikacji.

// apiCharacter to pojedynczy wpis z GET /character
type apiCharacter struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Realm   string `json:"realm"`
	Class   string `json:"class"`
	League  string `json:"league"`
	Level   int    `json:"level"`
	Expired bool   `json:"expired"`
}

// apiCharacterDetailResponse to odpowiedź GET /character/{id}
// Zawiera ekwipunek i osadzone gemy (skills via socketedItems).
type apiCharacterDetailResponse struct {
	Character apiCharacter `json:"character"`
	Items     []apiItem    `json:"items"`
}

// apiItem reprezentuje pojedynczy slot ekwipunku.
type apiItem struct {
	ID            string    `json:"id"`
	TypeLine      string    `json:"typeLine"`
	BaseType      string    `json:"baseType"`
	InventoryID   string    `json:"inventoryId"` // np. "BodyArmour", "Weapon", "Helm"
	SocketedItems []apiItem `json:"socketedItems,omitempty"`
}
