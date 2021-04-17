package repo

import (
	"context"
	"strings"

	"github.com/egorka-gh/photocycle"
	"github.com/jmoiron/sqlx"
)

type basicRepository struct {
	db *sqlx.DB
	//	Source   int
	readOnly bool
}

//New creates new Repository
func New(connection string, readOnly bool) (photocycle.Repository, error) {
	rep, _, err := NewTest(connection, readOnly)
	return rep, err
}

//NewTest creates new Repository, expect mysql connection sqlx.DB
func NewTest(connection string, readOnly bool) (photocycle.Repository, *sqlx.DB, error) {
	var db *sqlx.DB
	db, err := sqlx.Connect("mysql", connection)
	if err != nil {
		return nil, nil, err
	}

	return &basicRepository{
		db:       db,
		readOnly: readOnly,
	}, db, nil
}

func (b *basicRepository) Close() {
	b.db.Close()
}

func (b *basicRepository) GetSourceUrls(ctx context.Context) ([]photocycle.SourceURL, error) {
	//	var sql string = "SELECT s.id, s.type,  s1.url, s1.appkey FROM sources s INNER JOIN services s1 ON s.id = s1.src_id AND s1.srvc_id = 1 AND s1.url!='' WHERE s.online>0"
	var sql string = "SELECT s.id, s.type,  s1.url, s1.appkey, s.has_boxes FROM sources s INNER JOIN services s1 ON s.id = s1.src_id AND s1.srvc_id = 1 AND s1.url!=''"
	res := []photocycle.SourceURL{}
	err := b.db.SelectContext(ctx, &res, sql)
	return res, err
}

func (b *basicRepository) GetNewPackages(ctx context.Context) ([]photocycle.PackageNew, error) {
	//var sql string = "SELECT source, id, client_id, created, attempt FROM package_new WHERE attempt < 10"
	var sql string = "SELECT source, id, client_id, created, attempt FROM package_new"
	res := []photocycle.PackageNew{}
	err := b.db.SelectContext(ctx, &res, sql)
	return res, err
}

func (b *basicRepository) NewPackageUpdate(ctx context.Context, g photocycle.PackageNew) error {
	if b.readOnly {
		return nil
	}
	sql := "UPDATE package_new SET attempt = ? WHERE source = ? AND id = ?"
	_, err := b.db.ExecContext(ctx, sql, g.Attempt, g.Source, g.ID)
	return err
}

func (b *basicRepository) PackageAddWithBoxes(ctx context.Context, packages []*photocycle.Package) error {
	if b.readOnly || len(packages) == 0 {
		return nil
	}
	//insert packages
	oSQL := "INSERT IGNORE INTO package (source, id, client_id, state, state_date, id_name, execution_date, delivery_id, delivery_name, src_state, src_state_name, mail_service, orders_num) VALUES "
	oVals := make([]string, 0, len(packages))
	oArgs := []interface{}{}

	//TODO save props
	//INSERT INTO package_prop (source, id, property, value)
	propSQL := "INSERT IGNORE INTO package_prop (source, id, property, value) VALUES "
	propVals := make([]string, 0, len(packages)*10)
	propArgs := []interface{}{}

	//TODO save barcodes
	//INSERT INTO package_barcode (source, id, barcode, bar_type, box_number) VALUES
	barSQL := "INSERT IGNORE INTO package_barcode (source, id, barcode, bar_type, box_number) VALUES "
	barVals := make([]string, 0, len(packages)*10)
	barArgs := []interface{}{}

	xSQL := "INSERT INTO package_box (source, package_id, box_id, box_num, barcode, price, weight, state, state_date) VALUES "
	xVals := make([]string, 0, len(packages)*5)
	xArgs := []interface{}{}

	pSQL := "INSERT INTO package_box_item (box_id, order_id, alias, item_from, item_to, type ,state ,state_date) VALUES "
	pVals := make([]string, 0, len(packages)*5*3)
	pArgs := []interface{}{}
	for _, o := range packages {
		//packages
		oVals = append(oVals, "(?, ?, ?, 200, NOW(), ?, ?, ?, ?, ?, ?, ?, 0)")
		//source, id, client_id, state, state_date, id_name, execution_date, delivery_id, delivery_name, src_state, src_state_name, mail_service, orders_num
		oArgs = append(oArgs, o.Source, o.ID, o.ClientID, o.IDName, o.ExecutionDate.String(), o.DeliveryID, o.DeliveryName, o.SrcState, o.SrcStateName, o.MailService)
		//props
		for _, prop := range o.Properties {
			//source, id, property, value
			propVals = append(propVals, "(?, ?, ?, ?)")
			propArgs = append(propArgs, prop.Source, prop.PackageID, prop.Property, prop.Value)
		}
		//barcodes
		for _, bar := range o.Barcodes {
			//(source, id, barcode, bar_type, box_number)
			barVals = append(barVals, "(?, ?, ?, ?, ?)")
			barArgs = append(barArgs, bar.Source, bar.PackageID, bar.Barcode, bar.BarcodeType, bar.BoxNumber)
		}
		//boxes
		for _, x := range o.Boxes {
			xVals = append(xVals, "(?, ?, ?, ?, ?, ?, ?, 200, NOW())")
			xArgs = append(xArgs, x.Source, x.PackageID, x.ID, x.Num, x.Barcode, x.Price, x.Weight)
			//box items
			for _, p := range x.Items {
				pVals = append(pVals, "(?, ?, ?, ?, ?, ?, 200, NOW())")
				pArgs = append(pArgs, p.BoxID, p.OrderID, p.Alias, p.From, p.To, p.Type)
			}
		}
	}

	//run in transaction
	t, err := b.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	oSQL = oSQL + strings.Join(oVals, ",")
	_, err = t.Exec(oSQL, oArgs...)
	if err != nil {
		t.Rollback()
		return err
	}

	if len(propVals) > 0 {
		propSQL = propSQL + strings.Join(propVals, ",")
		_, err = t.Exec(propSQL, propArgs...)
		if err != nil {
			t.Rollback()
			return err
		}
	}
	if len(barVals) > 0 {
		barSQL = barSQL + strings.Join(barVals, ",")
		_, err = t.Exec(barSQL, barArgs...)
		if err != nil {
			t.Rollback()
			return err
		}
	}
	if len(xVals) > 0 {
		xSQL = xSQL + strings.Join(xVals, ",")
		_, err = t.Exec(xSQL, xArgs...)
		if err != nil {
			t.Rollback()
			return err
		}
	}
	if len(pVals) > 0 {
		pSQL = pSQL + strings.Join(pVals, ",")
		_, err = t.Exec(pSQL, pArgs...)
		if err != nil {
			t.Rollback()
			return err
		}
	}

	//del from package_new
	dSQL := "DELETE FROM package_new WHERE source =  ? AND id = ?"
	for _, o := range packages {
		_, err = t.Exec(dSQL, o.Source, o.ID)
		if err != nil {
			t.Rollback()
			return err
		}
	}

	return t.Commit()
}

func (b *basicRepository) GetLastNetprintSync(ctx context.Context, source int) (int64, error) {
	var sql string = "SELECT ss.np_sync_tstamp FROM sources_sync ss WHERE ss.id = ?"
	var res int64
	err := b.db.GetContext(ctx, &res, sql, source)
	return res, err
}

func (b *basicRepository) SetLastNetprintSync(ctx context.Context, source int, tstamp int64) error {
	sql := "UPDATE sources_sync ss SET ss.np_sync_tstamp = ? WHERE ss.id = ?"
	_, err := b.db.ExecContext(ctx, sql, tstamp, source)
	return err
}

func (b *basicRepository) AddNetprints(ctx context.Context, netprints []photocycle.GroupNetprint) error {
	if b.readOnly || len(netprints) < 1 {
		return nil
	}
	//batch insert
	oSQL := "INSERT IGNORE INTO group_netprint (source,group_id,netprint_id,state,box_number) VALUES "
	var oVals []string
	var oArgs []interface{}
	//limit bath size
	batch := 300
	for i, o := range netprints {
		if i%batch == 0 {
			if len(oVals) > 0 {
				//run batch
				sql := oSQL + strings.Join(oVals, ",")
				if _, err := b.db.ExecContext(ctx, sql, oArgs...); err != nil {
					return err
				}
			}
			//reset arrays
			oVals = make([]string, 0, batch)
			oArgs = make([]interface{}, 0, batch*5)
		}
		oVals = append(oVals, "(?,?,?,?,?)")
		oArgs = append(oArgs, o.Source, o.GroupID, o.NetprintID, o.State, o.BoxNumber)
	}
	if len(oVals) > 0 {
		//run batch
		sql := oSQL + strings.Join(oVals, ",")
		if _, err := b.db.ExecContext(ctx, sql, oArgs...); err != nil {
			return err
		}
	}
	return nil
}

func (b *basicRepository) CreateOrder(ctx context.Context, o photocycle.Order) error {
	if b.readOnly {
		return nil
	}
	var sb strings.Builder
	//INSERT IGNORE  ??
	sb.WriteString("INSERT INTO orders (id, source, src_id, src_date, data_ts, state, state_date, group_id, ftp_folder, fotos_num, client_id, production)")
	sb.WriteString(" VALUES (?, ?, ?, ?, ?, ?, NOW(), ?, ?, ?, ?, ?)")
	var ssql = sb.String()
	_, err := b.db.ExecContext(ctx, ssql, o.ID, o.Source, o.SourceID, o.SourceDate, o.DataTS, o.State, o.GroupID, o.FtpFolder, o.FotosNum, o.ClientID, o.Production)
	return err
}

func (b *basicRepository) FillOrders(ctx context.Context, orders []photocycle.Order) error {
	if b.readOnly {
		return nil
	}
	//insert orders
	oSQL := "INSERT INTO orders (id, source, src_id, src_date, data_ts, state, state_date, group_id, ftp_folder, fotos_num, client_id, production) VALUES "
	oVals := make([]string, 0, len(orders))
	oArgs := []interface{}{}

	xSQL := "INSERT INTO order_extra_info (id, endpaper, interlayer, cover, format, corner_type, kaptal, cover_material, books, sheets, date_in, book_thickness, group_id, remark, paper, calc_alias, calc_title, weight) VALUES "
	xVals := make([]string, 0, len(orders))
	xArgs := []interface{}{}

	pSQL := "INSERT INTO print_group (id, order_id, state, state_date, width, height, paper, frame, correction, cutting, path, alias, file_num, book_type, book_part, book_num, sheet_num, is_pdf, is_duplex, prints, butt) VALUES "
	pVals := make([]string, 0, len(orders)*2)
	pArgs := []interface{}{}

	fSQL := "INSERT INTO print_group_file (print_group, file_name, prt_qty, book_num, page_num, caption, book_part) VALUES"
	fVals := make([]string, 0, len(orders)*2*10)
	fArgs := []interface{}{}

	for _, o := range orders {
		//orders
		oVals = append(oVals, "(?, ?, ?, ?, ?, ?, NOW(), ?, ?, ?, ?, ?)")
		oArgs = append(oArgs, o.ID, o.Source, o.SourceID, o.SourceDate, o.DataTS, o.State, o.GroupID, o.FtpFolder, o.FotosNum, o.ClientID, o.Production)
		//extra info
		ei := o.ExtraInfo
		xVals = append(xVals, "(?, LEFT(?, 100), LEFT(?, 100), LEFT(?, 250), LEFT(?, 250), LEFT(?, 100), LEFT(?, 100), LEFT(?, 250), ?, ?, ?, ?, ?, LEFT(?, 250), LEFT(?, 250), LEFT(?, 50), LEFT(?, 250), ?)")
		xArgs = append(xArgs, ei.ID, ei.EndPaper, ei.InterLayer, ei.Cover, ei.Format, ei.CornerType, ei.Kaptal, ei.CoverMaterial, ei.Books, ei.Sheets, ei.Date, ei.BookThickness, ei.GroupID, ei.Remark, ei.Paper, ei.Alias, ei.Title, ei.Weight)
		//print groups
		for _, p := range o.PrintGroups {
			pVals = append(pVals, "(?, ?, ?, NOW(), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			pArgs = append(pArgs, p.ID, p.OrderID, p.State, p.Width, p.Height, p.Paper, p.Frame, p.Correction, p.Cutting, p.Path, p.Alias, p.FileNum, p.BookType, p.BookPart, p.BookNum, p.SheetNum, p.IsPDF, p.IsDuplex, p.Prints, p.Butt)
			//files
			for _, f := range p.Files {
				fVals = append(fVals, "(?, ?, ?, ?, ?, ?, ?)")
				fArgs = append(fArgs, f.PrintGroupID, f.FileName, f.PrintQtty, f.Book, f.Page, f.Caption, f.BookPart)
			}
		}
	}

	oSQL = oSQL + strings.Join(oVals, ",")
	xSQL = xSQL + strings.Join(xVals, ",")

	pSQL = pSQL + strings.Join(pVals, ",")
	fSQL = fSQL + strings.Join(fVals, ",")

	//run in transaction
	t, err := b.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	_, err = t.Exec(oSQL, oArgs...)
	if err != nil {
		t.Rollback()
		return err
	}

	_, err = t.Exec(xSQL, xArgs...)
	if err != nil {
		t.Rollback()
		return err
	}
	if len(pVals) > 0 {
		_, err = t.Exec(pSQL, pArgs...)
		if err != nil {
			t.Rollback()
			return err
		}
	}
	if len(fVals) > 0 {
		_, err = t.Exec(fSQL, fArgs...)
		if err != nil {
			t.Rollback()
			return err
		}
	}

	return t.Commit()
}

func (b *basicRepository) ClearGroup(ctx context.Context, source, group int, keepID string) error {
	if b.readOnly {
		return nil
	}
	sql := "DELETE FROM orders WHERE source =? AND group_id = ? AND ID != ?"
	_, err := b.db.ExecContext(ctx, sql, source, group, keepID)
	return err
}

func (b *basicRepository) SetGroupState(ctx context.Context, source, state, group int, keepID string) error {
	if b.readOnly {
		return nil
	}
	sql := "UPDATE orders SET state = ? WHERE source =? AND group_id = ? AND ID != ?"
	_, err := b.db.ExecContext(ctx, sql, state, source, group, keepID)
	return err
}

func (b *basicRepository) StartOrders(ctx context.Context, source, group int, skipID string) error {
	if b.readOnly {
		return nil
	}
	sql := "call pp_StartOrders(?, ?, ?)"
	_, err := b.db.ExecContext(ctx, sql, source, group, skipID)
	return err
}

func (b *basicRepository) LoadOrder(ctx context.Context, id string) (photocycle.Order, error) {
	var res photocycle.Order
	//ssql := "SELECT id, source, src_id, src_date, data_ts, state, state_date, group_id, ftp_folder, fotos_num, client_id, production FROM orders WHERE id = ?"
	ssql := "SELECT id, source, src_id, src_date, state, state_date, group_id, ftp_folder, fotos_num, client_id, production FROM orders WHERE id = ?"
	err := b.db.GetContext(ctx, &res, ssql, id)
	return res, err
}

func (b *basicRepository) LogState(ctx context.Context, orderID string, state int, message string) error {
	if b.readOnly {
		return nil
	}
	ssql := "INSERT INTO state_log (order_id, state, state_date, comment) VALUES (?, ?, NOW(), LEFT(?, 250))"
	_, err := b.db.ExecContext(ctx, ssql, orderID, state, message)
	return err
}

func (b *basicRepository) SetOrderState(ctx context.Context, orderID string, state int) error {
	if b.readOnly {
		return nil
	}
	ssql := "UPDATE orders o SET o.state = ?, o.state_date = Now() WHERE o.id = ?"
	_, err := b.db.ExecContext(ctx, ssql, state, orderID)
	return err
}

func (b *basicRepository) LoadAlias(ctx context.Context, alias string) (photocycle.Alias, error) {
	var res photocycle.Alias
	ssql := "SELECT id, synonym, book_type, synonym_type, (SELECT IFNULL(MAX(1), 0) FROM book_pg_template bpt WHERE bpt.book = bs.id AND bpt.book_part IN (1, 3, 4, 5)) has_cover FROM book_synonym bs WHERE bs.src_type = 4 AND bs.synonym = ? ORDER BY bs.synonym_type DESC"
	err := b.db.GetContext(ctx, &res, ssql, alias)
	return res, err
}

func (b *basicRepository) AddExtraInfo(ctx context.Context, ei photocycle.OrderExtraInfo) error {
	if b.readOnly {
		return nil
	}
	var sb strings.Builder
	//INSERT IGNORE  ??
	sb.WriteString("INSERT INTO order_extra_info (id, endpaper, interlayer, cover, format, corner_type, kaptal, cover_material, books, sheets, date_in, book_thickness, group_id, remark, paper, calc_alias, calc_title, weight)")
	sb.WriteString(" VALUES (?, LEFT(?, 100), LEFT(?, 100), LEFT(?, 250), LEFT(?, 250), LEFT(?, 100), LEFT(?, 100), LEFT(?, 250), ?, ?, ?, ?, ?, LEFT(?, 250), LEFT(?, 250), LEFT(?, 50), LEFT(?, 250), ?)")
	var sql = sb.String()
	_, err := b.db.ExecContext(ctx, sql, ei.ID, ei.EndPaper, ei.InterLayer, ei.Cover, ei.Format, ei.CornerType, ei.Kaptal, ei.CoverMaterial, ei.Books, ei.Sheets, ei.Date, ei.BookThickness, ei.GroupID, ei.Remark, ei.Paper, ei.Alias, ei.Title, ei.Weight)
	return err
}

func (b *basicRepository) GetGroupState(ctx context.Context, baseID string, source, group int) (photocycle.GroupState, error) {
	var res photocycle.GroupState
	sql := "SELECT IFNULL(o.group_id, 0) group_id, IFNULL(MAX(IF(o.id = ?, o.state, 0)), 0) basestate, IFNULL(MAX(IF(o.id = ?, 0, o.state)), 0) childstate, NOW() state_date FROM orders o WHERE o.source = ? AND o.group_id = ?"
	err := b.db.GetContext(ctx, &res, sql, baseID, baseID, source, group)
	return res, err
}

func (b *basicRepository) LoadBaseOrderByState(ctx context.Context, source, state int) (photocycle.Order, error) {
	var res photocycle.Order
	//ssql := "SELECT id, source, src_id, src_date, data_ts, state, state_date, group_id, ftp_folder, fotos_num, client_id, production FROM orders WHERE source = ? AND id LIKE '%@' AND state = ? LIMIT 1"
	ssql := "SELECT id, source, src_id, src_date, state, state_date, group_id, ftp_folder, fotos_num, client_id, production FROM orders WHERE source = ? AND id LIKE '%@' AND state = ? LIMIT 1"
	err := b.db.GetContext(ctx, &res, ssql, source, state)
	return res, err
}

func (b *basicRepository) LoadBaseOrderByChildState(ctx context.Context, source, baseState, childState int) ([]photocycle.Order, error) {
	res := []photocycle.Order{}
	var sb strings.Builder
	//sb.WriteString("SELECT o.id, o.source, o.src_id, o.src_date, o.data_ts, o.state, o.state_date, o.group_id, o.ftp_folder, o.fotos_num, o.client_id, o.production")
	sb.WriteString("SELECT o.id, o.source, o.src_id, o.src_date, o.state, o.state_date, o.group_id, o.ftp_folder, o.fotos_num, o.client_id, o.production")
	sb.WriteString(" FROM orders o")
	sb.WriteString(" WHERE o.source = ? AND o.id LIKE '%@' AND o.state = ? AND EXISTS (SELECT 1 FROM orders o1 WHERE o1.group_id = o.group_id AND o1.state = ?)")
	sql := sb.String()
	err := b.db.SelectContext(ctx, &res, sql, source, baseState, childState)

	return res, err
}

func (b *basicRepository) CountCurrentOrders(ctx context.Context, source int) (int, error) {
	var res int
	sql := "SELECT COUNT(DISTINCT o.group_id) FROM orders o WHERE o.state BETWEEN 100 AND 450 AND o.source = ?"
	err := b.db.GetContext(ctx, &res, sql, source)
	return res, err
}

func (b *basicRepository) GetCurrentOrders(ctx context.Context, source int) ([]photocycle.GroupState, error) {
	res := []photocycle.GroupState{}
	var sb strings.Builder
	sb.WriteString("SELECT o.group_id, MAX(o.state) basestate, MIN(o.state) childstate, MAX(o.state_date) state_date")
	sb.WriteString(" FROM orders o")
	sb.WriteString(" WHERE o.state BETWEEN 100 AND 450 AND o.source = ?")
	sb.WriteString(" GROUP BY o.group_id")
	sql := sb.String()
	err := b.db.SelectContext(ctx, &res, sql, source)
	return res, err
}

func (b *basicRepository) GetJSONMaps(ctx context.Context) (map[int][]photocycle.JSONMap, error) {
	res := []photocycle.JSONMap{}
	var sb strings.Builder
	sb.WriteString("SELECT jm.*, at.attr_fml family, at.field, at.list, at.name field_name")
	sb.WriteString(" FROM attr_json_map jm")
	sb.WriteString(" INNER JOIN attr_type at ON jm.attr_type=at.id")
	sb.WriteString(" WHERE jm.src_type IN(0,4)")
	sb.WriteString(" ORDER BY at.attr_fml, at.field")
	sql := sb.String()
	err := b.db.SelectContext(ctx, &res, sql)
	if err != nil {
		return nil, err
	}
	resMap := make(map[int][]photocycle.JSONMap)
	for _, m := range res {
		v, ok := resMap[m.Family]
		if !ok {
			v = make([]photocycle.JSONMap, 0)
		}
		resMap[m.Family] = append(v, m)
	}

	return resMap, err
}

func (b *basicRepository) GetDeliveryMaps(ctx context.Context) (map[int]map[int]photocycle.DeliveryTypeMapping, error) {
	res := []photocycle.DeliveryTypeMapping{}
	sql := "SELECT dtd.*  FROM delivery_type_dictionary dtd WHERE dtd.delivery_type!=0 ORDER BY dtd.source, dtd.delivery_type"
	err := b.db.SelectContext(ctx, &res, sql)
	if err != nil {
		return nil, err
	}
	resMap := make(map[int]map[int]photocycle.DeliveryTypeMapping)
	for _, m := range res {
		_, ok := resMap[m.Source]
		if !ok {
			resMap[m.Source] = make(map[int]photocycle.DeliveryTypeMapping)
		}
		resMap[m.Source][m.SiteID] = m
	}

	return resMap, err
}
