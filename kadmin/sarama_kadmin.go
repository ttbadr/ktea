package kadmin

import (
	"context"
	"github.com/IBM/sarama"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

type SaramaKafkaAdmin struct {
	client   sarama.Client
	admin    sarama.ClusterAdmin
	addrs    []string
	config   *sarama.Config
	producer sarama.SyncProducer
}

func New(cd ConnectionDetails) (*SaramaKafkaAdmin, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	if cd.SASLConfig != nil {
		config.Net.TLS.Enable = true
		config.Net.SASL.Enable = true
		config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		config.Net.SASL.User = cd.SASLConfig.Username
		config.Net.SASL.Password = cd.SASLConfig.Password
	}

	client, err := sarama.NewClient(cd.BootstrapServers, config)
	if err != nil {
		return nil, err
	}

	admin, err := sarama.NewClusterAdmin(cd.BootstrapServers, config)
	if err != nil {
		return nil, err
	}

	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, err
	}

	return &SaramaKafkaAdmin{
		client:   client,
		admin:    admin,
		addrs:    cd.BootstrapServers,
		producer: producer,
		config:   config,
	}, nil
}

func (ka *SaramaKafkaAdmin) DeleteTopic(topic string) tea.Msg {
	maybeIntroduceLatency()
	err := ka.admin.DeleteTopic(topic)
	if err != nil {
		return KAdminErrorMsg{err}
	}
	return TopicDeletedMsg{topic}
}

func (ka *SaramaKafkaAdmin) CreateTopic(tcd TopicCreationDetails) tea.Msg {
	created := make(chan bool)
	err := make(chan error)

	go ka.doCreateTopic(tcd, created, err)

	return TopicCreationStartedMsg{
		Created: created,
		Err:     err,
	}

}

func (ka *SaramaKafkaAdmin) doCreateTopic(tcd TopicCreationDetails, created chan bool, errChan chan error) {
	err := ka.admin.CreateTopic(tcd.Name, &sarama.TopicDetail{
		NumPartitions:     int32(tcd.NumPartitions),
		ReplicationFactor: 1,
		ReplicaAssignment: nil,
		ConfigEntries:     nil,
	}, false)
	if err != nil {
		errChan <- err
	}
	created <- true
}

func (ka *SaramaKafkaAdmin) ListTopics() tea.Msg {
	errChan := make(chan error)
	topicsChan := make(chan []Topic)

	go ka.doListTopics(errChan, topicsChan)

	return TopicListingStartedMsg{errChan, topicsChan}
}

func (ka *SaramaKafkaAdmin) doListTopics(errChan chan error, topicsChan chan []Topic) {
	maybeIntroduceLatency()
	listResult, err := ka.admin.ListTopics()
	if err != nil {
		log.Errorf("Error %v while listing topics.", err)
		errChan <- err
	}
	partByTopic := make(map[string]Topic)
	for name, topic := range listResult {
		partByTopic[name] = Topic{
			Name:       name,
			Partitions: int(topic.NumPartitions),
			Replicas:   int(topic.ReplicationFactor),
			Isr:        0,
		}
	}
	var topics []Topic
	for _, t := range partByTopic {
		topics = append(topics, Topic{t.Name, t.Partitions, t.Replicas, t.Isr})
	}
	topicsChan <- topics
}

func (ka *SaramaKafkaAdmin) PublishRecord(p *ProducerRecord) PublicationStartedMsg {
	errChan := make(chan error)
	published := make(chan bool)

	go ka.doPublishRecord(p, errChan, published)

	return PublicationStartedMsg{
		Err:       errChan,
		Published: published,
	}
}

func (ka *SaramaKafkaAdmin) doPublishRecord(p *ProducerRecord, errChan chan error, published chan bool) {
	maybeIntroduceLatency()
	var partition int32
	if p.Partition == nil {
		ka.config.Producer.Partitioner = sarama.NewHashPartitioner
	} else {
		partition = int32(*p.Partition)
		ka.config.Producer.Partitioner = sarama.NewManualPartitioner
	}
	_, _, err := ka.producer.SendMessage(&sarama.ProducerMessage{
		Topic:     p.Topic,
		Key:       sarama.StringEncoder(p.Key),
		Value:     sarama.StringEncoder(p.Value),
		Partition: partition,
	})
	if err != nil {
		errChan <- err
	}
	published <- true
}

func (ka *SaramaKafkaAdmin) ReadRecords(ctx context.Context, rd ReadDetails) ReadingStartedMsg {
	rsm := ReadingStartedMsg{
		ConsumerRecord: make(chan ConsumerRecord),
		Err:            make(chan error),
	}
	var partitions []int
	if len(rd.Partitions) == 0 {
		partitions = make([]int, rd.Topic.Partitions)

		for i := range partitions {
			partitions[i] = i
		}
	} else {
		partitions = rd.Partitions
	}
	client, err := sarama.NewConsumerFromClient(ka.client)
	if err != nil {
		return ReadingStartedMsg{}
	}
	for _, partition := range partitions {
		go func(partition int) {
			consumer, err := client.ConsumePartition(rd.Topic.Name, int32(partition), 0)

			if err != nil {
				rsm.Err <- err
				return
			}

			defer consumer.Close()

			messages := consumer.Messages()

			for {
				select {
				case err = <-consumer.Errors():
					rsm.Err <- err
					return
				case <-ctx.Done():
					return
				case msg := <-messages:
					headers := make(map[string]string)
					for _, h := range msg.Headers {
						headers[string(h.Key)] = string(h.Value)
					}

					rsm.ConsumerRecord <- ConsumerRecord{
						Key:       string(msg.Key),
						Value:     string(msg.Value),
						Partition: int64(msg.Partition),
						Offset:    msg.Offset,
						Headers:   headers,
					}
				}
			}
		}(partition)

	}
	return rsm
}

func (ka *SaramaKafkaAdmin) ListConfigs(topic string) tea.Msg {
	errChan := make(chan error)
	configsChan := make(chan map[string]string)

	go ka.doListConfigs(topic, configsChan, errChan)

	return TopicConfigListingStartedMsg{
		errChan,
		configsChan,
	}
}

func (ka *SaramaKafkaAdmin) doListConfigs(topic string, configsChan chan map[string]string, errorChan chan error) {
	maybeIntroduceLatency()
	configsResp, err := ka.admin.DescribeConfig(sarama.ConfigResource{
		Type: TOPIC_RESOURCE_TYPE,
		Name: topic,
	})
	if err != nil {
		errorChan <- err
		return
	}
	configs := make(map[string]string)
	for _, e := range configsResp {
		configs[e.Name] = e.Value
	}
	configsChan <- configs
}

func (ka *SaramaKafkaAdmin) UpdateConfig(t TopicConfigToUpdate) tea.Msg {
	err := ka.admin.AlterConfig(
		TOPIC_RESOURCE_TYPE,
		t.Topic,
		map[string]*string{t.Key: &t.Value},
		false,
	)
	if err != nil {
		return KAdminErrorMsg{err}
	}
	return TopicConfigUpdatedMsg{}
}

type TopicPartitionOffset struct {
	Topic     string
	Partition int32
	Offset    int64
}

func (ka *SaramaKafkaAdmin) ListOffsets(group string) tea.Msg {
	errChan := make(chan error)
	offsets := make(chan []TopicPartitionOffset)

	go ka.doListOffsets(group, offsets, errChan)

	return OffsetListingStartedMsg{
		errChan,
		offsets,
	}
}

func (ka *SaramaKafkaAdmin) doListOffsets(group string, offsetsChan chan []TopicPartitionOffset, errChan chan error) {
	maybeIntroduceLatency()
	listResult, err := ka.admin.ListConsumerGroupOffsets(group, nil)
	if err != nil {
		errChan <- err
	}

	var topicPartitionOffsets []TopicPartitionOffset
	for t, m := range listResult.Blocks {
		for p, block := range m {
			topicPartitionOffsets = append(topicPartitionOffsets, TopicPartitionOffset{
				Topic:     t,
				Partition: p,
				Offset:    block.Offset,
			})
		}
	}

	offsetsChan <- topicPartitionOffsets
}

func (ka *SaramaKafkaAdmin) ListConsumerGroups() tea.Msg {
	errChan := make(chan error)
	groupsChan := make(chan []*ConsumerGroup)

	go ka.doListConsumerGroups(groupsChan, errChan)

	return ConsumerGroupListingStartedMsg{errChan, groupsChan}
}

func (ka *SaramaKafkaAdmin) doListConsumerGroups(groupsChan chan []*ConsumerGroup, errorChan chan error) {
	maybeIntroduceLatency()
	if listGroupResponse, err := ka.admin.ListConsumerGroups(); err != nil {
		errorChan <- err
	} else {
		var consumerGroups []*ConsumerGroup
		var groupNames []string
		var groupByName = make(map[string]*ConsumerGroup)

		for name, _ := range listGroupResponse {
			consumerGroup := ConsumerGroup{Name: name}
			consumerGroups = append(consumerGroups, &consumerGroup)
			groupByName[name] = &consumerGroup
			groupNames = append(groupNames, name)
		}

		describeConsumerGroupResponse, err := ka.admin.DescribeConsumerGroups(groupNames)
		if err != nil {
			errorChan <- err
			return
		}

		for _, groupDescription := range describeConsumerGroupResponse {
			group := groupByName[groupDescription.GroupId]
			var groupMembers []GroupMember
			for _, m := range groupDescription.Members {
				member := GroupMember{}
				member.MemberId = m.MemberId
				member.ClientId = m.ClientId
				member.ClientHost = m.ClientHost
				groupMembers = append(groupMembers, member)
			}
			group.Members = groupMembers
		}
		groupsChan <- consumerGroups
	}
}
