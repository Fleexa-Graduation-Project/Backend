package validation

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Fleexa-Graduation-Project/Backend/internal/devices"
	"github.com/Fleexa-Graduation-Project/Backend/models"
)
var (
	nvalidEvent    = errors.New("invalid event")
	InvalidTopic    = errors.New("invalid topic")
	InvalidEnvelope = errors.New("invalid envelope")
	ErrInvalidPayload  = errors.New("invalid payload")
)

