package listener

import (
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/events"
	"github.com/Alessandro-Massarotti-Jr/go-api/internal/providers"
	incrementposttotalcomments "github.com/Alessandro-Massarotti-Jr/go-api/internal/usecases/increment_post_total_comments"
)

type RegisterListenersInput struct {
	EventProvider providers.EventProvider
}

func RegisterListeners(input RegisterListenersInput) {
	input.EventProvider.AddListener(providers.AddListenerInput{
		Topic:   events.PostCommentCreatedTopic,
		Handler: incrementposttotalcomments.PostCommentCreatedListener.Handle,
	})
}
