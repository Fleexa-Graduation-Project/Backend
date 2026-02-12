package validation

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Fleexa-Graduation-Project/Backend/internal/devices"
	"github.com/Fleexa-Graduation-Project/Backend/models"
)
var (
	ErrInvalidEvent    = errors.New("invalid event")
	ErrInvalidTopic    = errors.New("invalid topic")
	ErrInvalidEnvelope = errors.New("invalid envelope")
	ErrInvalidPayload  = errors.New("invalid payload")
)

