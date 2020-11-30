package photocycle

import (
	"context"
	"encoding/json"
	"time"
)

// Repository describes the persistence on model
type Repository interface {
	//4 netprint boxes
	GetLastNetprintSync(ctx context.Context, source int) (int64, error)
	SetLastNetprintSync(ctx context.Context, source int, tstamp int64) error
	AddNetprints(ctx context.Context, netprints []GroupNetprint) error

	//common
	//ListSource(ctx context.Context, source string) ([]Source, error)
	//Package & boxes
	GetSourceUrls(ctx context.Context) ([]SourceURL, error)
	GetNewPackages(ctx context.Context) ([]PackageNew, error)
	NewPackageUpdate(ctx context.Context, g PackageNew) error
	PackageAddWithBoxes(ctx context.Context, packages []*Package) error

	CreateOrder(ctx context.Context, o Order) error
	LoadOrder(ctx context.Context, id string) (Order, error)
	LogState(ctx context.Context, orderID string, state int, message string) error
	SetOrderState(ctx context.Context, orderID string, state int) error
	LoadAlias(ctx context.Context, alias string) (Alias, error)
	ClearGroup(ctx context.Context, source, group int, keepID string) error
	AddExtraInfo(ctx context.Context, ei OrderExtraInfo) error
	SetGroupState(ctx context.Context, source, state, group int, keepID string) error
	GetGroupState(ctx context.Context, baseID string, source, group int) (GroupState, error)
	//LoadBaseOrderByState loads one base order (limit 1) by state
	LoadBaseOrderByState(ctx context.Context, source, state int) (Order, error)
	LoadBaseOrderByChildState(ctx context.Context, source, baseState, childState int) ([]Order, error)
	FillOrders(ctx context.Context, orders []Order) error
	StartOrders(ctx context.Context, source, group int, skipID string) error
	CountCurrentOrders(ctx context.Context, source int) (int, error)
	GetCurrentOrders(ctx context.Context, source int) ([]GroupState, error)
	GetJSONMaps(ctx context.Context) (map[int][]JSONMap, error)
	GetDeliveryMaps(ctx context.Context) (map[int]map[int]DeliveryTypeMapping, error)
	Close()
}

//GroupNetprint represents the group_netprint db object
type GroupNetprint struct {
	Source     int       `json:"source" db:"source"`
	GroupID    int       `json:"group_id" db:"group_id"`
	NetprintID string    `json:"netprint_id" db:"netprint_id"`
	Created    time.Time `json:"created" db:"created"`
	State      int       `json:"state" db:"state"`
	BoxNumber  int       `json:"box_number" db:"box_number"`
	Send       bool      `json:"send" db:"send"`
}

//Alias represents the book_synonym db object
type Alias struct {
	ID       int    `json:"id" db:"id"`
	Alias    string `json:"synonym" db:"synonym"`
	Type     int    `json:"book_type" db:"book_type"`
	SubType  int    `json:"synonym_type" db:"synonym_type"`
	HasCover bool   `json:"has_cover" db:"has_cover"`
}

//Order represents the Order db object
type Order struct {
	ID          string    `json:"id" db:"id"`
	Source      int       `json:"source" db:"source"`
	SourceID    string    `json:"src_id" db:"src_id"`
	SourceDate  time.Time `json:"src_date" db:"src_date"`
	DataTS      time.Time `json:"data_ts" db:"data_ts"`
	State       int       `json:"state" db:"state"`
	StateDate   time.Time `json:"state_date" db:"state_date"`
	FtpFolder   string    `json:"ftp_folder" db:"ftp_folder"`
	LocalFolder string    `json:"local_folder" db:"local_folder"`
	FotosNum    int       `json:"fotos_num" db:"fotos_num"`
	GroupID     int       `json:"group_id" db:"group_id"`
	ClientID    int       `json:"client_id" db:"client_id"`
	Production  int       `json:"production" db:"production"`

	//4 internal use
	ExtraInfo   OrderExtraInfo
	HasCover    bool
	PrintGroups []PrintGroup
}

//GroupState is dto for orders states by GroupID
type GroupState struct {
	GroupID    int       `json:"group_id" db:"group_id"`
	BaseState  int       `json:"basestate" db:"basestate"`
	ChildState int       `json:"childstate" db:"childstate"`
	StateDate  time.Time `json:"state_date" db:"state_date"`
}

//OrderExtraInfo represents the OrderExtraInfo of db object
type OrderExtraInfo struct {
	ID            string    `json:"id" db:"id"`
	GroupID       int       `json:"group_id" db:"group_id"`
	EndPaper      string    `json:"endpaper" db:"endpaper"`
	InterLayer    string    `json:"interlayer" db:"interlayer"`
	Cover         string    `json:"cover" db:"cover"`
	Format        string    `json:"format" db:"format"`
	CornerType    string    `json:"corner_type" db:"corner_type"`
	Kaptal        string    `json:"kaptal" db:"kaptal"`
	CoverMaterial string    `json:"cover_material" db:"cover_material"`
	Weight        int       `json:"weight" db:"weight"`
	Books         int       `json:"books" db:"books"`
	Sheets        int       `json:"sheets" db:"sheets"`
	Date          time.Time `json:"date_in" db:"date_in"`
	BookThickness float32   `json:"book_thickness" db:"book_thickness"`
	Remark        string    `json:"remark" db:"remark"`
	Paper         string    `json:"paper" db:"paper"`
	Alias         string    `json:"calc_alias" db:"calc_alias"`
	Title         string    `json:"calc_title" db:"calc_title"`
}

//PrintGroup represents the PrintGroup of db object
type PrintGroup struct {
	ID         string    `json:"id" db:"id"`
	OrderID    string    `json:"order_id" db:"order_id"`
	State      int       `json:"state" db:"state"`
	StateDate  time.Time `json:"state_date" db:"state_date"`
	Width      int       `json:"width" db:"width"`
	Height     int       `json:"height" db:"height"`
	Paper      int       `json:"paper" db:"paper"`
	Frame      int       `json:"frame" db:"frame"`
	Correction int       `json:"correction" db:"correction"`
	Cutting    int       `json:"cutting" db:"cutting"`
	Path       string    `json:"path" db:"path"`
	Alias      string    `json:"alias" db:"alias"`
	FileNum    int       `json:"file_num" db:"file_num"`
	BookType   int       `json:"book_type" db:"book_type"`
	BookPart   int       `json:"book_part" db:"book_part"`
	BookNum    int       `json:"book_num" db:"book_num"`
	SheetNum   int       `json:"sheet_num" db:"sheet_num"`
	IsPDF      bool      `json:"is_pdf" db:"is_pdf"`
	IsDuplex   bool      `json:"is_duplex" db:"is_duplex"`
	Prints     int       `json:"prints" db:"prints"`
	Butt       int       `json:"butt" db:"butt"`

	//4 internal use
	Files []PrintGroupFile
}

//PrintGroupFile represents the print_group_file of db object
type PrintGroupFile struct {
	PrintGroupID string `json:"print_group" db:"print_group"`
	FileName     string `json:"file_name" db:"file_name"`
	PrintQtty    int    `json:"prt_qty" db:"prt_qty"`
	Book         int    `json:"book_num" db:"book_num"`
	Page         int    `json:"page_num" db:"page_num"`
	Caption      string `json:"caption" db:"caption"`
	BookPart     int    `json:"book_part" db:"book_part"`
}

//PackageNew represents the new mail package (package to create in cycle database)
type PackageNew struct {
	ID       int       `json:"id" db:"id"`
	Source   int       `json:"source" db:"source"`
	ClientID int       `json:"client_id" db:"client_id"`
	Created  time.Time `db:"created"`
	Attempt  int       `db:"attempt"`
	Boxes    []PackageBox
}

//DeliveryTypeMapping represents the delivery_type_dictionary
type DeliveryTypeMapping struct {
	Source       int  `json:"source" db:"source"`
	DeliveryType int  `json:"delivery_type" db:"delivery_type"`
	SiteID       int  `json:"site_id" db:"site_id"`
	SetSend      bool `json:"set_send" db:"set_send"`
}

//Package represents the  mail package (order group)
type Package struct {
	ID               int       `json:"id" db:"id"`
	Source           int       `json:"source" db:"source"`
	IDName           string    `json:"number" db:"id_name"`
	ClientID         int       `json:"client_id" db:"client_id"`
	State            int       `json:"state" db:"state"`
	StateDate        time.Time `json:"state_date" db:"state_date"`
	ExecutionDate    Date      `json:"execution_date" db:"execution_date"`
	NativeDeliveryID int       `json:"delivery_id"`
	DeliveryID       int       `db:"delivery_id"`
	DeliveryName     string    `json:"delivery_name" db:"delivery_name"`
	SrcState         int       `json:"src_state" db:"src_state"`
	SrcStateName     string    `json:"src_state_name" db:"src_state_name"`
	MailService      int       `json:"mail_service" db:"mail_service"`
	OrdersNum        int       `json:"orders_num" db:"orders_num"`
	Boxes            []PackageBox
	Properties       []PackageProperty
	Barcodes         []PackageBarcode
}

//PackageProperty represents the package property
type PackageProperty struct {
	Source    int    `json:"source" db:"source"`
	PackageID int    `json:"id" db:"id"`
	Property  string `json:"property" db:"property"`
	Value     string `json:"value" db:"value"`
}

//PackageBarcode represents the package barcode
type PackageBarcode struct {
	Source      int    `json:"source" db:"source"`
	PackageID   int    `json:"id" db:"id"`
	Barcode     string `json:"barcode" db:"barcode"`
	BarcodeType int    `json:"bar_type" db:"bar_type"`
	BoxNumber   int    `json:"box_number" db:"box_number"`
}

//PackageBox represents the package mail box
type PackageBox struct {
	Source    int       `json:"source" db:"source"`
	PackageID int       `json:"package_id" db:"package_id"`
	ID        string    `json:"box_id" db:"box_id"`
	Num       int       `json:"box_num" db:"box_num"`
	Barcode   string    `json:"barcode" db:"barcode"`
	Price     float64   `json:"price" db:"price"`
	Weight    int       `json:"weight" db:"weight"`
	State     int       `json:"state" db:"state"`
	StateDate time.Time `json:"state_date" db:"state_date"`

	Processed bool
	Items     []PackageBoxItem
}

//PackageBoxItem represents the package mail box item (order or part of order )
type PackageBoxItem struct {
	BoxID   string `json:"box_id" db:"box_id"`
	OrderID string `json:"order_id" db:"order_id"`
	Alias   string `json:"alias" db:"alias"`
	Type    string `json:"type" db:"type"`
	From    int    `json:"item_from" db:"item_from"`
	To      int    `json:"item_to" db:"item_to"`
}

//SourceURL dto to get url for api calls
type SourceURL struct {
	ID     int    `db:"id"`
	URL    string `db:"url"`
	Type   int    `db:"type"`
	AppKey string `db:"appkey"`
}

//JSONMap dto to get url for api calls
type JSONMap struct {
	SrcType   int    `db:"src_type"`
	Family    int    `db:"family"`
	AttrType  int    `db:"attr_type"`
	JSONKey   string `db:"json_key"`
	Field     string `db:"field"`
	FieldName string `db:"field_name"`
	IsList    bool   `db:"list"`
}

//Date is time.Time, used to Marshal/Unmarshal custom date format (dd.mm.yyyy)
type Date time.Time

//UnmarshalJSON  Unmarshal custom date format
func (d *Date) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		return nil
	}
	//unmarshal to string??
	var s string
	json.Unmarshal(b, &s)
	t, err := time.Parse("02.01.2006", s)
	if err == nil {
		*d = Date(t)
	}
	return err
}

//MarshalJSON  Marshal custom date format
func (d *Date) MarshalJSON() ([]byte, error) {
	s := d.format("02.01.2006")
	j, err := json.Marshal(s)
	return j, err
}

func (d Date) format(s string) string {
	t := time.Time(d)
	return t.Format(s)
}

//String implementing Stringer interface
func (d Date) String() string {
	return d.format(time.RFC3339)
}
