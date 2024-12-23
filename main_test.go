package structsqlpostgres

import (
	"testing"
)

// Test struct for all the tests
type TestStruct struct {
	ID    int64 `json:"test_struct_id"`
	Flags int64 `json:"test_struct_flags"`

	// Test email validation
	PrimaryEmail   string `json:"email"`
	EmailSecondary string `json:"email2"`

	// Test length validation
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`

	// Test int value validation
	Age   int `json:"age"`
	Price int `json:"price"`

	// Test regular expression
	PostCode  string `json:"post_code"`
	PostCode2 string `json:"post_code2"`

	// Test HTTP endpoint tags
	Password        string `json:"password"`
	CreatedByUserID int64  `json:"created_by_user_id"`

	// Test unique tag
	Key string `json:"key" 2sql:"uniq db_type:varchar(2000)"`
}

// Test structs for INNER JOIN
type ProductKind struct {
	ID   int64
	Name string
}

type ProductGroup struct {
	ID          int64
	Name        string
	Description string
	Code        string
}

type Product struct {
	ID            int64
	Name          string
	Price         int
	ProductKindID int64
	ProductGrpID  int64
}

type Product_WithDetails struct {
	ID               int64
	Name             string
	Price            int
	ProductKindID    int64
	ProductGrpID     int64
	ProductKind      *ProductKind `2sql:"join"`
	ProductKind_Name string
	ProductGrp       *ProductGroup `2sql:"join"`
	ProductGrp_Code  string
}

// Instance of the test object
var testStructObj = &TestStruct{}

func TestSQLQueries(t *testing.T) {
	h := NewStructSQL(testStructObj, StructSQLOptions{})

	got := h.GetQueryDropTable()
	want := "DROP TABLE IF EXISTS test_structs"
	if got != want {
		t.Fatalf("Want %v, got %v", want, got)
	}

	got = h.GetQueryCreateTable()
	want = "CREATE TABLE test_structs (test_struct_id SERIAL PRIMARY KEY,test_struct_flags BIGINT NOT NULL DEFAULT 0,primary_email VARCHAR(255) NOT NULL DEFAULT '',email_secondary VARCHAR(255) NOT NULL DEFAULT '',first_name VARCHAR(255) NOT NULL DEFAULT '',last_name VARCHAR(255) NOT NULL DEFAULT '',age BIGINT NOT NULL DEFAULT 0,price BIGINT NOT NULL DEFAULT 0,post_code VARCHAR(255) NOT NULL DEFAULT '',post_code2 VARCHAR(255) NOT NULL DEFAULT '',password VARCHAR(255) NOT NULL DEFAULT '',created_by_user_id BIGINT NOT NULL DEFAULT 0,key VARCHAR(2000) NOT NULL DEFAULT '' UNIQUE)"
	if got != want {
		t.Fatalf("Want %v, got %v", want, got)
	}
}

func TestSQLInsertQueries(t *testing.T) {
	h := NewStructSQL(testStructObj, StructSQLOptions{})

	got := h.GetQueryInsert()
	want := "INSERT INTO test_structs(test_struct_flags,primary_email,email_secondary,first_name,last_name,age,price,post_code,post_code2,password,created_by_user_id,key) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING test_struct_id"
	if got != want {
		t.Fatalf("Want %v, got %v", want, got)
	}
}

func TestSQLUpdateByIdQueries(t *testing.T) {
	h := NewStructSQL(testStructObj, StructSQLOptions{})

	got := h.GetQueryUpdateById()
	want := "UPDATE test_structs SET test_struct_flags=$1,primary_email=$2,email_secondary=$3,first_name=$4,last_name=$5,age=$6,price=$7,post_code=$8,post_code2=$9,password=$10,created_by_user_id=$11,key=$12 WHERE test_struct_id = $13"
	if got != want {
		t.Fatalf("Want %v, got %v", want, got)
	}
}

func TestSQLInsertOnConflictUpdateQueries(t *testing.T) {
	h := NewStructSQL(testStructObj, StructSQLOptions{})

	got := h.GetQueryInsertOnConflictUpdate()
	want := "INSERT INTO test_structs(test_struct_id,test_struct_flags,primary_email,email_secondary,first_name,last_name,age,price,post_code,post_code2,password,created_by_user_id,key)"
	want += " VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)"
	want += " ON CONFLICT (test_struct_id) DO UPDATE SET test_struct_flags=$14,primary_email=$15,email_secondary=$16,first_name=$17,last_name=$18,age=$19,price=$20,post_code=$21,post_code2=$22,password=$23,created_by_user_id=$24,key=$25"
	want += " RETURNING test_struct_id"
	if got != want {
		t.Fatalf("Want %v, got %v", want, got)
	}
}

func TestSQLDeleteQueries(t *testing.T) {
	h := NewStructSQL(testStructObj, StructSQLOptions{})

	got := h.GetQueryDeleteById()
	want := "DELETE FROM test_structs WHERE test_struct_id = $1"
	if got != want {
		t.Fatalf("Want %v, got %v", want, got)
	}
}

func TestSQLSelectQueries(t *testing.T) {
	h := NewStructSQL(testStructObj, StructSQLOptions{})

	selectPrefix := "SELECT test_struct_id,test_struct_flags,primary_email,email_secondary,first_name,last_name,age,price,post_code,post_code2,password,created_by_user_id,key FROM test_structs"

	got := h.GetQuerySelectById()
	want := selectPrefix + " WHERE test_struct_id = $1"
	if got != want {
		t.Fatalf("Want %v, got %v", want, got)
	}

	got = h.GetQuerySelect(nil, 67, 13, nil, nil, nil)
	want = selectPrefix + " LIMIT 67 OFFSET 13"
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}

	got = h.GetQuerySelect([]string{"EmailSecondary", "desc", "Age", "asc"}, 67, 13, map[string]interface{}{"Price": 4444, "PostCode2": "11-111"}, nil, nil)
	want = selectPrefix + " WHERE post_code2=$1 AND price=$2 ORDER BY email_secondary DESC,age ASC LIMIT 67 OFFSET 13"
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}

	got = h.GetQuerySelect([]string{"EmailSecondary", "desc", "Age", "asc"}, 67, 13, map[string]interface{}{"Price": 4444, "PostCode2": "11-111"}, map[string]bool{"EmailSecondary": true}, nil)
	want = selectPrefix + " WHERE post_code2=$1 AND price=$2 ORDER BY email_secondary DESC LIMIT 67 OFFSET 13"
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}

	got = h.GetQuerySelect([]string{"EmailSecondary", "desc", "Age", "asc"}, 67, 13, map[string]interface{}{"Price": 4444, "PostCode2": "11-111"}, map[string]bool{"EmailSecondary": true}, map[string]bool{"Price": true})
	want = selectPrefix + " WHERE price=$1 ORDER BY email_secondary DESC LIMIT 67 OFFSET 13"
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}

	got = h.GetQuerySelect([]string{"EmailSecondary", "asc", "Age", "desc"}, 1, 3, map[string]interface{}{
		"Price":     33,
		"PostCode2": "11-111",
		"_raw": []interface{}{
			".Price=? OR (.EmailSecondary=? AND .Age IN (?)) OR (.Age IN (?)) OR (.EmailSecondary IN (?))",
			// We do not really care about the values, the query contains $x only symbols
			// However, we need to pass either value or an array so that an array can be extracted into multiple $x's
			0,
			0,
			[]int{0, 0, 0, 0},
			[]int{0, 0, 0},
			[]int{0, 0},
		},
		"_rawConjuction": RawConjuctionOR,
	}, map[string]bool{
		"EmailSecondary": true,
	}, map[string]bool{
		"Price": true,
	})
	want = selectPrefix + " WHERE"
	want += " (price=$1) OR (price=$2 OR (email_secondary=$3 AND age IN ($4,$5,$6,$7)) OR (age IN ($8,$9,$10)) OR (email_secondary IN ($11,$12)))"
	want += " ORDER BY email_secondary ASC LIMIT 1 OFFSET 3"
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}

	got = h.GetQuerySelect([]string{"EmailSecondary", "desc", "Age", "asc"}, 67, 13, map[string]interface{}{
		"Price:>":            4443,
		"PostCode2:%":        "11%",
		"Age:<=":             99,
		"FirstName:~":        "^[A-Z][a-z]+$",
		"CreatedByUserID:>=": 100,
		"Flags:&":            8,
	}, map[string]bool{"EmailSecondary": true}, nil)
	want = selectPrefix + " WHERE age<=$1 AND created_by_user_id>=$2 AND first_name ~ $3 AND test_struct_flags&$4>0 AND post_code2 LIKE $5 AND price>$6"
	want += " ORDER BY email_secondary DESC LIMIT 67 OFFSET 13"
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}

	got = h.GetQuerySelect([]string{"EmailSecondary", "desc", "Age", "asc"}, 67, 13, map[string]interface{}{
		"Price:%": 44,
		"Age:~":   99,
	}, map[string]bool{"EmailSecondary": true}, nil)
	want = selectPrefix + " WHERE CAST(age AS TEXT) ~ $1 AND CAST(price AS TEXT) LIKE $2"
	want += " ORDER BY email_secondary DESC LIMIT 67 OFFSET 13"
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestSQLSelectCountQueries(t *testing.T) {
	h := NewStructSQL(testStructObj, StructSQLOptions{})

	got := h.GetQuerySelectCount(map[string]interface{}{"Price": 4444, "PostCode2": "11-111"}, map[string]bool{"Price": true})
	want := "SELECT COUNT(*) AS cnt FROM test_structs WHERE price=$1"
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestSQLDeleteWithFiltersQueries(t *testing.T) {
	h := NewStructSQL(testStructObj, StructSQLOptions{})

	got := h.GetQueryDelete(map[string]interface{}{"Price": 4444, "PostCode2": "11-111"}, map[string]bool{"Price": true})
	want := "DELETE FROM test_structs WHERE price=$1"
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}

	got = h.GetQueryDelete(map[string]interface{}{
		"Price":          4444,
		"PostCode2":      "11-111",
		"_rawConjuction": RawConjuctionAND,
		"_raw": []interface{}{
			".Price=? OR .EmailSecondary=? OR .Age IN (?)",
			0,
			0,
			[]int{0, 0, 0},
		},
	}, map[string]bool{"Price": true})
	want = "DELETE FROM test_structs WHERE (price=$1) AND (price=$2 OR email_secondary=$3 OR age IN ($4,$5,$6))"
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}

	got = h.GetQueryDelete(map[string]interface{}{
		"_raw": []interface{}{
			".Price=? OR .EmailSecondary=? OR .Age IN (?)",
			0,
			0,
			[]int{0, 0, 0},
		},
	}, map[string]bool{"Price": true})
	want = "DELETE FROM test_structs WHERE (price=$1 OR email_secondary=$2 OR age IN ($3,$4,$5))"
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}

	got = h.GetQueryDeleteReturningID(map[string]interface{}{
		"_raw": []interface{}{
			".Price=? OR .EmailSecondary=? OR .Age IN (?)",
			0,
			0,
			[]int{0, 0, 0},
		},
	}, map[string]bool{"Price": true})
	want = "DELETE FROM test_structs WHERE (price=$1 OR email_secondary=$2 OR age IN ($3,$4,$5)) RETURNING test_struct_id"
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestSQLUpdateQueries(t *testing.T) {
	h := NewStructSQL(testStructObj, StructSQLOptions{})

	got := h.GetQueryUpdate(
		map[string]interface{}{"Price": 1234, "PostCode2": "12-345"},
		map[string]interface{}{"PrimaryEmail": "primary@example.com"},
		nil,
		nil,
	)
	want := "UPDATE test_structs SET post_code2=$1,price=$2 WHERE primary_email=$3"
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}

	got = h.GetQueryUpdate(
		map[string]interface{}{"Price": 1234, "PostCode2": "12-345", "FirstName": "Jane", "LastName": "Doe"},
		map[string]interface{}{"PrimaryEmail": "primary@example.com", "PostCode": "11-111"},
		map[string]bool{"FirstName": true, "LastName": true},
		map[string]bool{"PostCode": true},
	)
	want = "UPDATE test_structs SET first_name=$1,last_name=$2 WHERE post_code=$3"
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestPluralName(t *testing.T) {
	type Category struct{}
	type Cross struct{}
	type ProductCategory struct{}
	type UserCart struct{}

	h1 := NewStructSQL(&Category{}, StructSQLOptions{})
	h2 := NewStructSQL(&Cross{}, StructSQLOptions{})
	h3 := NewStructSQL(&ProductCategory{}, StructSQLOptions{})
	h4 := NewStructSQL(&UserCart{}, StructSQLOptions{})

	got := h1.GetQueryDropTable()
	want := "DROP TABLE IF EXISTS categories"
	if got != want {
		t.Fatalf("Want %v, got %v", want, got)
	}

	got = h2.GetQueryDropTable()
	want = "DROP TABLE IF EXISTS crosses"
	if got != want {
		t.Fatalf("Want %v, got %v", want, got)
	}

	got = h3.GetQueryDropTable()
	want = "DROP TABLE IF EXISTS product_categories"
	if got != want {
		t.Fatalf("Want %v, got %v", want, got)
	}

	got = h4.GetQueryDropTable()
	want = "DROP TABLE IF EXISTS user_carts"
	if got != want {
		t.Fatalf("Want %v, got %v", want, got)
	}
}

func TestSQLSelectQueriesWithJoin(t *testing.T) {
	h := NewStructSQL(&Product_WithDetails{}, StructSQLOptions{
		ForceName: "Product",
	})

	selectPrefix := "SELECT t1.product_id,t1.name,t1.price,t1.product_kind_id,t1.product_grp_id,t2.name,t3.code"
	selectPrefix += " FROM products t1 INNER JOIN product_kinds t2 ON t1.product_kind_id=t2.product_kind_id"
	selectPrefix += " INNER JOIN product_groups t3 ON t1.product_grp_id=t3.product_group_id"

	got := h.GetQuerySelectById()
	want := selectPrefix + " WHERE t1.product_id = $1"
	if got != want {
		t.Fatalf("Want %v, got %v", want, got)
	}

	got = h.GetQuerySelect([]string{"ProductKind_Name", "asc", "Name", "desc"}, 1, 3, map[string]interface{}{
		"Name":             "Product Name",
		"ProductGrp_Code":  "CODE1",
		"Price":            4400,
		"ProductKind_Name": "Kind 1",
		"_raw": []interface{}{
			".Price=? OR (.ProductGrp_Code IN (?)) OR .ProductKind_Name=?",
			// We do not really care about the values, the query contains $x only symbols
			// However, we need to pass either value or an array so that an array can be extracted into multiple $x's
			0,
			[]int{0, 0, 0, 0},
			0,
		},
		"_rawConjuction": RawConjuctionOR,
	}, nil, nil)
	want = selectPrefix + " WHERE"
	want += " (t1.name=$1 AND t1.price=$2 AND t3.code=$3 AND t2.name=$4) OR (t1.price=$5 OR (t3.code IN ($6,$7,$8,$9)) OR t2.name=$10)"
	want += " ORDER BY t2.name ASC,t1.name DESC LIMIT 1 OFFSET 3"
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}

	got = h.GetQuerySelectCount(map[string]interface{}{
		"Name":             "Product Name",
		"ProductGrp_Code":  "CODE1",
		"Price":            4400,
		"ProductKind_Name": "Kind 1",
		"_raw": []interface{}{
			".Price=? OR (.ProductGrp_Code IN (?)) OR .ProductKind_Name=?",
			// We do not really care about the values, the query contains $x only symbols
			// However, we need to pass either value or an array so that an array can be extracted into multiple $x's
			0,
			[]int{0, 0, 0, 0},
			0,
		},
		"_rawConjuction": RawConjuctionOR,
	}, nil)

	want = "SELECT COUNT(*) AS cnt"
	want += " FROM products t1 INNER JOIN product_kinds t2 ON t1.product_kind_id=t2.product_kind_id"
	want += " INNER JOIN product_groups t3 ON t1.product_grp_id=t3.product_group_id"
	want += " WHERE"
	want += " (t1.name=$1 AND t1.price=$2 AND t3.code=$3 AND t2.name=$4) OR (t1.price=$5 OR (t3.code IN ($6,$7,$8,$9)) OR t2.name=$10)"
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}

}
