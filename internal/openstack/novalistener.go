// SPDX short identifier: BSD-2-Clause
// Copyright (c) 2012-2019, Sean Treadway, SoundCloud Ltd.
// All rights reserved.

// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:

// Redistributions of source code must retain the above copyright notice, this
// list of conditions and the following disclaimer.

// Redistributions in binary form must reproduce the above copyright notice, this
// list of conditions and the following disclaimer in the documentation and/or
// other materials provided with the distribution.

// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// +build openstack

package openstack

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

const (
	notificationCreateType = "instance.create.end"
	notificationDeleteType = "instance.delete.end"
)

var initOnce sync.Once
var c *consumer

var messageFilter = []string{notificationCreateType, notificationDeleteType}

type osloMsg struct {
	Payload string `json:"oslo.message"`
	Version string `json:"oslo.version"`
	Type    string
}

var (
	//uri          = "amqp://guest:guest@localhost:5672/"
	exchange     = "nova"
	exchangeType = "topic"
	queue        = "rmd-queue"
	//bindingKey   = "notifications.info"
	consumerTag  = "rmd-consumer"
	curNotifType = ""
)

// NovaListenerStart starts listening for nova notifications
func NovaListenerStart(uri string, bindingKey string) error {

	if bindingKey == "versioned_notifications.info" {
		curNotifType = "versioned"
	} else {
		curNotifType = "unversioned"
	}
	log.Println("Init openstack amqp")
	var err error
	initOnce.Do(func() {
		c, err = newConsumer(uri, exchange, exchangeType, queue, bindingKey, consumerTag)
		if err != nil {
			log.Printf("Failed to create newConsumer: %s", err)
		}
	})
	return err
}

type consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	tag     string
	done    chan error
}

func newConsumer(amqpURI, exchange, exchangeType, queueName, key, ctag string) (*consumer, error) {
	c := &consumer{
		conn:    nil,
		channel: nil,
		tag:     ctag,
		done:    make(chan error),
	}

	var err error
	log.Printf("dialing %q", amqpURI)
	c.conn, err = amqp.Dial(amqpURI)
	if err != nil {
		return nil, fmt.Errorf("Dial: %s", err)
	}

	go func() {
		log.Printf("closing: %s", <-c.conn.NotifyClose(make(chan *amqp.Error)))
	}()

	log.Printf("got Connection, getting Channel")
	c.channel, err = c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("Channel: %s", err)
	}

	log.Printf("got Channel, declaring Exchange (%q)", exchange)
	if err = c.channel.ExchangeDeclare(
		exchange,     // name of the exchange
		exchangeType, // type
		false,        // durable
		false,        // delete when complete
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); err != nil {
		return nil, fmt.Errorf("Exchange Declare: %s", err)
	}

	log.Printf("declared Exchange, declaring Queue %q", queueName)
	queue, err := c.channel.QueueDeclare(
		queueName, // name of the queue
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Queue Declare: %s", err)
	}

	log.Printf("declared Queue (%q %d messages, %d consumers), binding to Exchange (key %q)",
		queue.Name, queue.Messages, queue.Consumers, key)

	if err = c.channel.QueueBind(
		queue.Name, // name of the queue
		key,        // bindingKey
		exchange,   // sourceExchange
		false,      // noWait
		nil,        // arguments
	); err != nil {
		return nil, fmt.Errorf("Queue Bind: %s", err)
	}

	log.Printf("Queue bound to Exchange, starting Consume (consumer tag %q)", c.tag)
	deliveries, err := c.channel.Consume(
		queue.Name, // name
		c.tag,      // consumerTag,
		false,      // noAck
		false,      // exclusive
		false,      // noLocal
		false,      // noWait
		nil,        // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Queue Consume: %s", err)
	}

	go handle(deliveries, c.done)

	return c, nil
}

// Close listener
func Close() error {
	// will close() the deliveries channel
	if err := c.channel.Cancel(c.tag, true); err != nil {
		return fmt.Errorf("Consumer cancel failed: %s", err)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %s", err)
	}

	defer log.Printf("AMQP shutdown OK")

	// wait for handle() to exit
	return <-c.done
}

func filter(input string, words []string) (bool, string) {
	for _, word := range words {
		if strings.Index(input, word) > -1 {
			return true, word
		}
	}
	return false, ""
}

func handle(deliveries <-chan amqp.Delivery, done chan error) {
	for d := range deliveries {
		body := string(d.Body)

		//filter notifications before unmarshaling
		result, _ := filter(body, messageFilter)

		if result {
			//log.Printf("body: \n %s", body)
			osloMsg := osloMsg{}
			err := json.Unmarshal([]byte(body), &osloMsg)
			if err != nil {
				log.Errorf("Oslo unmarshal error: %s", err)
			} else {
				osloMsg.Type = curNotifType
				handleNovaNotification(osloMsg)
			}
		}
		d.Ack(false)
	}
	log.Infoln("handle: deliveries channel closed")
	done <- nil
}
