package chatserver

import (
	"log"
	"math/rand"
	sync "sync"
	"time"
)

// handle messages in a server
type messageUnit struct {
	ClientName        string
	MessageBody       string
	MessageUniqueCode int
	ClientUniqueCode  int
}

// hold slices of messageUnit and client
type messageHandle struct {
	MQue []messageUnit
	mu   sync.Mutex // handle asynchronous read write operations
}

var messageHandleObject = messageHandle{}

type ChatServer struct {
}

// register each stream to a unique code for broadcasting purposes
type clients struct {
	streamlist map[int]Services_ChatServiceServer
}

// used to search the list of streams in client struct
var clientCodeList []int

var clientList clients

// initialise map
func init() {
	clientList = clients{
		streamlist: make(map[int]Services_ChatServiceServer),
	}
}

// define ChatService
func (is *ChatServer) ChatService(csi Services_ChatServiceServer) error {

	clientUniqueCode := rand.Intn(1e6)
	clientList.streamlist[clientUniqueCode] = csi
	clientCodeList = append(clientCodeList, clientUniqueCode)

	errch := make(chan error)

	// asynchronously run receiveFromStream and sendToStream methods
	// receive messages - init a go routine
	go receiveFromStream(csi, clientUniqueCode, errch)

	// send messages - init a go routine
	go sendToStream(csi, clientUniqueCode, errch)
	return <-errch // if an error is received the channel will be terminated
}

// receive messages
func receiveFromStream(csi_ Services_ChatServiceServer, clientUniqueCode_ int, errch_ chan error) {
	// loop to continuously receive message from client
	for {
		msg, err := csi_.Recv()
		if err != nil {
			log.Printf("Error in receiving message from client :: %v", err)
			errch_ <- err
		} else {
			// add message to MQue
			messageHandleObject.mu.Lock()

			messageHandleObject.MQue = append(messageHandleObject.MQue, messageUnit{
				ClientName:        msg.Name,
				MessageBody:       msg.Body,
				MessageUniqueCode: rand.Intn(1e8),
				ClientUniqueCode:  clientUniqueCode_,
			})

			messageHandleObject.mu.Unlock()

			log.Printf("%v", messageHandleObject.MQue[len(messageHandleObject.MQue)-1])
		}
	}
}

// send message
func sendToStream(csi_ Services_ChatServiceServer, clientUniqueCode_ int, errch_ chan error) {
	for {
		// loop through messages on MQue
		for {
			time.Sleep(500 * time.Millisecond)
			messageHandleObject.mu.Lock()

			if len(messageHandleObject.MQue) == 0 {
				messageHandleObject.mu.Unlock()
				break
			}

			senderUniqueCode := messageHandleObject.MQue[0].ClientUniqueCode
			senderName4Client := messageHandleObject.MQue[0].ClientName
			message4Client := messageHandleObject.MQue[0].MessageBody

			messageHandleObject.mu.Unlock()

			// send message to each registered client
			for i := len(clientCodeList) - 1; i >= 0; i-- {
				// send message to designated client (do not send to the same client)
				if senderUniqueCode == clientCodeList[i] {
					i--
				}
				if i < 0 {
					break
				}
				clientCode := clientCodeList[i]
				err := clientList.streamlist[clientCode].Send(&FromServer{Name: senderName4Client, Body: message4Client})

				if err != nil {
					errch_ <- err
				}

			}
			messageHandleObject.mu.Lock()
			if len(messageHandleObject.MQue) > 1 {
				messageHandleObject.MQue = messageHandleObject.MQue[1:] // delete the message
			} else {
				messageHandleObject.MQue = []messageUnit{}
			}
			messageHandleObject.mu.Unlock()
		}

		time.Sleep(100 * time.Millisecond)
	}
}
