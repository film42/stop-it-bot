package main

import (
	"fmt"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"os"
)

var ConsumerKey = os.Getenv("CONSUMER_KEY")
var ConsumerSecret = os.Getenv("CONSUMER_SECRET")
var AccessToken = os.Getenv("ACCESS_TOKEN")
var AccessTokenSecret = os.Getenv("ACCESS_TOKEN_SECRET")

func configsPresent() bool {
	if len(ConsumerKey) == 0 {
		fmt.Println("You're missing a CONSUMER_KEY")
		return false
	}

	if len(ConsumerSecret) == 0 {
		fmt.Println("You're missing a CONSUMER_SECRET")
		return false
	}

	if len(AccessToken) == 0 {
		fmt.Println("You're missing a ACCESS_TOKEN")
		return false
	}

	if len(AccessTokenSecret) == 0 {
		fmt.Println("You're missing a ACCESS_TOKEN_SECRET")
		return false
	}

	return true
}

// A gif of MJ saying, "Stop it! Get some help!"
var StopItGifIds []int64 = []int64{896447314411368450}

func replyToTweet(client *twitter.Client, tweet *twitter.Tweet) error {
	status := fmt.Sprintf(
		"Hey, @%s! This is a friendly reminder to let you know that you are the problem.",
		tweet.User.ScreenName,
	)
	params := &twitter.StatusUpdateParams{
		InReplyToStatusID: tweet.ID,
		MediaIds:          StopItGifIds,
	}
	_, _, err := client.Statuses.Update(status, params)
	return err
}

func main() {
	if !configsPresent() {
		return
	}

	config := oauth1.NewConfig(ConsumerKey, ConsumerSecret)
	token := oauth1.NewToken(AccessToken, AccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	// Twitter client
	client := twitter.NewClient(httpClient)

	// Verify Credentials
	verifyParams := &twitter.AccountVerifyParams{
		SkipStatus:   twitter.Bool(true),
		IncludeEmail: twitter.Bool(true),
	}
	user, _, err := client.Accounts.VerifyCredentials(verifyParams)
	if err != nil {
		fmt.Println("Error verifying user! Exiting.")
		return
	}
	fmt.Println("Verified Creds For User:", user.ScreenName)

	timelineSubscriber := NewTimelineSubscriber(client)

	tweetChannel := make(chan *twitter.Tweet, 100)
	timelineSubscriber.SubscribeToNew(tweetChannel)
	err = timelineSubscriber.ReloadFollowingIDs()
	if err != nil {
		// This method logs its own error.
		return
	}

	// This will schedule and start required go-routines.
	timelineSubscriber.Start()

	for {
		select {
		case tweet := <-tweetChannel:
			err := replyToTweet(client, tweet)
			fmt.Println("Replying to:", tweet.User.ScreenName, tweet.ID, tweet.FullText)

			if err != nil {
				fmt.Println("Received an error while posting reply:", err)
			}
		}
	}
}
