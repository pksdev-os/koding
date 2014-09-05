package popularpost

import (
	"fmt"
	"socialapi/config"
	"socialapi/models"
	"time"

	"github.com/koding/logging"
	"github.com/koding/redis"
	"github.com/streadway/amqp"
)

var (
	PopularPostKey = "popularpost"
)

type Controller struct {
	log   logging.Logger
	redis *redis.RedisSession
}

func (t *Controller) DefaultErrHandler(delivery amqp.Delivery, err error) bool {
	if delivery.Redelivered {
		t.log.Error("Redelivered message gave error again, putting to maintenance queue", err)
		delivery.Ack(false)
		return true
	}

	t.log.Error("an error occured putting message back to queue", err)
	delivery.Nack(false, true)
	return false
}

func New(log logging.Logger, redis *redis.RedisSession) *Controller {
	return &Controller{
		log:   log,
		redis: redis,
	}
}

func (f *Controller) InteractionSaved(i *models.Interaction) error {
	return f.handleInteractionEvent(1, i)
}

func (f *Controller) InteractionDeleted(i *models.Interaction) error {
	return f.handleInteractionEvent(-1, i)
}

func (f *Controller) handleInteractionEvent(incrementCount int, i *models.Interaction) error {
	cm, err := models.ChannelMessageById(i.MessageId)
	if err != nil {
		return err
	}

	c, err := models.ChannelById(cm.InitialChannelId)
	if err != nil {
		return err
	}

	if !f.isEligible(c, cm) {
		f.log.Error(fmt.Sprintf("Not eligible Interaction Id:%d", i.Id))
		return nil
	}

	_, err = f.redis.SortedSetIncrBy(GetDailyKey(c, cm, i), incrementCount, i.MessageId)
	if err != nil {
		return err
	}

	return nil
}

func (f *Controller) isEligible(c *models.Channel, cm *models.ChannelMessage) bool {
	if c.MetaBits.Is(models.Troll) {
		fmt.Println("channel troll")
		return false
	}

	if cm.MetaBits.Is(models.Troll) {
		fmt.Println("channel message troll")
		return false
	}

	if c.PrivacyConstant != models.Channel_PRIVACY_PUBLIC {
		fmt.Println("channel is private")
		return false
	}

	if cm.TypeConstant != models.ChannelMessage_TYPE_POST {
		fmt.Println("channel message is not a post")
		return false
	}

	return true
}

func PreparePopularPostKey(group, channelName, statisticName string, year, dateNumber int) string {
	return fmt.Sprintf(
		"%s:%s:%s:%s:%d:%s:%d",
		config.MustGet().Environment,
		group,
		PopularPostKey,
		channelName,
		year,
		statisticName,
		dateNumber,
	)
}

func GetDailyKey(c *models.Channel, cm *models.ChannelMessage, i *models.Interaction) string {
	day := 0
	year := 2014

	if !i.CreatedAt.IsZero() {
		day = i.CreatedAt.UTC().YearDay()
		year, _, _ = i.CreatedAt.UTC().Date()
	} else {
		// no need to convert it to UTC
		now := time.Now().UTC()
		day = now.YearDay()
		year, _, _ = now.Date()
	}

	return PreparePopularPostKey(c.GroupName, c.Name, "daily", year, day)
}
