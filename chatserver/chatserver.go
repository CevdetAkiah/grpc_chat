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
	CQue []client
	mu   sync.Mutex // handle asynchronous read write operations
}

var messageHandleObject = messageHandle{}

type ChatServer struct {
}

// define ChatService
func (is *ChatServer) ChatService(csi Services_ChatServiceServer) error {
	clientUniqueCode := rand.Intn(1e6)
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

			if messageHandleObject.CQue[]

			messageHandleObject.mu.Unlock()

			log.Printf("%v", messageHandleObject.MQue[len(messageHandleObject.MQue)-1])
		}
	}
}

// send message
func sendToStream(csi_ Services_ChatServiceServer, clientUniqueCode_ int, errch_ chan error) {
	// TODO: try cycling through each messageUnit and send to each client using length of MQue. Then delete message
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

			// send message to designated client (do not send to the same client)
			if senderUniqueCode != clientUniqueCode_ {
				err := csi_.Send(&FromServer{Name: senderName4Client, Body: message4Client})

				if err != nil {
					errch_ <- err
				}

				messageHandleObject.mu.Lock()

				if len(messageHandleObject.MQue) > 1 {
					messageHandleObject.MQue = messageHandleObject.MQue[1:] // delete the message
				} else {
					messageHandleObject.MQue = []messageUnit{}
				}
				messageHandleObject.mu.Unlock()
			}
		}

		time.Sleep(100 * time.Millisecond)
	}
}
