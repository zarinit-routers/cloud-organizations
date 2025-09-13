package queue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/charmbracelet/log"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spf13/viper"
	"github.com/zarinit-routers/cloud-organizations/pkg/models"
)

// Publisher publishes domain events to RabbitMQ (topic exchange).
type Publisher struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	exchange string
	enabled  bool
}

// NewFromViper creates a publisher; if rabbitmq.url is empty, publisher is disabled (logs only).
func NewFromViper() *Publisher {
	url := viper.GetString("rabbitmq.url")
	exchange := viper.GetString("rabbitmq.exchange")
	if url == "" {
		log.Warn("rabbitmq.url not configured, events will be logged only")
		return &Publisher{enabled: false, exchange: exchange}
	}
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Error("failed to connect to rabbitmq", "error", err)
		return &Publisher{enabled: false, exchange: exchange}
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Error("failed to open channel", "error", err)
		_ = conn.Close()
		return &Publisher{enabled: false, exchange: exchange}
	}
	// Declare topic exchange idempotently
	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		log.Error("failed to declare exchange", "error", err)
		_ = ch.Close()
		_ = conn.Close()
		return &Publisher{enabled: false, exchange: exchange}
	}
	log.Info("rabbitmq publisher ready", "exchange", exchange)
	return &Publisher{conn: conn, ch: ch, exchange: exchange, enabled: true}
}

func (p *Publisher) Close() {
	if p.ch != nil {
		_ = p.ch.Close()
	}
	if p.conn != nil {
		_ = p.conn.Close()
	}
}

type orgEventPayload struct {
	ID         string              `json:"id"`
	TenantID   *string             `json:"tenantId,omitempty"`
	Name       string              `json:"name"`
	LegalCode  *string             `json:"legalCode,omitempty"`
	Status     string              `json:"status"`
	Tags       []string            `json:"tags,omitempty"`
	Addresses  []models.OrgAddress `json:"addresses,omitempty"`
	Contacts   []models.OrgContact `json:"contacts,omitempty"`
	OccurredAt string              `json:"occurredAt"`
	TraceID    string              `json:"traceId"`
	// Optional versioning
	SchemaVersion string `json:"schemaVersion,omitempty"`
}

func (p *Publisher) publish(ctx context.Context, routingKey string, pl any) error {
	body, _ := json.Marshal(pl)
	if !p.enabled {
		log.Debug("event", "routing_key", routingKey, "payload", string(body))
		return nil
	}
	return p.ch.PublishWithContext(ctx, p.exchange, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
		Timestamp:    time.Now(),
	})
}

func (p *Publisher) OrganizationCreated(ctx context.Context, org models.Organization, traceID string) error {
	key := viper.GetString("rabbitmq.routing_keys.created")
	if key == "" {
		key = "organization.created"
	}
	return p.publish(ctx, key, toPayload(org, traceID))
}

func (p *Publisher) OrganizationUpdated(ctx context.Context, org models.Organization, traceID string) error {
	key := viper.GetString("rabbitmq.routing_keys.updated")
	if key == "" {
		key = "organization.updated"
	}
	return p.publish(ctx, key, toPayload(org, traceID))
}

func (p *Publisher) OrganizationDeleted(ctx context.Context, org models.Organization, traceID string) error {
	key := viper.GetString("rabbitmq.routing_keys.deleted")
	if key == "" {
		key = "organization.deleted"
	}
	return p.publish(ctx, key, toPayload(org, traceID))
}

func toPayload(org models.Organization, traceID string) orgEventPayload {
	var tenant *string
	if org.TenantID != nil {
		t := org.TenantID.String()
		tenant = &t
	}
	return orgEventPayload{
		ID:            org.ID.String(),
		TenantID:      tenant,
		Name:          org.Name,
		LegalCode:     org.LegalCode,
		Status:        org.Status,
		Tags:          org.Tags,
		Addresses:     org.Addresses,
		Contacts:      org.Contacts,
		OccurredAt:    org.UpdatedAt.UTC().Format(time.RFC3339),
		TraceID:       traceID,
		SchemaVersion: "1",
	}
}
