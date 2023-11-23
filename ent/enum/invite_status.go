package enum

type InviteStatus string

const (
	New      InviteStatus = "NEW"
	Sent     InviteStatus = "SENT"
	Shown    InviteStatus = "SHOWN"
	Accepted InviteStatus = "ACCEPTED"
	Declined InviteStatus = "DECLINED"
	Canceled InviteStatus = "CANCELED"
)

func inviteStatusValues() []InviteStatus {
	return []InviteStatus{New, Sent, Shown, Accepted, Declined, Canceled}
}

func (InviteStatus) Values() (kinds []string) {
	for _, value := range inviteStatusValues() {
		kinds = append(kinds, string(value))
	}
	return
}

func (m InviteStatus) Value() string {
	return string(m)
}

func (m InviteStatus) IsValid() bool {
	for _, value := range inviteStatusValues() {
		if m == value {
			return true
		}
	}
	return false
}
