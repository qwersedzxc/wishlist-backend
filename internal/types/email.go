package types

// FriendRequestData данные для уведомления о запросе в друзья
type FriendRequestData struct {
	RecipientName string
	SenderName    string
	SenderEmail   string
	AppURL        string
}

// BirthdayReminderData данные для напоминания о дне рождения
type BirthdayReminderData struct {
	RecipientName string
	FriendName    string
	BirthDate     string
	DaysLeft      int
	AppURL        string
}