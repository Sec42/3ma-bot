# 3ma-bot

## This is a simple threema bot frame based on the [o3](http://github.com/o3ma) library

It is written in go. If this is your first time working with go, you need to setup `$GOPATH` similar to this:

```
export GOPATH=~/.go
mkdir $GOPATH
go get github.com/o3ma/o3rest
go get github.com/o3ma/o3
```

### Usage

Build and run with `go run simple-bot.go`.

It will create an threema ID file on first run which will be saved to `threema.id`. I suggest add this to your backup and do not publish it.

The addressbook of the people that communicate with your bot will be saved to `address.book` if you kill the bot.

All incoming text messages will be passed on to a binary `utfe.bot`, and the output will be sent back to the originator.

This `utfe.bot` binary/script is not part of this repo and can be written in any language. Use your own imagination.

### Authentication

On startup, the bot prints the string necessary to create the QR code used for authentication in the threema mobile app.

To create the QR code you can use: 
```
qrencode -t ANSI 3mid:....
```

### Licence

All code in this repo is herby licenced under the 2-clause BSD licence. 

### Thanks

Thanks to [@NerdingByDoing](https://twitter.com/NerdingByDoing) and [@twillnix](https://twitter.com/twillnix) for their [talk](https://media.ccc.de/v/33c3-8062-a_look_into_the_mobile_messaging_black_box) and their personal late-night support at [33c3](https://events.ccc.de/congress/2016/wiki/Main_Page).
