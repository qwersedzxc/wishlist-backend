package email

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"net/smtp"
	"strings"

	"github.com/qwersedzxc/wishlist-backend/internal/config"
	"github.com/qwersedzxc/wishlist-backend/internal/types"
)

// Service предоставляет функции для отправки email
type Service struct {
	cfg config.MultiSMTPCfg
	log *slog.Logger
}

// New создает новый email сервис
func New(cfg config.MultiSMTPCfg, log *slog.Logger) *Service {
	return &Service{
		cfg: cfg,
		log: log,
	}
}

// SendFriendRequest отправляет уведомление о запросе в друзья
func (s *Service) SendFriendRequest(to string, data types.FriendRequestData) error {
	subject := "Новый запрос в друзья в Wishlist App"
	
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Запрос в друзья</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #2c3e50;">👋 Новый запрос в друзья!</h2>
        
        <p>Привет, {{.RecipientName}}!</p>
        
        <p><strong>{{.SenderName}}</strong> ({{.SenderEmail}}) отправил вам запрос в друзья в Wishlist App.</p>
        
        <div style="background: #f8f9fa; padding: 15px; border-radius: 5px; margin: 20px 0;">
            <p style="margin: 0;">Теперь вы сможете:</p>
            <ul style="margin: 10px 0;">
                <li>Просматривать вишлисты друг друга</li>
                <li>Бронировать подарки</li>
                <li>Получать напоминания о днях рождения</li>
            </ul>
        </div>
        
        <p>
            <a href="{{.AppURL}}" 
               style="background: #007bff; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; display: inline-block;">
                Принять запрос
            </a>
        </p>
        
        <hr style="margin: 30px 0; border: none; border-top: 1px solid #eee;">
        <p style="color: #666; font-size: 14px;">
            Это письмо отправлено автоматически из Wishlist App.<br>
            Если вы не ожидали это письмо, просто проигнорируйте его.
        </p>
    </div>
</body>
</html>`

	return s.sendEmail(to, subject, tmpl, data)
}

// SendBirthdayReminder отправляет напоминание о дне рождения друга
func (s *Service) SendBirthdayReminder(to string, data types.BirthdayReminderData) error {
	var subject string
	if data.DaysLeft == 0 {
		subject = fmt.Sprintf("🎉 Сегодня день рождения у %s!", data.FriendName)
	} else {
		subject = fmt.Sprintf("🎂 Скоро день рождения у %s (%d дн.)", data.FriendName, data.DaysLeft)
	}
	
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Напоминание о дне рождения</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        {{if eq .DaysLeft 0}}
            <h2 style="color: #e74c3c;">🎉 Сегодня день рождения!</h2>
            <p>Привет, {{.RecipientName}}!</p>
            <p><strong>Сегодня день рождения у {{.FriendName}}!</strong></p>
        {{else}}
            <h2 style="color: #f39c12;">🎂 Скоро день рождения</h2>
            <p>Привет, {{.RecipientName}}!</p>
            <p>Через <strong>{{.DaysLeft}} {{if eq .DaysLeft 1}}день{{else if le .DaysLeft 4}}дня{{else}}дней{{end}}</strong> день рождения у <strong>{{.FriendName}}</strong> ({{.BirthDate}}).</p>
        {{end}}
        
        <div style="background: #fff3cd; padding: 15px; border-radius: 5px; margin: 20px 0; border-left: 4px solid #ffc107;">
            <p style="margin: 0;">💡 <strong>Не забудьте:</strong></p>
            <ul style="margin: 10px 0;">
                <li>Посмотреть вишлист {{.FriendName}}</li>
                <li>Забронировать подарок</li>
                <li>Поздравить с днем рождения!</li>
            </ul>
        </div>
        
        <p>
            <a href="{{.AppURL}}" 
               style="background: #28a745; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; display: inline-block;">
                Открыть Wishlist App
            </a>
        </p>
        
        <hr style="margin: 30px 0; border: none; border-top: 1px solid #eee;">
        <p style="color: #666; font-size: 14px;">
            Это письмо отправлено автоматически из Wishlist App.<br>
            Вы можете отключить уведомления в настройках профиля.
        </p>
    </div>
</body>
</html>`

	return s.sendEmail(to, subject, tmpl, data)
}

// selectSMTPConfig выбирает подходящий SMTP сервер на основе email получателя
func (s *Service) selectSMTPConfig(recipientEmail string) config.SMTPCfg {
	domain := s.extractDomain(recipientEmail)
	
	// Выбираем SMTP сервер в зависимости от домена получателя
	switch domain {
	case "yandex.ru", "yandex.com", "ya.ru":
		if s.cfg.Primary.Host != "" {
			s.log.Info("using primary SMTP for yandex domain", "domain", domain)
			return s.cfg.Primary
		}
	case "gmail.com", "googlemail.com":
		if s.cfg.Secondary.Host != "" {
			s.log.Info("using secondary SMTP for gmail domain", "domain", domain)
			return s.cfg.Secondary
		}
	case "mail.ru", "inbox.ru", "list.ru", "bk.ru":
		if s.cfg.Tertiary.Host != "" {
			s.log.Info("using tertiary SMTP for mail.ru domain", "domain", domain)
			return s.cfg.Tertiary
		}
	}
	
	// Fallback: используем первый доступный SMTP
	if s.cfg.Primary.Host != "" {
		s.log.Info("using primary SMTP as fallback", "recipientDomain", domain)
		return s.cfg.Primary
	}
	if s.cfg.Secondary.Host != "" {
		s.log.Info("using secondary SMTP as fallback", "recipientDomain", domain)
		return s.cfg.Secondary
	}
	if s.cfg.Tertiary.Host != "" {
		s.log.Info("using tertiary SMTP as fallback", "recipientDomain", domain)
		return s.cfg.Tertiary
	}
	
	// Возвращаем пустую конфигурацию если ничего не настроено
	return config.SMTPCfg{}
}

// extractDomain извлекает домен из email адреса
func (s *Service) extractDomain(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}
	return strings.ToLower(parts[1])
}

// sendEmail отправляет email с HTML шаблоном
func (s *Service) sendEmail(to, subject, htmlTemplate string, data interface{}) error {
	// Выбираем подходящий SMTP сервер
	smtpCfg := s.selectSMTPConfig(to)
	
	if smtpCfg.Host == "" || smtpCfg.Username == "" {
		s.log.Warn("SMTP not configured, skipping email", "to", to, "subject", subject)
		return nil
	}

	// Парсим шаблон
	tmpl, err := template.New("email").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Рендерим шаблон
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Формируем сообщение
	msg := fmt.Sprintf("To: %s\r\n"+
		"From: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", to, smtpCfg.From, subject, body.String())

	// Настраиваем аутентификацию
	auth := smtp.PlainAuth("", smtpCfg.Username, smtpCfg.Password, smtpCfg.Host)

	// Отправляем email
	addr := fmt.Sprintf("%s:%d", smtpCfg.Host, smtpCfg.Port)
	err = smtp.SendMail(addr, auth, smtpCfg.From, []string{to}, []byte(msg))
	if err != nil {
		s.log.Error("failed to send email", "error", err, "to", to, "subject", subject, "smtp", smtpCfg.Host)
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.log.Info("email sent successfully", "to", to, "subject", subject, "smtp", smtpCfg.Host, "from", smtpCfg.From)
	return nil
}
