package main

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"

	"github.com/tkanos/gonfig"
)

type configuration struct {
	AddressFrom string
	AddressTo   string
	Password    string
	SMTPServer  string
	SMTPPort    int
	Subject     string
}

func sendEmail(clientEmail string, comment string) {
	fmt.Println("sending email...")
	configuration := configuration{}
	confError := gonfig.GetConf("config.cfg", &configuration)
	if confError != nil {
		fmt.Println(confError)
	}

	// Connect to the SMTP Server
	servername := configuration.SMTPServer + ":" + strconv.Itoa(configuration.SMTPPort)

	host, _, _ := net.SplitHostPort(servername)
	// Set up authentication information.
	auth := smtp.PlainAuth("",
		configuration.AddressFrom,
		configuration.Password,
		host)

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	from := mail.Address{Name: configuration.AddressFrom, Address: configuration.AddressFrom}
	to := mail.Address{Name: "Рассылка", Address: configuration.AddressTo}
	title := configuration.Subject
	body := clientEmail + " оставил комментарий: \r\n" + comment + ".\r\n"

	header := make(map[string]string)
	header["From"] = from.Address
	header["To"] = to.Address
	header["Subject"] = encodeRFC2047(title)
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

	err := smtp.SendMail(configuration.SMTPServer+":"+strconv.Itoa(configuration.SMTPPort), auth, configuration.AddressTo, []string{to.Address}, []byte(message))
	if err != nil {
		log.Fatal(err)
	}

	c, err := smtp.Dial(servername)
	if err != nil {
		log.Panic(err)
	}

	c.StartTLS(tlsconfig)

	// Auth
	if err = c.Auth(auth); err != nil {
		log.Panic(err)
	}

	// To && From
	if err = c.Mail(from.Address); err != nil {
		log.Panic(err)
	}

	if err = c.Rcpt(to.Address); err != nil {
		log.Panic(err)
	}

	// Data
	w, err := c.Data()
	if err != nil {
		log.Panic(err)
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		log.Panic(err)
	}

	err = w.Close()
	if err != nil {
		log.Panic(err)
	}

	c.Quit()
}
func encodeRFC2047(String string) string {
	// use mail's rfc2047 to encode any string
	addr := mail.Address{String, ""}
	return strings.Trim(addr.String(), " <@>")
}
