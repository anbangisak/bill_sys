package main

import (
	"database/sql"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/jung-kurt/gofpdf"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sony/sonyflake"
)

func genSonyflake() uint64 {
	flake := sonyflake.NewSonyflake(sonyflake.Settings{})
	id, err := flake.NextID()
	if err != nil {
		log.Fatalf("flake.NextID() failed with %s\n", err)
	}
	// Note: this is base16, could shorten by encoding as base62 string
	fmt.Printf("github.com/sony/sonyflake:      %x\n", id)
	return id
}

// TaxInfo struct
type TaxInfo struct {
	Id            int
	Name          string
	InvoiceNumber string
	Date          string
	TanNumber     string
	Fy            string
	OfficeName    string
	// Description   string
	Description1 string
	Description2 string
	Description3 string
	Amount       string
	AmountInWord string
}

func dbConn() (db *sql.DB) {
	db, err := sql.Open("sqlite3", "./auro.db")
	// db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err.Error())
	}
	return db
}

func createTaxInfo(db *sql.DB) {
	varCharStr := "varchar(80) DEFAULT NULL"
	varCharDesc := "varchar(255) DEFAULT NULL"
	_, err := db.Exec("CREATE TABLE taxinfo (id INTEGER PRIMARY KEY AUTOINCREMENT, name " + varCharStr + ", invoicenumber " + varCharStr + ", dateval " + varCharStr + ", tannumber " + varCharStr + ", fy " + varCharStr + ", officename " + varCharStr + ", description1 " + varCharDesc + ", description2 " + varCharDesc + ", description3 " + varCharDesc + ", amount " + varCharStr + ", amountinword " + varCharStr + ")")
	checkErr(err)
}

func GeneratePdf(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	nId := r.URL.Query().Get("id")
	selDB, err := db.Query("SELECT * FROM taxinfo WHERE id=?", nId)
	if err != nil {
		panic(err.Error())
	}

	taxInfo := TaxInfo{}

	for selDB.Next() {
		var id int
		var name, invoiceNumber, dateVal, tanNumber, fy, officeName, desc1, desc2, desc3, amount, amountInWord string
		err := selDB.Scan(&id, &name, &invoiceNumber, &dateVal, &tanNumber, &fy, &officeName, &desc1, &desc2, &desc3, &amount, &amountInWord)
		if err != nil {
			panic(err.Error())
		}

		log.Println("Listing Row: Id " + string(id) + " | name " + name + " | invoiceNumber " + invoiceNumber + " | dateVal " + dateVal + " | tanNumber " + tanNumber + " | fy " + fy + " | officeName " + officeName + " | description1 " + desc1 + " | description2 " + desc2 + " | description3 " + desc3 + " | amount " + fmt.Sprintf("%f", amount) + " | amountInWord " + amountInWord)

		taxInfo.Id = id
		taxInfo.Name = name
		taxInfo.InvoiceNumber = invoiceNumber
		taxInfo.Date = dateVal
		taxInfo.TanNumber = tanNumber
		taxInfo.Fy = fy
		taxInfo.OfficeName = officeName
		taxInfo.Description1 = desc1
		taxInfo.Description2 = desc2
		taxInfo.Description3 = desc3
		taxInfo.Amount = amount
		taxInfo.AmountInWord = amountInWord
	}
	// tmpl.ExecuteTemplate(w, "Show", taxInfo)
	defer db.Close()
	// func GeneratePdf(filename string) error {

	// const (
	// 	colCount = 4
	// 	rowCount = 3
	// 	margin   = 32.0
	// 	fontHt   = 14.0 // point
	// )
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// ImageOptions(src, x, y, width, height, flow, options, link, linkStr)
	pdf.ImageOptions(
		"logo.png",
		85, 5,
		40, 30,
		false,
		gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true},
		0,
		"",
	)

	pdf.SetFont("Arial", "B", 28)
	pdf.SetY(40)
	// CellFormat(width, height, text, border, position after, align, fill, link, linkStr)
	pdf.CellFormat(190, 7, "KINGS SYSTEMS", "0", 2, "CM", false, 0, "")

	pdf.SetFont("Arial", "", 16)
	pdf.SetY(50)
	pdf.CellFormat(190, 7, "(ONLINE e-TDS/TCS SOLUTIONS)", "0", 2, "CM", false, 0, "")

	pdf.SetFont("Arial", "BU", 18)
	pdf.SetY(60)
	pdf.CellFormat(190, 7, "INVOICE", "0", 2, "CM", false, 0, "")

	pdf.SetFont("Arial", "", 12)
	pdf.SetXY(0, 70)
	// invoice_num := genSonyflake()
	// invoice_num_str := fmt.Sprintf("INVOICE NO: %x", invoice_num)
	invoice_num_str := "INVOICE NO: " + taxInfo.InvoiceNumber
	pdf.CellFormat(100, 7, invoice_num_str, "0", 0, "CM", false, 0, "")

	pdf.SetXY(135, 70)
	// date_val := time.Now()
	date_str := "Date: " + taxInfo.Date
	pdf.CellFormat(80, 7, date_str, "0", 0, "CM", false, 0, "")

	// tanOffiName table
	pdf.SetFont("Arial", "B", 11)
	pdf.SetXY(20, 80)
	pdf.CellFormat(30.0, 10.0, "TAN NO", "1", 0, "LM", false, 0, "")
	tanNo := taxInfo.TanNumber
	pdf.SetFont("Arial", "", 11)
	pdf.SetXY(50, 80)
	pdf.CellFormat(88.0, 10.0, tanNo, "1", 0, "LM", false, 0, "")
	pdf.SetFont("Arial", "B", 11)
	pdf.SetXY(20, 90)
	pdf.CellFormat(30.0, 24.0, "OFFICE NAME", "1", 0, "LM", false, 0, "")
	officeName := taxInfo.OfficeName
	pdf.SetFont("Arial", "", 11)
	pdf.SetXY(50, 90)
	pdf.CellFormat(88.0, 24.0, officeName, "1", 0, "LM", false, 0, "")

	// desc & amount table
	pdf.SetFont("Arial", "", 11)
	pdf.SetXY(20, 124)
	pdf.CellFormat(110.0, 14.0, "DESCRIPTION", "1", 0, "CM", false, 0, "")

	pdf.SetXY(130, 124)
	pdf.CellFormat(30.0, 14.0, "AMOUNT", "1", 0, "LM", false, 0, "")

	ht := pdf.PointConvert(8)
	// fyStr := taxInfo.Fy

	desc1 := taxInfo.Description1
	desc2 := taxInfo.Description2
	desc3 := taxInfo.Description3

	pdf.SetXY(20, 138)
	pdf.CellFormat(110.0, 24.0, "", "1", 1, "LM", false, 0, "")
	pdf.SetXY(28, 143)
	pdf.CellFormat(110.0, ht, desc1, "", 1, "LM", false, 0, "")
	pdf.Ln(ht)
	pdf.SetXY(28, 151)
	pdf.CellFormat(110.0, ht, desc2, "", 1, "LM", false, 0, "")
	pdf.Ln(ht)
	pdf.SetXY(28, 159)
	pdf.CellFormat(110.0, ht, desc3, "", 1, "LM", false, 0, "")
	pdf.Ln(ht)
	amount := taxInfo.Amount
	pdf.SetXY(130, 138)
	pdf.CellFormat(30.0, 24.0, fmt.Sprintf("%v", amount), "1", 0, "LM", false, 0, "")

	pdf.SetXY(20, 162)
	pdf.CellFormat(110.0, 20.0, "TOTAL BILL AMOUNT", "1", 0, "RM", false, 0, "")

	pdf.SetXY(130, 162)
	pdf.CellFormat(30.0, 20.0, fmt.Sprintf("%v", amount), "1", 0, "LM", false, 0, "")

	pdf.SetXY(15, 182)
	// date_val := time.Now()
	rupeesStr := "(Rupees " + taxInfo.AmountInWord + ")"
	pdf.CellFormat(110, 10, rupeesStr, "0", 0, "CM", false, 0, "")

	pdf.SetXY(20, 202)
	pdf.CellFormat(110.0, 40.0, "", "", 1, "LM", false, 0, "")
	pdf.SetXY(28, 202)
	pdf.CellFormat(110.0, ht, "Remarks", "", 1, "LM", false, 0, "")
	pdf.Ln(ht)
	pdf.SetXY(28, 210)
	pdf.CellFormat(110.0, ht, "Kindly make payments to,", "", 1, "LM", false, 0, "")
	pdf.Ln(ht)
	pdf.SetXY(28, 218)
	pdf.CellFormat(110.0, ht, "Name: J.Himmam Hussain", "", 1, "LM", false, 0, "")
	pdf.Ln(ht)
	pdf.SetXY(28, 224)
	pdf.CellFormat(110.0, ht, "Account No: 2697101023500", "", 1, "LM", false, 0, "")
	pdf.Ln(ht)
	pdf.SetXY(28, 232)
	pdf.CellFormat(110.0, ht, "IFSC Code : CNRB0002697", "", 1, "LM", false, 0, "")

	pdf.SetXY(135, 232)
	pdf.CellFormat(70.0, 24.0, "", "", 1, "LM", false, 0, "")
	pdf.SetXY(135, 232)
	pdf.CellFormat(70.0, ht, "(J.Himmam Hussain)", "", 1, "CM", false, 0, "")
	pdf.Ln(ht)
	pdf.SetXY(135, 240)
	pdf.CellFormat(70.0, ht, "FOR KINGS SYSTEMS,", "", 1, "CM", false, 0, "")
	pdf.Ln(ht)
	pdf.SetXY(135, 248)
	pdf.CellFormat(70.0, ht, "VILLUPURAM", "", 1, "CM", false, 0, "")

	pdf.Line(0, 253, 250, 253)

	ht = pdf.PointConvert(6)
	pdf.SetXY(35, 258)
	pdf.CellFormat(150.0, 18.0, "", "", 1, "LM", false, 0, "")
	pdf.SetXY(35, 258)
	pdf.CellFormat(150.0, ht, "Shop No: 14, Quber plaza, Opp. to Collectorate complex, Villupuram - 605602.", "", 1, "CM", false, 0, "")
	pdf.Ln(ht)
	pdf.SetXY(35, 264)
	pdf.CellFormat(150.0, ht, "Mobile : 9994471707, Landline : 04146 - 227707, 04146 - 356394", "", 1, "CM", false, 0, "")
	pdf.Ln(ht)
	pdf.SetXY(35, 270)
	pdf.CellFormat(150.0, ht, "E-mail : kingsetdscenter@gmail.com", "", 1, "CM", false, 0, "")

	err = pdf.OutputFileAndClose(taxInfo.TanNumber + ".pdf")
	if err != nil {
		panic(err.Error())
	}
	http.ServeFile(w, r, taxInfo.TanNumber+".pdf")

}

var tmpl = template.Must(template.ParseGlob("Templates/*"))

func GetTaxInfo(selDB *sql.Rows) TaxInfo {
    taxInfo := TaxInfo{}

	for selDB.Next() {
		var id int
		var name, invoiceNumber, dateVal, tanNumber, fy, officeName, desc1, desc2, desc3, amount, amountInWord string
		err := selDB.Scan(&id, &name, &invoiceNumber, &dateVal, &tanNumber, &fy, &officeName, &desc1, &desc2, &desc3, &amount, &amountInWord)
		if err != nil {
			panic(err.Error())
		}

		taxInfo.Id = id
		taxInfo.Name = name
		taxInfo.InvoiceNumber = invoiceNumber
		taxInfo.Date = dateVal
		taxInfo.TanNumber = tanNumber
		taxInfo.Fy = fy
		taxInfo.OfficeName = officeName
		taxInfo.Description1 = desc1
		taxInfo.Description2 = desc2
		taxInfo.Description3 = desc3
		taxInfo.Amount = amount
		taxInfo.AmountInWord = amountInWord
	}
	return taxInfo
}

func GetInvoiceNum() int {
    db := dbConn()
	defer db.Close()
	selDB, err := db.Query("SELECT * FROM taxinfo ORDER BY id DESC LIMIT 1")
	if err != nil {
		panic(err.Error())
	}
	taxInfo := GetTaxInfo(selDB)
	invoice_num := 1001
	prev_invoice_num, err := strconv.Atoi(taxInfo.InvoiceNumber)
	if err != nil {
		panic(err.Error())
	}else if (9999> prev_invoice_num) && (prev_invoice_num >1000){
	    invoice_num = prev_invoice_num + 1
	}
	return invoice_num
}

func New(w http.ResponseWriter, r *http.Request) {
    invoice_num := GetInvoiceNum()
	taxInfo := TaxInfo{InvoiceNumber: fmt.Sprintf("%d", invoice_num)}
	tmpl.ExecuteTemplate(w, "New", taxInfo)
}

func NewClone(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	nId := r.URL.Query().Get("id")
	selDB, err := db.Query("SELECT * FROM taxinfo WHERE id=?", nId)
	if err != nil {
		panic(err.Error())
	}
    invoice_num := GetInvoiceNum()
	taxInfo := TaxInfo{}

	for selDB.Next() {
		var id int
		var name, invoiceNumber, dateVal, tanNumber, fy, officeName, desc1, desc2, desc3, amount, amountInWord string
		err := selDB.Scan(&id, &name, &invoiceNumber, &dateVal, &tanNumber, &fy, &officeName, &desc1, &desc2, &desc3, &amount, &amountInWord)
		if err != nil {
			panic(err.Error())
		}

		taxInfo.Id = id
		taxInfo.Name = name
		taxInfo.InvoiceNumber = fmt.Sprintf("%d", invoice_num)
		taxInfo.Date = dateVal
		taxInfo.TanNumber = tanNumber
		taxInfo.Fy = fy
		taxInfo.OfficeName = officeName
		taxInfo.Description1 = desc1
		taxInfo.Description2 = desc2
		taxInfo.Description3 = desc3
		taxInfo.Amount = amount
		taxInfo.AmountInWord = amountInWord
	}

	tmpl.ExecuteTemplate(w, "NewClone", taxInfo)
	defer db.Close()
}

//Index handler
func Index(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	selDB, err := db.Query("SELECT * FROM taxinfo ORDER BY id DESC")
	// ("select * from product where name like ?", keyword + "%")
	if err != nil {
		panic(err.Error())
	}

	taxInfo := TaxInfo{}
	res := []TaxInfo{}

	for selDB.Next() {
		var id int
		var name, invoiceNumber, dateVal, tanNumber, fy, officeName, desc1, desc2, desc3, amount, amountInWord string
		err := selDB.Scan(&id, &name, &invoiceNumber, &dateVal, &tanNumber, &fy, &officeName, &desc1, &desc2, &desc3, &amount, &amountInWord)
		if err != nil {
			panic(err.Error())
		}
		log.Println("Listing Row: Id " + string(id) + " | name " + name + " | invoiceNumber " + invoiceNumber + " | dateVal " + dateVal + " | tanNumber " + tanNumber + " | fy " + fy + " | officeName " + officeName + " | description1 " + desc1 + " | description2 " + desc2 + " | description3 " + desc3 + " | amount " + fmt.Sprintf("%f", amount) + " | amountInWord " + amountInWord)

		taxInfo.Id = id
		taxInfo.Name = name
		taxInfo.InvoiceNumber = invoiceNumber
		taxInfo.Date = dateVal
		taxInfo.TanNumber = tanNumber
		taxInfo.Fy = fy
		taxInfo.OfficeName = officeName
		taxInfo.Description1 = desc1
		taxInfo.Description2 = desc2
		taxInfo.Description3 = desc3
		taxInfo.Amount = amount
		taxInfo.AmountInWord = amountInWord
		res = append(res, taxInfo)
	}
	tmpl.ExecuteTemplate(w, "Index", res)
	defer db.Close()
}

//Show handler
func Show(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	nId := r.URL.Query().Get("id")
	selDB, err := db.Query("SELECT * FROM taxinfo WHERE id=?", nId)
	if err != nil {
		panic(err.Error())
	}

	taxInfo := TaxInfo{}

	for selDB.Next() {
		var id int
		var name, invoiceNumber, dateVal, tanNumber, fy, officeName, desc1, desc2, desc3, amount, amountInWord string
		err := selDB.Scan(&id, &name, &invoiceNumber, &dateVal, &tanNumber, &fy, &officeName, &desc1, &desc2, &desc3, &amount, &amountInWord)
		if err != nil {
			panic(err.Error())
		}

		log.Println("Listing Row: Id " + string(id) + " | name " + name + " | invoiceNumber " + invoiceNumber + " | dateVal " + dateVal + " | tanNumber " + tanNumber + " | fy " + fy + " | officeName " + officeName + " | description1 " + desc1 + " | description2 " + desc2 + " | description3 " + desc3 + " | amount " + fmt.Sprintf("%f", amount) + " | amountInWord " + amountInWord)

		taxInfo.Id = id
		taxInfo.Name = name
		taxInfo.InvoiceNumber = invoiceNumber
		taxInfo.Date = dateVal
		taxInfo.TanNumber = tanNumber
		taxInfo.Fy = fy
		taxInfo.OfficeName = officeName
		taxInfo.Description1 = desc1
		taxInfo.Description2 = desc2
		taxInfo.Description3 = desc3
		taxInfo.Amount = amount
		taxInfo.AmountInWord = amountInWord
	}
	tmpl.ExecuteTemplate(w, "Show", taxInfo)
	defer db.Close()
}

func Edit(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	nId := r.URL.Query().Get("id")
	selDB, err := db.Query("SELECT * FROM taxinfo WHERE id=?", nId)
	if err != nil {
		panic(err.Error())
	}

	taxInfo := TaxInfo{}

	for selDB.Next() {
		var id int
		var name, invoiceNumber, dateVal, tanNumber, fy, officeName, desc1, desc2, desc3, amount, amountInWord string
		err := selDB.Scan(&id, &name, &invoiceNumber, &dateVal, &tanNumber, &fy, &officeName, &desc1, &desc2, &desc3, &amount, &amountInWord)
		if err != nil {
			panic(err.Error())
		}

		taxInfo.Id = id
		taxInfo.Name = name
		taxInfo.InvoiceNumber = invoiceNumber
		taxInfo.Date = dateVal
		taxInfo.TanNumber = tanNumber
		taxInfo.Fy = fy
		taxInfo.OfficeName = officeName
		taxInfo.Description1 = desc1
		taxInfo.Description2 = desc2
		taxInfo.Description3 = desc3
		taxInfo.Amount = amount
		taxInfo.AmountInWord = amountInWord
	}

	tmpl.ExecuteTemplate(w, "Edit", taxInfo)
	defer db.Close()
}

func Insert(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	if r.Method == "POST" {
		name := r.FormValue("name")
		invoiceNumber := r.FormValue("invoicenumber")
		dateVal := r.FormValue("dateval")
		tanNumber := r.FormValue("tannumber")
		fy := r.FormValue("fy")
		officeName := r.FormValue("officename")
		description1 := r.FormValue("desc1")
		description2 := r.FormValue("desc2")
		description3 := r.FormValue("desc3")
		amount := r.FormValue("amount")
		amountInWord := r.FormValue("amountinword")

		insForm, err := db.Prepare("INSERT INTO taxinfo (name, invoicenumber, dateval, tannumber, fy, officename, description1, description2, description3, amount, amountinword) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(name, invoiceNumber, dateVal, tanNumber, fy, officeName, description1, description2, description3, amount, amountInWord)
		log.Println("Insert Data: name " + name + " | invoiceNumber " + invoiceNumber + " | dateVal " + dateVal + " | tanNumber " + tanNumber + " | fy " + fy + " | officeName " + officeName + " | description1 " + description1 + " | description2 " + description2 + " | description3 " + description3 + " | amount " + amount + " | amountInWord " + amountInWord)
	}
	defer db.Close()
	http.Redirect(w, r, "/", 301)
}

func Update(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	if r.Method == "POST" {
		name := r.FormValue("name")
		invoiceNumber := r.FormValue("invoicenumber")
		dateVal := r.FormValue("dateval")
		tanNumber := r.FormValue("tannumber")
		fy := r.FormValue("fy")
		officeName := r.FormValue("officename")
		description1 := r.FormValue("desc1")
		description2 := r.FormValue("desc2")
		description3 := r.FormValue("desc3")
		amount := r.FormValue("amount")
		amountInWord := r.FormValue("amountinword")
		id := r.FormValue("uid")
		insForm, err := db.Prepare("UPDATE taxinfo SET name=?, invoicenumber=?, dateval=?, tannumber=?, fy=?, officename=?, description1=?, description2=?, description3=?, amount=?, amountinword=? WHERE id=?")
		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(name, invoiceNumber, dateVal, tanNumber, fy, officeName, description1, description2, description3, amount, amountInWord, id)
		log.Println("UPDATE Data: name " + name + " | invoiceNumber " + invoiceNumber + " | dateVal " + dateVal + " | tanNumber " + tanNumber + " | fy " + fy + " | officeName " + officeName + " | description1 " + description1 + " | description2 " + description2 + " | description3 " + description3 + " | amount " + amount + " | amountInWord " + amountInWord)
	}
	defer db.Close()
	http.Redirect(w, r, "/", 301)
}

func Search(w http.ResponseWriter, r *http.Request){
    db := dbConn()
	if r.Method == "GET"{
	    tmpl.ExecuteTemplate(w, "Search", nil)
	}else if r.Method == "POST"{
	    tanNumber := r.FormValue("tannumber")
		selDB, err := db.Query("SELECT * FROM taxinfo WHERE tannumber LIKE '"+ tanNumber + "%' ORDER BY id DESC LIMIT 1000")
		if err != nil {
			panic(err.Error())
		}

		taxInfo := TaxInfo{}
		res := []TaxInfo{}

		for selDB.Next() {
			var id int
			var name, invoiceNumber, dateVal, tanNumber, fy, officeName, desc1, desc2, desc3, amount, amountInWord string
			err := selDB.Scan(&id, &name, &invoiceNumber, &dateVal, &tanNumber, &fy, &officeName, &desc1, &desc2, &desc3, &amount, &amountInWord)
			if err != nil {
				panic(err.Error())
			}
			log.Println("Listing Row: Id " + string(id) + " | name " + name + " | invoiceNumber " + invoiceNumber + " | dateVal " + dateVal + " | tanNumber " + tanNumber + " | fy " + fy + " | officeName " + officeName + " | description1 " + desc1 + " | description2 " + desc2 + " | description3 " + desc3 + " | amount " + fmt.Sprintf("%f", amount) + " | amountInWord " + amountInWord)

			taxInfo.Id = id
			taxInfo.Name = name
			taxInfo.InvoiceNumber = invoiceNumber
			taxInfo.Date = dateVal
			taxInfo.TanNumber = tanNumber
			taxInfo.Fy = fy
			taxInfo.OfficeName = officeName
			taxInfo.Description1 = desc1
			taxInfo.Description2 = desc2
			taxInfo.Description3 = desc3
			taxInfo.Amount = amount
			taxInfo.AmountInWord = amountInWord
			res = append(res, taxInfo)
		}
		tmpl.ExecuteTemplate(w, "Index", res)
		defer db.Close()
	}
}

func main() {

	// use first time, when need to create db
	/*
		db := dbConn()
		defer db.Close()
		// fail-fast if can't connect to DB
		checkErr(db.Ping())
		createTaxInfo(db)
	*/

	http.HandleFunc("/new", New)
	http.HandleFunc("/", Index)
	http.HandleFunc("/show", Show)
	http.HandleFunc("/insert", Insert)
	http.HandleFunc("/edit", Edit)
	http.HandleFunc("/update", Update)
	http.HandleFunc("/exportpdf", GeneratePdf)
	http.HandleFunc("/clone-record", NewClone)
	http.HandleFunc("/search", Search)
	//http.HandleFunc("/list-search", ListSearch)
	port := ":8200"
	fmt.Println("Server is running on port" + port)

	// Start server on port specified above
	log.Fatal(http.ListenAndServe(port, nil))
	fmt.Println("Hello World!")
}

func checkErr(err error, args ...string) {
	if err != nil {
		fmt.Println("Error")
		fmt.Println("%q: %s", err, args)
	}
}
