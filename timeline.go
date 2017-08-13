package main

import (
	"fmt"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/kr/pretty"
	"sync"
	"time"
)

type TimelineSubscriber struct {
	cursor       time.Time
	chans        []chan *twitter.Tweet
	client       *twitter.Client
	followingIDs []int64
	reloadMutex  *sync.Mutex
}

func NewTimelineSubscriber(client *twitter.Client) *TimelineSubscriber {
	return &TimelineSubscriber{
		cursor:       time.Now(),
		chans:        []chan *twitter.Tweet{},
		client:       client,
		followingIDs: []int64{},
		reloadMutex:  &sync.Mutex{},
	}
}

func (ts *TimelineSubscriber) SubscribeToNew(channel chan *twitter.Tweet) {
	ts.chans = append(ts.chans, channel)
}

func (ts *TimelineSubscriber) publish(tweet *twitter.Tweet) {
	for _, channel := range ts.chans {
		channel <- tweet
	}
}

func (ts *TimelineSubscriber) tick() {
	// List tweets from "reploy"
	// Home Timeline
	homeTimelineParams := &twitter.HomeTimelineParams{
		Count:     25,
		TweetMode: "extended",
	}

	tweets, resp, err := ts.client.Timelines.HomeTimeline(homeTimelineParams)
	if err != nil {
		fmt.Println("There was error getting the latest tweets:", err)
		pretty.Println(resp.StatusCode)
		return
	}

	for i := 0; i < len(tweets); i++ {
		tweet := tweets[i]
		createdAt, err := tweet.CreatedAtTime()
		if err != nil {
			fmt.Println("Could read created at for:", tweet.User.ScreenName, tweet.ID, tweet.CreatedAt)
			continue
		}

		// Skip if the tweet is older than the current cursor.
		if ts.cursor.After(createdAt) {
			continue
		}

		// Skip if the tweet is does not match what we're looking for.
		if !ts.shouldReplyToTweet(&tweet) {
			continue
		}

		// Publish!!!
		ts.publish(&tweet)

		// Now we set the max tweet createdAt to be the new cursor. We could set a value for the "tick" but
		// that leaves a gap. This is safer.
		if ts.cursor.Before(createdAt) {
			// We need to add 1 ns because of how Before and After work.
			ts.cursor = createdAt.Add(1)
		}
	}
}

func (ts *TimelineSubscriber) Start() {
	go func() {
		for {
			fmt.Println("Main Loop Tick...")
			ts.tick()

			// The rate limit is 15 times per 15 minutes (once per minute). Let's go once per 2 minutes.
			time.Sleep(time.Second * 120)
		}
	}()

	go func() {
		for {
			// The rate limit is 15 times per 15 minutes (once per minute). Let's go once per 5 minutes.
			// NOTE: We sleep here first because the main func already primed the following cache.
			time.Sleep(time.Minute * 5)

			fmt.Println("Reload Tick...")
			ts.ReloadFollowingIDs()
		}
	}()
}

func (ts *TimelineSubscriber) isAUserWeFollow(userID int64) bool {
	ts.reloadMutex.Lock()
	defer ts.reloadMutex.Unlock()

	for _, id := range ts.followingIDs {
		if id == userID {
			return true
		}
	}

	return false
}

func (ts *TimelineSubscriber) shouldReplyToTweet(tweet *twitter.Tweet) bool {
	// Only reply when:
	// 1. user is on the list
	// 2. is a root tweet
	// 3. not retweeted

	return ts.isAUserWeFollow(tweet.User.ID) &&
		len(tweet.InReplyToScreenName) == 0 &&
		tweet.InReplyToUserID == 0 &&
		tweet.InReplyToStatusID == 0 &&
		!tweet.Retweeted &&
		tweet.RetweetedStatus == nil
}

func (ts *TimelineSubscriber) ReloadFollowingIDs() error {
	followingIDs, _, err := ts.client.Friends.IDs(&twitter.FriendIDParams{Count: 100})
	if err != nil {
		fmt.Println("Error attempting to reload the following list:", err)
		return err
	}

	ts.reloadMutex.Lock()
	defer ts.reloadMutex.Unlock()

	ts.followingIDs = followingIDs.IDs

	fmt.Println("Reloaded the following ID list. New size:", len(ts.followingIDs))

	return nil
}
