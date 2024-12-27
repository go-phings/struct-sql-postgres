package structsqlpostgres

// StructSQL reflects the object to generate and cache PostgreSQL queries (CREATE TABLE, INSERT, UPDATE etc.).
// Database table and column names are lowercase with underscore and they are generated from field names.
// StructSQL is created within Controller and there is no need to instantiate it
type StructSQL struct {
	queryDropTable              string
	queryCreateTable            string
	queryInsert                 string
	queryUpdateById             string
	queryInsertOnConflictUpdate string
	querySelectById             string
	queryDeleteById             string
	querySelectPrefix           string
	querySelectCountPrefix      string
	queryDeletePrefix           string
	queryUpdatePrefix           string

	dbTbl       string
	dbColPrefix string
	dbFieldCols map[string]string
	dbCols      map[string]string
	url         string
	fields      []string

	fieldsDefaultValue  map[string]string
	fieldsUniq          map[string]bool
	fieldsTags          map[string]map[string]string
	fieldsOverwriteType map[string]string
	fieldsNotString     map[string]bool

	flags int

	baseFieldsTags   map[string]map[string]string
	joinedFieldsTags map[string]map[string]map[string]string

	hasJoined bool
	joined    map[string]*StructSQL

	hasModificationFields bool

	err error

	tagName string
}

const RawConjuctionOR = 1
const RawConjuctionAND = 2

const ValueEqual = 1
const ValueNotEqual = 2
const ValueLike = 3
const ValueMatch = 4
const ValueGreater = 5
const ValueLower = 6
const ValueGreaterOrEqual = 7
const ValueLowerOrEqual = 8
const ValueBit = 9

type StructSQLOptions struct {
	DatabaseTablePrefix string
	ForceName           string
	TagName             string
	Joined              map[string]*StructSQL
	// In some cases, we might want to copy over tags from already existing StructSQL instance. Such is called Base in here.
	Base *StructSQL
	// When struct has a name like 'xx_yy' and it has joined structs, use 'xx' as a name for table and column names
	UseRootNameWhenJoinedPresent bool
}

// NewStructSQL takes object and database table name prefix as arguments and returns StructSQL instance.
func NewStructSQL(obj interface{}, options StructSQLOptions) *StructSQL {
	h := &StructSQL{}

	h.tagName = "2sql"
	if options.TagName != "" {
		h.tagName = options.TagName
	}

	// Get field tags from Base StructSQL to use them instead of the ones parsed out from obj
	h.setBaseTags(options.Base)

	// If there are any joined structs (fields that are pointers to another structs and have the 'join' tag),
	// then each of them require an StructSQL instance as well. Instead of getting new one, they can be passed
	// straight away within the 'Joined' option.
	if options.Joined != nil {
		h.joined = options.Joined
	} else {
		h.joined = make(map[string]*StructSQL)
	}
	h.setJoinedTags()

	h.reflectStruct(obj, options.DatabaseTablePrefix, options.ForceName, options.UseRootNameWhenJoinedPresent)
	return h
}

// Err returns error that occurred when reflecting struct
func (h *StructSQL) Err() error {
	return h.err
}

// GetFlags returns StructSQL flags.
func (h *StructSQL) GetFlags() int {
	return h.flags
}

// GetQueryDropTable returns a DROP TABLE query.
func (h StructSQL) GetQueryDropTable() string {
	// When a struct contains fields that are pointers to other structs and these are meant to be joined, only SELECT can be generated
	if h.hasJoined {
		return ""
	}
	return h.queryDropTable
}

// GetQueryCreateTable return a CREATE TABLE query.
// Columns in the query are ordered the same way as they are defined in the struct, eg. SELECT field1_column, field2_column, ... etc.
func (h StructSQL) GetQueryCreateTable() string {
	if h.hasJoined {
		return ""
	}

	return h.queryCreateTable
}

// GetQueryInsert returns an INSERT query.
// Columns in the INSERT query are ordered the same way as they are defined in the struct, eg. SELECT field1_column, field2_column, ... etc.
func (h *StructSQL) GetQueryInsert() string {
	if h.hasJoined {
		return ""
	}

	return h.queryInsert
}

// GetQueryUpdateById returns an UPDATE query with WHERE condition on ID field.
// Columns in the UPDATE query are ordered the same way as they are defined in the struct, eg. SELECT field1_column, field2_column, ... etc.
func (h *StructSQL) GetQueryUpdateById() string {
	if h.hasJoined {
		return ""
	}

	return h.queryUpdateById
}

// GetQueryInsertOnConflictUpdate returns an "upsert" query, which will INSERT data when it does not exist or UPDATE it otherwise.
// Columns in the query are ordered the same way as they are defined in the struct, eg. SELECT field1_column, field2_column, ... etc.
func (h *StructSQL) GetQueryInsertOnConflictUpdate() string {
	if h.hasJoined {
		return ""
	}

	return h.queryInsertOnConflictUpdate
}

// GetQuerySelectById returns a SELECT query with WHERE condition on ID field.
// Columns in the SELECT query are ordered the same way as they are defined in the struct, eg. SELECT field1_column, field2_column, ... etc.
func (h *StructSQL) GetQuerySelectById() string {
	return h.querySelectById
}

// GetQueryDeleteById returns a DELETE query with WHERE condition on ID field.
func (h *StructSQL) GetQueryDeleteById() string {
	if h.hasJoined {
		return ""
	}
	return h.queryDeleteById
}

// GetQuerySelect returns a SELECT query with WHERE condition built from 'filters' (field-value pairs).
// Struct fields in 'filters' argument are sorted alphabetically. Hence, when used with database connection, their values (or pointers to it) must be sorted as well.
// Columns in the SELECT query are ordered the same way as they are defined in the struct, eg. SELECT field1_column, field2_column, ... etc.
func (h *StructSQL) GetQuerySelect(order []string, limit int, offset int, filters map[string]interface{}, orderFieldsToInclude map[string]bool, filterFieldsToInclude map[string]bool) string {
	s := h.querySelectPrefix

	qOrder := h.getQueryOrder(order, orderFieldsToInclude)
	qLimitOffset := h.getQueryLimitOffset(limit, offset)
	qWhere := h.getQueryFilters(filters, filterFieldsToInclude, 1)

	if qWhere != "" {
		s += " WHERE " + qWhere
	}
	if qOrder != "" {
		s += " ORDER BY " + qOrder
	}
	if qLimitOffset != "" {
		s += " " + qLimitOffset
	}
	return s
}

// GetQuerySelectCount returns a SELECT COUNT(*) query to count rows with WHERE condition built from 'filters' (field-value pairs).
// Struct fields in 'filters' argument are sorted alphabetically. Hence, when used with database connection, their values (or pointers to it) must be sorted as well.
func (h *StructSQL) GetQuerySelectCount(filters map[string]interface{}, filterFieldsToInclude map[string]bool) string {
	s := h.querySelectCountPrefix
	qWhere := h.getQueryFilters(filters, filterFieldsToInclude, 1)
	if qWhere != "" {
		s += " WHERE " + qWhere
	}
	return s
}

// GetQueryDelete return a DELETE query with WHERE condition built from 'filters' (field-value pairs).
// Struct fields in 'filters' argument are sorted alphabetically. Hence, when used with database connection, their values (or pointers to it) must be sorted as well.
func (h *StructSQL) GetQueryDelete(filters map[string]interface{}, filterFieldsToInclude map[string]bool) string {
	if h.hasJoined {
		return ""
	}

	s := h.queryDeletePrefix
	qWhere := h.getQueryFilters(filters, filterFieldsToInclude, 1)
	if qWhere != "" {
		s += " WHERE " + qWhere
	}
	return s
}

// GetQueryDelete return a DELETE query with WHERE condition built from 'filters' (field-value pairs) with RETURNING id.
// Struct fields in 'filters' argument are sorted alphabetically. Hence, when used with database connection, their values (or pointers to it) must be sorted as well.
func (h *StructSQL) GetQueryDeleteReturningID(filters map[string]interface{}, filterFieldsToInclude map[string]bool) string {
	if h.hasJoined {
		return ""
	}

	s := h.queryDeletePrefix
	qWhere := h.getQueryFilters(filters, filterFieldsToInclude, 1)
	if qWhere != "" {
		s += " WHERE " + qWhere
	}
	s += " RETURNING " + h.dbFieldCols["ID"]
	return s
}

// GetQueryUpdate returns an UPDATE query where specified struct fields (columns) are updated and rows match specific WHERE condition built from 'filters' (field-value pairs).
// Struct fields in 'values' and 'filters' arguments, are sorted alphabetically. Hence, when used with database connection, their values (or pointers to it) must be sorted as well.
func (h *StructSQL) GetQueryUpdate(values map[string]interface{}, filters map[string]interface{}, valueFieldsToInclude map[string]bool, filterFieldsToInclude map[string]bool) string {
	if h.hasJoined {
		return ""
	}

	s := h.queryUpdatePrefix

	qSet, lastVarNumber := h.getQuerySet(values, valueFieldsToInclude)
	s += " " + qSet

	qWhere := h.getQueryFilters(filters, filterFieldsToInclude, lastVarNumber+1)
	if qWhere != "" {
		s += " WHERE " + qWhere
	}

	return s
}

// GetFieldNameFromDBCol returns field name from a table column.
func (h *StructSQL) GetFieldNameFromDBCol(n string) string {
	return h.dbCols[n]
}

// HasModificationFields returns true if struct has all of the following int64 fields: CreatedAt, CreatedBy, LastModifiedAt, LastModifiedBy
func (h *StructSQL) HasModificationFields() bool {
	return h.hasModificationFields
}
