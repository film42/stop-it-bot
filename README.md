Stop It (bot)
=============

![](stopit.gif)

#### About

This is a bot that can be used to reply to people on twitter with an MJ meme saying, "Stop it! Get some help!"

#### How to Use

Create a twitter app. Add a user to the app. You'll need the following env variables set:

```
export CONSUMER_KEY=XXXXXXXXXXXXXXXXXXXXXXXXXXXX
export CONSUMER_SECRET=XXXXXXXXXXXXXXXXXXXXXXXXXXXX
export ACCESS_TOKEN=XXXXXXXXXXXXXXXXXXXXXXXXXXXX
export ACCESS_TOKEN_SECRET=XXXXXXXXXXXXXXXXXXXXXXXXXXXX
```

Now that they're set you can build this app and run it.

You add people to the list by following them. The timeline subscriber will refresh your timeline every 2 minutes. The
following list will refresh every 5 minutes. The twitter API rate-limit for the endpoints used by this bot are 15 calls
per 15 min, so although we could speed up our intervals, it doesn't actually make much sense.

#### License

MIT License
