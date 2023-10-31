package email

import (
	"fmt"
	"html/template"
	"log"
	"net/smtp"

	"github.com/bb-consent/api/internal/config"
	"github.com/microcosm-cc/bluemonday"
)

// SMTPConfig Smtp configuration
var SMTPConfig config.SmtpConfig

var auth smtp.Auth

// Init Initialize the Smtp configuration
func Init(config *config.Configuration) {
	SMTPConfig = config.Smtp
}

// SendWelcomeEmail Send welcome email to user
func SendWelcomeEmail(username string, firstname string, subject string, body string, from string) {
	auth = smtp.PlainAuth("", SMTPConfig.Username, SMTPConfig.Password, SMTPConfig.Host)

	r := NewRequest([]string{username}, subject, body, from)
	escapedFirstName := template.HTMLEscapeString(firstname)

	emailTemplateString := `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN"
        "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html>
<head>
    <link href="https://fonts.googleapis.com/css?family=Roboto:100,300,400,500" rel="stylesheet">
    <meta name="viewport" content="width=device-width"/>
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8"/>
    <style>
        html, body {
            font-family: "Roboto", serif;
        }

        h1 {
            color: #eee;
        }
    </style>
</head>
<body>

<table width="100%" border="0" cellspacing="0" cellpadding="0">
    <tr>
        <td align="center" style="background-color: #fff;">

            <!-- logo -->
            <table>
                <tr>
                    <td>
                        <img class="logo" style="width:300px;height:100px;border-radius:20px;padding:16px;" id="logo"
                             src="https://storage.googleapis.com/igrant-api-images/email-images/iGrant_1900_600_Blue.png">
                    </td>
                </tr>
            </table>


            <!-- content -->
            <table style="width: 100%;" border="0" cellspacing="0" cellpadding="0">
                <tr>
                    <td></td>
                    <td width="600" style="font-size: 16px;">
                        <p style="font-weight: bold;font-size: 16px;color: #000;">Hi ` + escapedFirstName + `,</p>
                        <div style="color:#8c8a8a">
                            <p>
                                We are delighted that you are now registered to iGrant.io. Please check
                                <a href="https://privacy.igrant.io">privacy.igrant.io</a> page to view what
                                information we have about you, mark your consents and exercise your rights to your data.
                                Our aim at iGrant.io is to enable trust and transparency in a data sharing economy.
                                Empowering users to exercise
                                choice and control over their data will boost confidence in their relationship with the
                                organizations that manage
                                their data. In turn, compliant access to consumer data will improve transactions and
                                business processes, as well as
                                enhance consumer loyalty.
                            </p>
                            <p>
                                Please find below our recent whitepapers:
                            </p>
                            <p>
                                DPO WhitePaper:
                                <a href="https://igrant.io/papers/iGrant.io_Managing_Consent_in_a_Data_Sharing_Economy.pdf">https://igrant.io/papers/iGrant.io_Managing_Consent_in_a_Data_Sharing_Economy.pdf</a>
                            </p>
                            <p>CMO WhitePaper: <a href="https://igrant.io/papers/iGrant.io_ExecutiveBrief_v8.4.pdf">https://igrant.io/papers/iGrant.io_ExecutiveBrief_v8.4.pdf</a>
                            </p>
                            <p>Consumer Paper: <a
                                        href="https://igrant.io/papers/iGrant.io_SimplifiedConsumerGuide_to_GDPR.pdf">https://igrant.io/papers/iGrant.io_SimplifiedConsumerGuide_to_GDPR.pdf</a>
                            </p>
                            <p>
                                We hope you enjoy the journey to making managing consents and data sharing more
                                transparent and trustworthy!
                            </p>

                            <p>
                                If you wish to have a demo of iGrant.io platform, kindly reply this mail and we fix that
                                for you.
                            </p>


                            <div>
                                <div style="color: #000;line-height: 4px;">
                                    <p>Best Regards,</p>
                                    <p>The iGrant.io Team</p>
                                </div>
                            </div>
                        </div>
                    </td>
                    <td></td>
                </tr>
            </table>

        </td>
    </tr>
</table>

</body>

</html>`

	_, err := r.SendEmail(emailTemplateString)

	if err != nil {
		// Sending email failed
		log.Printf("Failed to send welcome email to username<%v> : %v", username, err)
		return
	}

}

// Request Request struct for constructing payload for sending email
type Request struct {
	from    string
	to      []string
	subject string
	body    string
}

// NewRequest For creating request object
func NewRequest(to []string, subject, body string, from string) *Request {
	return &Request{
		from:    from,
		to:      to,
		subject: subject,
		body:    body,
	}
}

// SendEmail For sending email
func (r *Request) SendEmail(body string) (bool, error) {

	p := bluemonday.UGCPolicy()

	p = p.AllowAttrs("border", "cellspacing", "cellpadding", "style").OnElements("table")
	p = p.AllowAttrs("align", "style").OnElements("td")
	p = p.AllowAttrs("style").Globally()
	p = p.AllowAttrs("class", "style", "id", "src").OnElements("img")
	p = p.AllowStyles("color", "width", "background-color", "height", "border-radius", "padding", "font-size", "font-weight", "line-height").Globally()

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subject := "Subject: " + r.subject + "!\n"
	msg := []byte(subject + mime + "\n" + body)

	// Sanitize the msg
	sanitizedMsg := p.Sanitize(string(msg))

	addr := fmt.Sprintf("%s:%d", SMTPConfig.Host, SMTPConfig.Port)

	if err := smtp.SendMail(addr, auth, r.from, r.to, []byte(sanitizedMsg)); err != nil {
		return false, err
	}
	return true, nil
}
