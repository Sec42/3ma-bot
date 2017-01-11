package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"github.com/o3ma/o3"
)

var (
	threemaID o3.ThreemaID
	tr        o3.ThreemaRest
	sc        o3.SessionContext
	pass      = []byte{0x1, 0x2, 0x3, 0x4}
	idpath    = "threema.id"
	abpath    = "address.book"
	nickpath  = "nick.txt"
)

func cleanup() {
	//save the address book
	fmt.Printf("Saving addressbook to %s\n", abpath)
	err := sc.ID.Contacts.SaveTo(abpath)
	if err != nil {
		fmt.Println("saving addressbook failed")
		log.Fatalln(err)
	}
}

func main(){
	// Check if ID already exists
	if  _,err := os.Stat(idpath); os.IsNotExist(err) {
		threemaID,err = tr.CreateIdentity()
		if (err != nil){
			log.Fatalln(err)
			os.Exit(1)
		}

		err= threemaID.SaveToFile(idpath, pass)
		if (err != nil){
			log.Fatalln(err)
			os.Exit(1)
		}
	} else{
		threemaID,err = o3.LoadIDFromFile(idpath,pass)
		if (err != nil){
			log.Fatalln(err)
			os.Exit(1)
		}
	}

	fmt.Print("Starting bot: "+threemaID.String())
	if rawnick, err := ioutil.ReadFile(nickpath); err == nil{
		nick := strings.TrimSpace(string(rawnick))
		threemaID.Nick = o3.NewPubNick(nick)
		fmt.Printf("(%s)", nick)
	}
	fmt.Println("")
	fmt.Printf("QR-Code: 3mid:%s,%s\n",threemaID.String(),hex.EncodeToString(threemaID.GetPubKey()[:]))

	sc = o3.NewSessionContext(threemaID)

	//check if we can load an addressbook
	if _, err := os.Stat(abpath); !os.IsNotExist(err) {
		fmt.Printf("Loading addressbook from %s\n", abpath)
		err = sc.ID.Contacts.ImportFrom(abpath)
		if err != nil {
			fmt.Println("loading addressbook failed")
			log.Fatalln(err)
		}
	}

	// Make sure to save the addressbook on exit
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup()
		os.Exit(1)
	}()

	// let the session begin
	fmt.Println("Starting session")
	sendMsgChan, receiveMsgChan, err := sc.Run()
	if err != nil {
		log.Fatalln(err)
	}

	for receivedMessage := range receiveMsgChan {
		if receivedMessage.Err != nil {
			fmt.Printf("Error Receiving Message: %s\n", receivedMessage.Err)
			continue
		}
		switch msg := receivedMessage.Msg.(type) {
		case o3.TextMessage:
			var (
				cmdOut  []byte
				err     error
			)
			fmt.Printf("[%16x] <%s> (%s): %s\n", msg.ID(), msg.Sender(), msg.PubNick(), msg.Text())
			// Send read reciept
			rm, err := o3.NewDeliveryReceiptMessage(&sc, msg.Sender().String(), msg.ID(), o3.MSGREAD)
			if err != nil {
				log.Fatalln(err)
				os.Exit(1)
			}
			sendMsgChan <- rm
			fmt.Fprintf(os.Stderr, "[%16x] marked as read ([%16x])\n", msg.ID(),rm.ID())

			// Send off to bot
			words:= strings.Fields(msg.Text())
			if cmdOut, err = exec.Command("./utfe.bot", words...).Output(); err != nil {
				fmt.Fprintln(os.Stderr, "Oops: ", err,cmdOut)
				if(true){
					rm, err := o3.NewDeliveryReceiptMessage(&sc, msg.Sender().String(), msg.ID(), o3.MSGDISAPPROVED)
					if err != nil {
						log.Fatalln(err)
						os.Exit(1)
					}
					sendMsgChan <- rm
				}else{
					os.Exit(1)
				}
			}else{
				tm, err := o3.NewTextMessage(&sc, msg.Sender().String(), string(cmdOut[:]))
				if err != nil {
					log.Fatalln(err)
					os.Exit(1)
				}
				sendMsgChan <- tm
				fmt.Printf("[%16x] is reply to [%16x] ", tm.ID(),msg.ID())
				fmt.Printf("%s", string(cmdOut[:]))
			}

		case o3.DeliveryReceiptMessage:
			fmt.Printf("[%16x] ",msg.MsgID());
			if      (msg.Status()== o3.MSGDELIVERED){
				fmt.Printf("delivered to ");
			}else if(msg.Status()==o3.MSGREAD){
				fmt.Printf("read by      ");
			}else if(msg.Status()==o3.MSGAPPROVED){
				fmt.Printf("approved by  ");
			}else if(msg.Status()==o3.MSGDISAPPROVED){
				fmt.Printf("negative by  ");
			}else{
				fmt.Printf("<unk: %x> ",msg.Status());
			}
			fmt.Printf("<%s> (%s) ([%16x])\n", msg.Sender(),msg.PubNick(),msg.ID());

		case o3.TypingNotificationMessage:
			fmt.Printf("[%16x] <%s> (%s) ",msg.ID(),msg.Sender(),msg.PubNick());
			if (msg.OnOff==1){
				fmt.Printf("is typing\n");
			}else{
				fmt.Printf("is idle\n");
			}

		// Images / Audio: Save content for now, and don't react
		case o3.ImageMessage:
			plainPicture, err := msg.GetImageData(sc)
			if err != nil {
				log.Fatalln(err)
			}
			imageFile, err := ioutil.TempFile("tmp", "image."+msg.Sender().String()+".")
			if err != nil {
				log.Fatalln(err)
			}
			_, err = imageFile.Write(plainPicture)
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Printf("[%16x] <%s> (%s): Image[%d] %s\n", msg.ID(), msg.Sender(), msg.PubNick(), msg.Size, imageFile.Name())
//			os.Remove(imageFile.Name())

		case o3.AudioMessage:
			plainAudio, err := msg.GetAudioData(sc)
			if err != nil {
				log.Fatalln(err)
			}
			audioFile, err := ioutil.TempFile("tmp", "audio."+msg.Sender().String()+".")
			if err != nil {
				log.Fatalln(err)
			}
			_, err = audioFile.Write(plainAudio)
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Printf("[%16x] <%s> (%s): Audio[%d] %s\n", msg.ID(), msg.Sender(), msg.PubNick(), msg.Duration, audioFile.Name())
//			os.Remove(audioFile.Name())

		// Do not react to group messages
		case o3.GroupImageMessage:
			plainPicture, err := msg.GetImageData(sc)
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Printf("%s:\n%s\n", msg.Sender(), plainPicture)
		case o3.GroupTextMessage:
			fmt.Printf("%s for Group [%x] created by [%s]:\n%s\n", msg.Sender(), msg.GroupID(), msg.GroupCreator(), msg.Text())
		case o3.GroupManageSetNameMessage:
			fmt.Printf("Group [%x] is now called %s\n", msg.GroupID(), msg.Name())
		case o3.GroupManageSetMembersMessage:
			fmt.Printf("Group [%x] now includes %v\n", msg.GroupID() , msg.Members())
		case o3.GroupMemberLeftMessage:
			fmt.Printf("Member [%s] left the Group [%x]\n", msg.Sender(), msg.GroupID())
		default:
			fmt.Printf("Unknown message type from: %s\nContent: %#v" , msg.Sender(), msg)
		}
	}
}
