package maintenance

import "strings"

// StandardWorkOrderTypes is the single source of truth for fixed work-order categories stored in DB.
// Order defines CountByWorkOrderTypeSQL column order — repository CountByWorkOrderType Scan must stay aligned (then other).
// Do not reorder or trim without updating that Scan and handler expectations.
var StandardWorkOrderTypes = []string{
	"Preventive",
	"Corrective",
	"Emergency",
	"Inspection",
}

var (
	standardWorkOrderTypeSet map[string]struct{}
	countByWorkOrderTypeSQL  string
)

func init() {
	standardWorkOrderTypeSet = make(map[string]struct{}, len(StandardWorkOrderTypes))
	for _, t := range StandardWorkOrderTypes {
		standardWorkOrderTypeSet[t] = struct{}{}
	}
	countByWorkOrderTypeSQL = buildCountByWorkOrderTypeSQL()
}

func sqlStringLiteral(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

func buildCountByWorkOrderTypeSQL() string {
	if len(StandardWorkOrderTypes) == 0 {
		panic("maintenance: StandardWorkOrderTypes must be non-empty")
	}
	var b strings.Builder
	b.WriteString(`SELECT `)
	for i, t := range StandardWorkOrderTypes {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(`COUNT(CASE WHEN "workOrderType" = `)
		b.WriteString(sqlStringLiteral(t))
		b.WriteString(` THEN 1 END)`)
	}
	b.WriteString(`, COUNT(CASE WHEN "workOrderType" NOT IN (`)
	for i, t := range StandardWorkOrderTypes {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(sqlStringLiteral(t))
	}
	b.WriteString(`) THEN 1 END)
FROM maintenance 
WHERE "serialNumber" = $1`)
	return b.String()
}

// CountByWorkOrderTypeSQL returns the parameterized query for per-type counts (built from StandardWorkOrderTypes).
func CountByWorkOrderTypeSQL() string {
	return countByWorkOrderTypeSQL
}

// IsStandardWorkOrderType reports whether the stored type is one of the fixed categories (vs free-text / "Other").
func IsStandardWorkOrderType(workOrderType string) bool {
	_, ok := standardWorkOrderTypeSet[workOrderType]
	return ok
}

// SetWorkOrderTypeStandardFlag sets WorkOrderTypeIsStandard for API JSON (field is not persisted).
func (m *Maintenance) SetWorkOrderTypeStandardFlag() {
	if m == nil {
		return
	}
	m.WorkOrderTypeIsStandard = IsStandardWorkOrderType(m.WorkOrderType)
}
