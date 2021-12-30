package main

import (
	"database/sql"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/jung-kurt/gofpdf"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sony/sonyflake"
)

type person struct {
	id         int
	first_name string
	last_name  string
	email      string
	ip_address string
}

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

type dashboard struct {
	Title       string
	Search_user string
	Last5_user  map[int]string
	Footer      string
}

// dashboard page var
//go:embed Templates/AllContent.html
var allContent string

func addPerson(db *sql.DB, newPerson person) {
	stmt, _ := db.Prepare("INSERT INTO people (id, first_name, last_name, email, ip_address) VALUES (?, ?, ?, ?, ?)")
	stmt.Exec(nil, newPerson.first_name, newPerson.last_name, newPerson.email, newPerson.ip_address)
	defer stmt.Close()

	fmt.Printf("Added %v %v \n", newPerson.first_name, newPerson.last_name)
}

func prepareAddPerson(db *sql.DB) {
	newPerson := person{
		first_name: "anban",
		last_name:  "gisak",
		email:      "anbangisak@gmail.com",
		ip_address: "127.0.0.1",
	}

	addPerson(db, newPerson)
}

func createInsertDefaultUser(db *sql.DB) {
	//3. create table
	_, err := db.Exec("create table USER (ID integer PRIMARY KEY, NAME string not null); delete from USER;")
	checkErr(err)

	//4. insert data
	//4.1 Begin transaction
	tx, err := db.Begin()
	checkErr(err)

	//4.2 Prepare insert stmt.
	stmt, err := tx.Prepare("insert into USER(ID, NAME) values(?, ?)")
	checkErr(err)
	defer stmt.Close()

	for i := 0; i < 10; i++ {
		_, err = stmt.Exec(i, fmt.Sprint("user-", i))
		checkErr(err)
	}

	//4.3 Commit transaction
	tx.Commit()
}

/*
func tableMulti(pdf *gofpdf.Fpdf, cols []float64, rows [][]string) {
    pagew, pageh := pdf.GetPageSize()
    _ = pagew
    mleft, mright, mtop, mbottom := pdf.GetMargins()
    _ = mleft
    _ = mright
    _ = mtop

    for _, row := range rows {
        curx, y := pdf.GetXY()
        x := curx

        height := 0.
        _, lineHt := pdf.GetFontSize()

        for i, txt := range row {
            lines := pdf.SplitLines([]byte(txt), cols[i])
            h := float64(len(lines))*lineHt + float64(MARGECELL*len(lines))
            if h > height {
                height = h
            }
        }
        // add a new page if the height of the row doesn't fit on the page
        if pdf.GetY()+height > pageh-mbottom {
            pdf.AddPage()
            y = pdf.GetY()
        }
        for i, txt := range row {
            width := cols[i]
            pdf.Rect(x, y, width, height, "")
            pdf.MultiCell(width, lineHt+MARGECELL, txt, "", "", false)
            x += width
            pdf.SetXY(x, y)
        }
        pdf.SetXY(curx, y+height)
    }
}
*/
func GeneratePdf(filename string) error {

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
	pdf.SetY(70)
	pdf.CellFormat(190, 7, "INVOICE", "0", 2, "CM", false, 0, "")

	pdf.SetFont("Arial", "", 12)
	pdf.SetXY(0, 80)
	invoice_num := genSonyflake()
	invoice_num_str := fmt.Sprintf("INVOICE NO: %x", invoice_num)
	pdf.CellFormat(100, 7, invoice_num_str, "0", 0, "CM", false, 0, "")

	pdf.SetXY(135, 80)
	date_val := time.Now()
	date_str := fmt.Sprintf("Date: %v", date_val.Format("02/01/2006"))
	pdf.CellFormat(80, 7, date_str, "0", 0, "CM", false, 0, "")

	return pdf.OutputFileAndClose(filename)
}

func main() {
	db, err := sql.Open("sqlite3", "./auro.db")
	// db, err := sql.Open("sqlite3", ":memory:")
	checkErr(err)
	defer db.Close()

	//2. fail-fast if can't connect to DB
	checkErr(db.Ping())

	// createInsertDefaultUser(db)

	//5. Query data
	rows, err := db.Query("select * from USER")
	checkErr(err)
	defer rows.Close()

	err = GeneratePdf("hello.pdf")
	if err != nil {
		panic(err)
	}

	dashboard_data := dashboard{
		Title:      "Kings Systems",
		Footer:     "Shop No: 14, Quber plaza, Opp. to Collectorate complex, Villupuram - 605602.",
		Last5_user: make(map[int]string),
	}

	//5.1 Iterate through result set
	for rows.Next() {
		var name string
		var id int
		err := rows.Scan(&id, &name)
		checkErr(err)
		fmt.Printf("id=%d, name=%s\n", id, name)
		dashboard_data.Last5_user[id] = name
	}

	fmt.Println(dashboard_data)

	//5.2 check error, if any, that were encountered during iteration
	err = rows.Err()
	checkErr(err)

	var templ *template.Template
	templ, err = template.New("allContentHtml").Parse(allContent)
	if err != nil {
		log.Fatalf("Template for Detail HTML publish failed : %v", err.Error())
	}

	// prepareAddPerson(db)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// fmt.Fprintf(w, "Hello world from GfG")
		templ.Execute(w, dashboard_data)
	})
	http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hi")
	})
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
