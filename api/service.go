package api

import (
	"context"
)

// FFService describes the fabrika-fotoknigi.ru service.
type FFService interface {
	GetGroups(ctx context.Context, status int, fromTS int64) ([]Group, error)
}

/*
 {
    {
        "id": 348534,
        "client_id": 12949,
        "client": {
            "id": 12949,
            "zoho_id": "2060216000006760351"
        },
        "status": {
            "value": 40,
            "name": "MADE",
            "title": "Произведен, ждет упаковки"
        },
        "tstamp": 1581253147,
        "datetime": "09.02.2020 15:59",
        "execution_tstamp": 1581315239,
        "execution_text": "10.02.2020",
        "adoption_tstamp": 1581142439,
        "adoption_text": "08.02.2020 09:13",
        "orders": {
            "860724": 1009689,
            "865400": 1014365,
            "865402": 1014367
        },
        "quantity": 3,
        "price": 0,
        "delivery_price": 0,
        "payment_price": 0,
        "total": 0,
        "isArchived": false,
        "currency": "RUB",
        "weight": 539,
        "payment": {
            "id": 7,
            "alias": "account",
            "title": "Со счета"
        },
        "delivery": {
            "id": 55,
            "alias": "pickup-np",
            "title": "Самовывоз НП"
        },
        "address": {
            "id": "352572",
            "type": "fotokniga",
            "lastname": "Тест",
            "firstname": "для Ирины",
            "middlename": "",
            "phone": "",
            "email": "",
            "passport": "",
            "passport_date": "",
            "postal": "",
            "region": "Москва",
            "district": "",
            "city": "Москва",
            "street": "",
            "home": "",
            "flat": "",
            "comment": null,
            "address": "",
            "metro": "",
            "delivery_time": "",
            "slCity": "0",
            "slPickup": "0",
            "atDeliveryType": "0",
            "dlDeliveryType": "0",
            "person": "",
            "city_id": "182905",
            "dellinCity": "0",
            "dellinTerminal": "0",
            "dellinAddressId": "0",
            "dellinPhoneId": "0",
            "dellinPersonId": "0",
            "template_id": "259141",
            "dellinCounterAgent": "0",
            "slSrokDostavki": "",
            "npCity": "0",
            "npPickup": "0",
            "text": "Тест для Ирины\nМосква,"
        },
        "comment": null,
        "hash": "f176c3c5892985b90105469ae01ddee31719c688",
        "hash_source": "id:348534;client_id:12949;status:MADE;quantity:3;delivery_price:0.00;currency:RUB;payment:account;delivery:pickup-np",
        "boxes": [
            {
                "id": 287672,
                "number": 1,
                "barcode": "",
                "deliveryPrice": 0,
                "baseDeliveryPrice": 0,
                "comment": "",
                "orderId": "9099000992124224493",
                "orderNumber": "20191408-FFM37-608-242157",
                "added_tstamp": 1581143116,
                "status": 111,
                "status_title": "#Отгружен вместе с партией",
                "status_tstamp": 1581264297
            }
        ],
        "checkouStep": null,
        "basePrice": 528,
        "baseDeliveryPrice": 0,
        "credit": 311503,
        "messages": "",
        "npfactory": true
	}
*/

//Group dto (orders group)
type Group struct {
	ID        int   `json:"Id"`
	CreatedTS int64 `json:"tstamp"`
	Boxes     []Box `json:"boxes"`
}

//Box dto (group post box)
type Box struct {
	BoxNumber   int    `json:"number"`
	OrderNumber string `json:"orderNumber"`
}
