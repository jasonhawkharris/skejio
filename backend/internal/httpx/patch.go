package httpx

import (
	"fmt"
	"strings"
)

// PatchBuilder accumulates SET clauses for a dynamic "UPDATE ... SET ..."
// query, as used by PATCH handlers that only touch fields present in the
// request body. Zero value is ready to use.
type PatchBuilder struct {
	setClauses []string
	args       []any
}

// Set records that column should be updated to value, assigning it the next
// placeholder in call order.
func (b *PatchBuilder) Set(column string, value any) {
	b.args = append(b.args, value)
	b.setClauses = append(b.setClauses, fmt.Sprintf("%s = $%d", column, len(b.args)))
}

// Empty reports whether no fields have been set yet - callers should reject
// the request with "no updatable fields provided" in that case.
func (b *PatchBuilder) Empty() bool {
	return len(b.setClauses) == 0
}

// NextArg returns the 1-based placeholder position the next-added argument
// would receive. Use it to number WHERE-clause placeholders that follow the
// SET clauses, e.g. fmt.Sprintf("id = $%d AND user_id = ANY($%d::uuid[])",
// b.NextArg(), b.NextArg()+1).
func (b *PatchBuilder) NextArg() int {
	return len(b.args) + 1
}

// Build returns the finished "UPDATE <table> SET <sets> WHERE <where>
// RETURNING <returning>" query and its full argument list (accumulated Set
// values followed by whereArgs). where's placeholders must start at NextArg()
// as it stood before this call.
func (b *PatchBuilder) Build(table, where, returning string, whereArgs ...any) (string, []any) {
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s RETURNING %s",
		table, strings.Join(b.setClauses, ", "), where, returning)
	return query, append(b.args, whereArgs...)
}
