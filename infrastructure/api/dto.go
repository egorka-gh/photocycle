package api

//GroupBoxes top dto object 4 boxes
type GroupBoxes struct {
	ID int `json:"orderGroupId"`
	/*
	   "price": 62.77,
	   "delivery_price": 3,
	   "payment_price": 0,
	   "total": 65.77,
	*/
	Boxes []Box `json:"boxes"`
}

//Box dto
type Box struct {
	ID      int       `json:"boxId"`
	Number  int       `json:"boxNumber"`
	Barcode string    `json:"barcode"`
	Price   float64   `json:"boxTotalPrice"`
	Weight  int       `json:"weight"`
	Items   []BoxItem `json:"orders"`
}

//BoxItem dto
type BoxItem struct {
	OrderID int    `json:"orderId"`
	Alias   string `json:"alias"`
	Type    string `json:"type"`
	From    int    `json:"order_items_from"`
	To      int    `json:"order_items_to"`
}

//NPGroup netprint group dto (orders group)
type NPGroup struct {
	ID        int     `json:"id"`
	Status    Status  `json:"status"`
	CreatedTS int64   `json:"tstamp"`
	Boxes     []NPBox `json:"boxes"`
	Npfactory bool    `json:"npfactory"`
}

//NPBox dto (netprint maip box)
type NPBox struct {
	BoxNumber   int    `json:"number"`
	OrderNumber string `json:"orderNumber"`
}

//Status dto (group post box)
type Status struct {
	Value int `json:"value"`
}
