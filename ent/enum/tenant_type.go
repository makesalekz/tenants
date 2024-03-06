package enum

type TenantType string

const (
	Personal TenantType = "PERSONAL"
	Business TenantType = "BUSINESS"
)

var tenantTypes = map[TenantType]struct{}{Personal: {}, Business: {}}

func (TenantType) Values() (kinds []string) {
	for _, value := range inviteStatusValues() {
		kinds = append(kinds, string(value))
	}
	return
}

func (m TenantType) Value() string {
	return string(m)
}

func (m TenantType) IsValid() bool {
	if _, ok := tenantTypes[m]; ok {
		return true
	}

	return false
}

func (m TenantType) DefaultIfInvalid() TenantType {
	if m.IsValid() {
		return m
	}
	return Personal
}
