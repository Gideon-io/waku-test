package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/waku-org/go-waku/waku/v2/node"
	"github.com/waku-org/go-waku/waku/v2/payload"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	"github.com/waku-org/go-waku/waku/v2/protocol/pb"
	"github.com/waku-org/go-waku/waku/v2/protocol/relay"
	"github.com/waku-org/go-waku/waku/v2/utils"
	"google.golang.org/protobuf/proto"
)

var pubsubTopicStr string

func main() {
	//===== Defining variables for creating the waku instance =====//

	//define content topic here
	/*
		Content topics have 2 purposes: filtering and routing. Filtering is done by changing the {content-topic-name} field. As this part is not hashed, it will not affect routing (shard selection).
		The {application-name} and {version-of-the-application} fields do affect routing. Using multiple content topics with different {application-name} field has advantages and disadvantages.
		It increases the traffic a relay node is subjected to when subscribed to all topics. It also allows relay and light nodes to subscribe to a subset of all topics.
	*/
	cTopic, err := protocol.NewContentTopic("relay-test", "1", "test", "proto") //short length used here /{application-name}/{version-of-the-application}/{content-topic-name}/{encoding}
	if err != nil {
		log.Fatal(err)
	}
	contentTopic := cTopic.String()

	//define host address
	hostAddr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:0")

	//generate a random hex and set it to a private key
	key, err := randomHex(32)
	if err != nil {
		log.Fatal("Could not generate random key", err)
	} else {
		fmt.Printf("Key generated: %v\n", key)
	}
	prvKey, err := crypto.HexToECDSA(key)
	if err != nil {
		log.Fatal("Could not convert hex into ecdsa key", err)

	}

	// define the context
	ctx := context.Background()
	//===== *END* for Defining variables for creating the waku instance =====//

	//create the waku node/instance

	wakuNode, err := node.New(
		node.WithWakuRelay(),
		node.WithPrivateKey(prvKey),
		node.WithHostAddress(hostAddr),
		node.WithNTP(),
		node.WithClusterID(uint16(1)), // ..?
	)

	if err != nil {
		log.Fatal(err)
	}
	//start the waku node/instance
	if err := wakuNode.Start(context.Background()); err != nil {
		fmt.Println(err)
		return
	}

	pubsubTopic := protocol.NewStaticShardingPubsubTopic(uint16(1), uint16(0))
	pubsubTopicStr = pubsubTopic.String()

	write(ctx, wakuNode, contentTopic, "hello world")
	readLoop(ctx, wakuNode, contentTopic)

}

func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func write(ctx context.Context, wakuNode *node.WakuNode, contentTopic string, msgContent string) {

	//populate payload (the message essentially with all the details)
	var version uint32 = 0

	msg := &pb.WakuMessage{
		Payload:      []byte(msgContent),
		Version:      proto.Uint32(version),
		ContentTopic: contentTopic,
		Timestamp:    utils.GetUnixEpoch(wakuNode.Timesource()),
	}
	//publish the message
	hash, err := wakuNode.Relay().Publish(ctx, msg, relay.WithPubSubTopic(pubsubTopicStr))
	if err != nil {
		log.Fatalf("Error sending a message: %v", err)
	} else {
		fmt.Printf("Succesfully sent || Message ID: %v", hash)
	}
}

func readLoop(ctx context.Context, wakuNode *node.WakuNode, contentTopic string) {
	contentFilter := protocol.NewContentFilter(pubsubTopicStr, contentTopic)
	sub, err := wakuNode.Relay().Subscribe(ctx, contentFilter)
	if err != nil {
		fmt.Println("Failed to subscribe", err)
		return
	}

	for envelope := range sub[0].Ch {
		if envelope.Message().ContentTopic != contentTopic {
			continue
		}

		payload, err := payload.DecodePayload(envelope.Message(), &payload.KeyInfo{Kind: payload.None})
		if err != nil {
			log.Println("Error decoding payload", err)
			continue
		}

		log.Print("Received msg, ", string(payload.Data))
	}
}
