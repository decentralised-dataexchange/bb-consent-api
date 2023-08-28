package notifications

import (
	"context"
	"fmt"
	"log"
	"strconv"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/org"
	"github.com/bb-consent/api/src/user"
	"google.golang.org/api/option"
)

var app *firebase.App

// Init Initializes the firebase admin SDK
func Init() error {
	var err error
	opt := option.WithCredentialsFile("/opt/igrant/api/config/jenkins-189019-firebase-adminsdk-p48w5-0e7b33da7b.json")
	app, err = firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
		return err
	}
	return nil
}

// SendEulaUpdateNotification Send purpose update notification to user
func SendEulaUpdateNotification(u user.User, o org.Organization) error {
	title := "Organization EULA updated"
	body := "Organization EULA updated with new EULA"

	var notification Notification
	notification.UserID = u.ID.Hex()
	notification.OrgID = o.ID.Hex()
	notification.Type = EulaUpdate
	notification.Title = title

	n, err := Add(notification)
	if err != nil {
		fmt.Printf("db add error")
		return err
	}

	count, err := GetUnReadCountByUserID(u.ID.Hex())
	if err != nil {
		fmt.Printf("db get count error")
		return err
	}

	fmt.Printf("sending Eula notification \n")
	return Send(u.Client.Token, u.Client.Type, title, body, notification.Type, n.ID.Hex(), count)
}

// SendDataBreachNotification Send Data breach notification to all users of organization.
func SendDataBreachNotification(dataBreachID string, u user.User, orgID string, orgName string) error {
	title := "Data Breach notification"
	body := "Data breach notification from " + orgName

	var notification Notification
	notification.UserID = u.ID.Hex()
	notification.OrgID = orgID
	notification.Type = DataBreach
	notification.Title = title
	notification.DataBreachID = dataBreachID

	n, err := Add(notification)
	if err != nil {
		fmt.Printf("db add error")
		return err
	}

	count, err := GetUnReadCountByUserID(u.ID.Hex())
	if err != nil {
		fmt.Printf("db get count error")
		return err
	}

	fmt.Printf("sending Data Breach notification \n")
	return Send(u.Client.Token, u.Client.Type, title, body, notification.Type, n.ID.Hex(), count)
}

// Send Sends notification to the user registered mobile device
func Send(registrationToken string, deviceType int, title string, body string, notificationType int, notificationID string, notificationUnReadCount int) error {
	// Obtain a messaging.Client from the App.
	log.Printf("token: %v ", registrationToken)
	ctx := context.Background()
	client, err := app.Messaging(ctx)
	if err != nil {
		log.Printf("Failed to create messaging client: %v", err)
		return err
	}

	var message *messaging.Message

	if deviceType == common.ClientTypeIos {
		// See documentation on defining a message payload.
		badge := notificationUnReadCount
		message = &messaging.Message{
			//		Data: map[string]string{
			//			"score": "850",
			//			"time":  "2:45",
			//		},
			/*Notification: &messaging.Notification{
				Title: "$GOOG up 1.43% on the day",
				Body:  "$GOOG gained 11.80 points to close at 835.67, up 1.43% on the day.",
			}, */
			APNS: &messaging.APNSConfig{
				Headers: map[string]string{
					"apns-priority": "10",
				},
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{
						Alert: &messaging.ApsAlert{
							Title: title,
							Body:  body,
						},
						Badge: &badge,
						CustomData: map[string]interface{}{
							"type": notificationType,
							"ID":   notificationID,
						},
					},
				},
			},
			Token: registrationToken,
		}
	} else {
		/*
			message = &messaging.Message{
				Android: &messaging.AndroidConfig{
					Priority: "normal",
					Notification: &messaging.AndroidNotification{
						Title: "$GOOG up 1.43% on the day",
						Body:  "$GOOG gained 11.80 points to close at 835.67, up 1.43% on the day.",
						Icon:  "stock_ticker_update",
						Color: "#f45342",
					},
				},
				Topic: "industry-tech",
			}
		*/
		message = &messaging.Message{
			Android: &messaging.AndroidConfig{
				Priority: "normal",
				Notification: &messaging.AndroidNotification{
					Title: title,
					Body:  body,
					Color: "#f45342",
				},
				Data: map[string]string{
					"type": strconv.Itoa(notificationType),
					"ID":   notificationID,
				},
			},
			Token: registrationToken,
		}
	}

	// Send a message to the device corresponding to the provided
	// registration token.
	//response, err := client.SendDryRun(ctx, message)
	response, err := client.Send(ctx, message)
	if err != nil {
		return err
	}
	// Response is a message ID string.
	fmt.Println("Successfully sent message:", response)
	return nil
}
