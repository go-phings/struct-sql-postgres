# struct-sql-postgres

This package generates PostgreSQL SQL queries based on a struct instance. The concept is to define a struct, create a corresponding table to store its instances, and generate queries for managing the rows in that table, such as creating, updating, deleting, and selecting records.

The following queries can be generated:

* `CREATE TABLE`
* `DROP TABLE`
* `INSERT`
* `UPDATE ... WHERE id = ...`
* `INSERT ... ON CONFLICT UPDATE ...` (upsert)
* `SELECT ... WHERE id = ...`
* `DELETE ... WHERE id = ...`
* `SELECT ... WHERE ...`
* `DELETE ... WHERE ...`
* `UPDATE ... WHERE ...`


## How to use

### TL;DR

Check the code in `main_test.go` file that contains tests for all use cases.

### Defining a struct

Create a struct to define an object to be stored in a database table.  In the example below, let's create a `Product`.

As of now, a field called `ID` is required.  Also, the corresponding table column name that is generated for it is always prefixed with struct name. Hence, for `ID` field in `Product` that would be `product_id`.

There is an another field called `Flags` that could be added, and it treated the same (meaning `product_flags` is generated, instead of `flags`).

````go
type Product struct {
  ID int64
  Flags int64
  Name string
  Description string `2sql:"db_type:varchar(2000)"` // "db_type" is used to force a specific string type for a column
  Code string `2sql:"uniq"` // "uniq" tells the module to make this column uniq
  ProductionYear int
  CreatedByUserID int64
  LastModifiedByUserID int64
}
````

#### Field tags

In the above definition, a special tag `2sql` is used to add specific configuration for the column when creating a table.

| Tag key | Description |
|---|-----------|
| `uniq` | When passed, the column will get a `UNIQUE` constraint|
| `db_type` | Overwrites default `VARCHAR(255)` column type for string field. Possible values are: `TEXT`, `BPCHAR(X)`, `CHAR(X)`, `VARCHAR(X)`, `CHARACTER VARYING(X)`, `CHARACTER(X)` where `X` is the size. See [PostgreSQL character types](https://www.postgresql.org/docs/current/datatype-character.html) for more information. |

A different than `2sql` tag can be used by passing `TagName` in `StructSQLOptions{}` when calling `NewStructSQL` function (see below.)

### Create a controller for the struct

To generate an SQL query based on a struct, a `StructSQL` object is used.  One per struct.

````go

import (
  stsql "github.com/mikolajgs/prototyping/pkg/struct-sql-postgres"
)

(...)

s := stsql.NewStructSQL(Product{}, stsql.StructSQLOptions{})
````

#### StructSQLOptions

There are certain options that can be provided to the `NewStructSQL` function which change the functionally of the object.  See the table below.

| Key | Type | Description |
|---|---|---|
| DatabaseTablePrefix | `string` | Prefix for the table name, eg. `myprefix_` |
| ForceName | `string` | Table name is created out of the struct name, eg. for `MyProduct` that would be `my_products`. It is possible to overwrite the struct name, and further table name.  Hence, when `ForceName` is `AnotherStruct` then table name that will be used becomes `another_structs`. |
| TagName | `string` | Uses a different tag than `2sql`.  It is very useful when another module uses this module. |
| Base | `*StructSQL` | In some cases, we'd like to use already existing instance of this object as a base, instead of parsing the struct again.  For example, tags are already defined in another struct and should be re-used.  This is often used by other modules. |
| Joined | `map[string]*StructSQL` | When struct is used to describe a `SELECT` query with `INNER JOIN` to another structs (tables), this map can be used to overwrite `StructSQL` objects for children structs.  If not passed, then children structs that are meant to be used with `INNER JOIN` will be created using `NewStructSQL`.  See one of below sections on joined select queries for more details. |
| UseRootNameWhenJoinedPresent | bool | When struct is used to describe a `SELECT` query with `INNER JOIN` to another structs (tables) and the parent struct has a name like `Product_WithDetails` then it's so-called root name is `Product`, and that will be used as a base for the table name (so it'll be `products`). |

### Get SQL queries

Use any of the following `GetQuery*` commands to get a desired SQL query.  See examples below.

````go
drop := s.GetQueryDropTable() // retursn 'DROP TABLE IF EXISTS products'

create := s.GetQueryCreateTable() // returns 'CREATE TABLE products (...)'

updateById := s.GetQueryUpdateById() // returns 'UPDATE products SET product_flags = $1, name = $2 ... WHERE product_id = $8
````

### Get SQL queries with conditions

It is possible to generate queries such as `SELECT`, `DELETE` or `UPDATE` with conditions based on fields.  In the following examples below, all the condition (called "filters" in the code) are optional - there is no need to pass them.

The `_raw` (and `_rawConjuction`) is a special filter that allows passing a raw query.

#### SELECT

````go
// SELECT * FROM products WHERE (created_by_user_id=$1 AND name=$2) OR (product_year > $3
// AND product_year > $4 AND last_modified_by_user_id IN ($5,$6,$7,$8))
// ORDER BY production_year ASC, name ASC
// LIMIT 100 OFFSET 10
sqlSelect := s.GetQuerySelect(
  []string{"ProductionYear", "asc", "Name", "asc"},
  100, 10, 
  map[string]interface{
    "CreatedByUserID": 4,
    "Name": "Magic Sock",
    "_raw": []interface{}{
      ".ProductYear > ? AND .ProductYear < ? AND .LastModifiedByUserID(?)",
      // Below values are not important but the overall number of args match question marks
      0,
      0,
      []int{0,0,0,0}, // this list must contain same number of items as values
    },
    "_rawConjuction": stsql.RawConjuctionOR,
  }, nil, nil)
````

#### SELECT COUNT(*)

````
// Use GetQuerySelectCount without th first 3 arguments to get SELECT COUNT(*)
````

#### DELETE

````go
// DELETE FROM products WHERE (created_by_user_id=$1 AND name=$2) OR (product_year > $3
// AND product_year > $4 AND last_modified_by_user_id IN ($5,$6,$7,$8))
sqlDelete := s.GetQuerySelect(
  map[string]interface{
    "CreatedByUserID": 4,
    "Name": "Magic Sock",
    "_raw": []interface{}{
      ".ProductYear > ? AND .ProductYear < ? AND .LastModifiedByUserID(?)",
      // Below values are not important but the overall number of args match question marks
      0,
      0,
      []int{0,0,0,0}, // this list must contain same number of items as values
    },
    "_rawConjuction": stsql.RawConjuctionOR,
  }, nil, nil)
````

#### UPDATE

````go
// UPDATE products SET production_year=$1, last_modified_by_user_id=$2
// WHERE name LIKE $3;
sqlUpdate := s.GetQueryUpdate(
  map[string]interface{
    "ProductionYear": 1984,
    "LastModifiedByUserID": 13
  },
  map[string]interface{}{
    "_raw": ".Name LIKE ?",
    0, // One question mark, hence one additional value
  }, nil, nil)
````

### Get `SELECT` query with `INNER JOIN`

With `struct-sql-postgres` it is possible to build a query that would select data from multiple tables joined with `INNER JOIN`.


#### Creating a struct with joined struct

Suppose we need a query to select products with information on users that recently created and modified them.

````go
type Users struct {
  ID int64
  FirstName string
  LastName string
}

type Product_WithDetails struct {
  ID int64
  Flags int64
  Name string
  Description string `2sql:"db_type:varchar(2000)"` // "db_type" is used to force a specific string type for a column
  Code string `2sql:"uniq"` // "uniq" tells the module to make this column uniq
  ProductionYear int
  CreatedByUserID int64
  LastModifiedByUserID int64
  CreatedByUser *User `2sql:"join"`
  CreatedByUser_FirstName string
  CreatedByUser_LastName string
  LastModifiedByUser *User `2sql:"join"`
  LastModifiedByUser_FirstName string
  LastModifiedByUser_LastName string
}
````

#### Getting SELECT query for joined structs

An example query such as:

````SQL
SELECT t1.product_id,t1.product_flags,t1.name,t1.description,t1.code,t1.production_year,
t1.created_by_user_id,t1.last_modified_by_user_id,t2.first_name,t2.last_name,t3.first_name,t3.last_name
FROM products t1 INNER JOIN users t2 ON t1.created_by_user_id=t2.user_id
INNER JOIN users t3 ON t1.last_modified_by_user_id=t3.user_id
WHERE (t1.production_year=$1) AND (t2.first_name=$2 AND t3.first_name=$2)
ORDER BY t2.first_name ASC,t1.name DESC LIMIT 100 OFFSET 10
````

can be generated with the following code:

````go
got = h.GetQuerySelect([]string{"CreatedByUser_FirstName", "asc", "Name", "desc"}, 100, 10, map[string]interface{}{
  "ProductionYear": 1984,
  "_raw": []interface{}{
    ".CreatedByUser_FirstName=? AND .LastModifiedByUser_FirstName=?",
    // We do not really care about the values, the query contains $x only symbols
    // However, we need to pass either value or an array so that an array can be extracted into multiple $x's
    0,
    0,
  },
  "_rawConjuction": RawConjuctionAND,
}, nil, nil)
````
