package item

//Inventory : structure of items
type Inventory struct {
	Barcode  string
	Itemname string
	Price    float64
}

type items []Inventory
	