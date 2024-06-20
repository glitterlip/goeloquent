package goeloquent

import "github.com/glitterlip/goeloquent/query"

const (
	CONDITION_TYPE_BASIC          = query.CONDITION_TYPE_BASIC
	CONDITION_TYPE_COLUMN         = query.CONDITION_TYPE_COLUMN
	CONDITION_TYPE_RAW            = query.CONDITION_TYPE_RAW
	CONDITION_TYPE_IN             = query.CONDITION_TYPE_IN
	CONDITION_TYPE_NOT_IN         = query.CONDITION_TYPE_NOT_IN
	CONDITION_TYPE_NULL           = query.CONDITION_TYPE_NULL
	CONDITION_TYPE_BETWEEN        = query.CONDITION_TYPE_BETWEEN
	CONDITION_TYPE_BETWEEN_COLUMN = query.CONDITION_TYPE_BETWEEN_COLUMN
	CONDITION_TYPE_NOT_BETWEEN    = query.CONDITION_TYPE_NOT_BETWEEN
	CONDITION_TYPE_DATE           = query.CONDITION_TYPE_DATE
	CONDITION_TYPE_TIME           = query.CONDITION_TYPE_TIME
	CONDITION_TYPE_DATETIME       = query.CONDITION_TYPE_DATETIME
	CONDITION_TYPE_DAY            = query.CONDITION_TYPE_DAY
	CONDITION_TYPE_MONTH          = query.CONDITION_TYPE_MONTH
	CONDITION_TYPE_YEAR           = query.CONDITION_TYPE_YEAR
	CONDITION_TYPE_CLOSURE        = query.CONDITION_TYPE_CLOSURE //todo
	CONDITION_TYPE_NESTED         = query.CONDITION_TYPE_NESTED
	CONDITION_TYPE_SUB            = query.CONDITION_TYPE_SUB
	CONDITION_TYPE_EXIST          = query.CONDITION_TYPE_EXIST
	CONDITION_TYPE_NOT_EXIST      = query.CONDITION_TYPE_NOT_EXIST
	CONDITION_TYPE_ROW_VALUES     = query.CONDITION_TYPE_ROW_VALUES
	BOOLEAN_AND                   = query.BOOLEAN_AND
	BOOLEAN_OR                    = query.BOOLEAN_OR
	CONDITION_JOIN_NOT            = query.CONDITION_JOIN_NOT //todo
	JOIN_TYPE_LEFT                = query.JOIN_TYPE_LEFT
	JOIN_TYPE_RIGHT               = query.JOIN_TYPE_RIGHT
	JOIN_TYPE_INNER               = query.JOIN_TYPE_INNER
	JOIN_TYPE_CROSS               = query.JOIN_TYPE_CROSS
	JOIN_TYPE_LATERAL             = query.JOIN_TYPE_LATERAL
	ORDER_ASC                     = query.ORDER_ASC
	ORDER_DESC                    = query.ORDER_DESC
	TYPE_SELECT                   = query.TYPE_SELECT
	TYPE_FROM                     = query.TYPE_FROM
	TYPE_JOIN                     = query.TYPE_JOIN
	TYPE_WHERE                    = query.TYPE_WHERE
	TYPE_GROUP_BY                 = query.TYPE_GROUP_BY
	TYPE_HAVING                   = query.TYPE_HAVING
	TYPE_ORDER                    = query.TYPE_ORDER
	TYPE_UNION                    = query.TYPE_UNION
	TYPE_UNION_ORDER              = query.TYPE_UNION_ORDER
	TYPE_COLUMN                   = query.TYPE_COLUMN
	TYPE_AGGREGRATE               = query.TYPE_AGGREGRATE
	TYPE_OFFSET                   = query.TYPE_OFFSET
	TYPE_LIMIT                    = query.TYPE_LIMIT
	TYPE_LOCK                     = query.TYPE_LOCK
	TYPE_INSERT                   = query.TYPE_INSERT
	TYPE_UPDATE                   = query.TYPE_UPDATE
	CONDITION_TYPE_JSON_CONTAINS  = query.CONDITION_TYPE_JSON_CONTAINS
)
