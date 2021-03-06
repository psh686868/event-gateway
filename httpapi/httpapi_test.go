package httpapi_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/julienschmidt/httprouter"
	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/httpapi"
	"github.com/serverless/event-gateway/mock"
	"github.com/serverless/event-gateway/subscription"
	"github.com/stretchr/testify/assert"

	httpprovider "github.com/serverless/event-gateway/providers/http"
)

func TestGetEventType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, eventTypes, _, _ := setup(ctrl)

	t.Run("event type returned", func(t *testing.T) {
		returnedType := &event.Type{
			Space: "default",
			Name:  "test.event",
		}
		eventTypes.EXPECT().GetEventType("default", event.TypeName("test.event")).Return(returnedType, nil)

		resp := request(router, http.MethodGet, "/v1/spaces/default/eventtypes/test.event", nil)

		eventType := &event.Type{}
		json.Unmarshal(resp.Body.Bytes(), eventType)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "default", eventType.Space)
		assert.Equal(t, event.TypeName("test.event"), eventType.Name)
	})

	t.Run("not found", func(t *testing.T) {
		returnedErr := &event.ErrEventTypeNotFound{Name: event.TypeName("test.event")}
		eventTypes.EXPECT().GetEventType(gomock.Any(), gomock.Any()).Return(nil, returnedErr)

		resp := request(router, http.MethodGet, "/v1/spaces/default/eventtypes/test.event", nil)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		assert.Equal(t, `Event Type "test.event" not found.`, httpresp.Errors[0].Message)
	})

	t.Run("internal error", func(t *testing.T) {
		eventTypes.EXPECT().GetEventType(gomock.Any(), gomock.Any()).Return(nil, errors.New("processing failed"))

		resp := request(router, http.MethodGet, "/v1/spaces/default/eventtypes/test.event", nil)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.Equal(t, "processing failed", httpresp.Errors[0].Message)
	})
}

func TestGetEventTypes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, eventTypes, _, _ := setup(ctrl)

	t.Run("event types returned", func(t *testing.T) {
		returnedList := event.Types{{
			Space: "default",
			Name:  event.TypeName("test.event"),
		}}
		eventTypes.EXPECT().GetEventTypes("default").Return(returnedList, nil)

		resp := request(router, http.MethodGet, "/v1/spaces/default/eventtypes", nil)

		types := &httpapi.EventTypesResponse{}
		json.Unmarshal(resp.Body.Bytes(), types)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "default", types.EventTypes[0].Space)
		assert.Equal(t, event.TypeName("test.event"), types.EventTypes[0].Name)
	})

	t.Run("internal error", func(t *testing.T) {
		eventTypes.EXPECT().GetEventTypes(gomock.Any()).Return(nil, errors.New("processing failed"))

		resp := request(router, http.MethodGet, "/v1/spaces/default/eventtypes", nil)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.Equal(t, "processing failed", httpresp.Errors[0].Message)
	})
}

func TestCreateEventType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, eventTypes, _, _ := setup(ctrl)

	typePayload := []byte(`{"name":"test.event","space":"test1"}`)

	t.Run("event type created", func(t *testing.T) {
		eventType := &event.Type{Space: "default", Name: event.TypeName("test.event")}
		eventTypes.EXPECT().CreateEventType(eventType).Return(eventType, nil)

		resp := request(router, http.MethodPost, "/v1/spaces/default/eventtypes", typePayload)

		returnedType := &event.Type{}
		json.Unmarshal(resp.Body.Bytes(), returnedType)
		assert.Equal(t, http.StatusCreated, resp.Code)
		assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
		assert.Equal(t, event.TypeName("test.event"), returnedType.Name)
		assert.Equal(t, "default", returnedType.Space)
	})

	t.Run("event type already exists", func(t *testing.T) {
		eventTypes.EXPECT().CreateEventType(gomock.Any()).
			Return(nil, &event.ErrEventTypeAlreadyExists{Name: event.TypeName("test.event")})

		resp := request(router, http.MethodPost, "/v1/spaces/default/eventtypes", typePayload)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, `Event Type "test.event" already exists.`, httpresp.Errors[0].Message)
	})

	t.Run("validation error", func(t *testing.T) {
		eventTypes.EXPECT().CreateEventType(gomock.Any()).
			Return(nil, &event.ErrEventTypeValidation{Message: "some error"})

		payload := []byte(`{"name":"test"}`)
		resp := request(router, http.MethodPost, "/v1/spaces/default/eventtypes", payload)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, "Event Type doesn't validate. Validation error: some error", httpresp.Errors[0].Message)
	})

	t.Run("malformed JSON", func(t *testing.T) {
		resp := request(router, http.MethodPost, "/v1/spaces/default/eventtypes", []byte("{"))

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, "Event Type doesn't validate. Validation error: unexpected EOF", httpresp.Errors[0].Message)
	})

	t.Run("internal error", func(t *testing.T) {
		eventTypes.EXPECT().CreateEventType(gomock.Any()).Return(nil, errors.New("processing error"))

		resp := request(router, http.MethodPost, "/v1/spaces/default/eventtypes", typePayload)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.Equal(t, `processing error`, httpresp.Errors[0].Message)
	})
}

func TestUpdateEventType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, eventTypes, _, _ := setup(ctrl)

	typePayload := []byte(`{"name":"test.event","space":"test1"}`)

	t.Run("event type updated", func(t *testing.T) {
		eventType := &event.Type{Space: "default", Name: event.TypeName("test.event")}
		eventTypes.EXPECT().UpdateEventType(eventType).Return(eventType, nil)

		resp := request(router, http.MethodPut, "/v1/spaces/default/eventtypes/test.event", typePayload)

		returnedType := &event.Type{}
		json.Unmarshal(resp.Body.Bytes(), returnedType)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
		assert.Equal(t, event.TypeName("test.event"), returnedType.Name)
		assert.Equal(t, "default", returnedType.Space)
	})

	t.Run("event type doesn't exists", func(t *testing.T) {
		eventTypes.EXPECT().UpdateEventType(gomock.Any()).
			Return(nil, &event.ErrEventTypeNotFound{Name: event.TypeName("test.event")})

		resp := request(router, http.MethodPut, "/v1/spaces/default/eventtypes/test.event", typePayload)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		assert.Equal(t, `Event Type "test.event" not found.`, httpresp.Errors[0].Message)
	})

	t.Run("authorizer doesn't exists error", func(t *testing.T) {
		eventTypes.EXPECT().UpdateEventType(gomock.Any()).
			Return(nil, &event.ErrAuthorizerDoesNotExists{})

		payload := []byte(`{"name":"test"}`)
		resp := request(router, http.MethodPut, "/v1/spaces/default/eventtypes/test.event", payload)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, "Authorizer function doesn't exists.", httpresp.Errors[0].Message)
	})

	t.Run("validation error", func(t *testing.T) {
		eventTypes.EXPECT().UpdateEventType(gomock.Any()).
			Return(nil, &event.ErrEventTypeValidation{Message: "some error"})

		payload := []byte(`{"name":"test"}`)
		resp := request(router, http.MethodPut, "/v1/spaces/default/eventtypes/test.event", payload)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, "Event Type doesn't validate. Validation error: some error", httpresp.Errors[0].Message)
	})

	t.Run("malformed JSON", func(t *testing.T) {
		resp := request(router, http.MethodPut, "/v1/spaces/default/eventtypes/test.event", []byte("{"))

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, "Event Type doesn't validate. Validation error: unexpected EOF", httpresp.Errors[0].Message)
	})

	t.Run("internal error", func(t *testing.T) {
		eventTypes.EXPECT().UpdateEventType(gomock.Any()).Return(nil, errors.New("processing error"))

		resp := request(router, http.MethodPut, "/v1/spaces/default/eventtypes/test.event", typePayload)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.Equal(t, `processing error`, httpresp.Errors[0].Message)
	})
}

func TestDeleteEventType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, eventTypes, _, _ := setup(ctrl)

	t.Run("event type deleted", func(t *testing.T) {
		eventTypes.EXPECT().DeleteEventType("default", event.TypeName("test.event")).Return(nil)

		resp := request(router, http.MethodDelete, "/v1/spaces/default/eventtypes/test.event", nil)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusNoContent, resp.Code)
	})

	t.Run("event type has subscriptions", func(t *testing.T) {
		eventTypes.EXPECT().DeleteEventType(gomock.Any(), gomock.Any()).Return(&event.ErrEventTypeHasSubscriptions{})

		resp := request(router, http.MethodDelete, "/v1/spaces/default/eventtypes/test.event", nil)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, "Event type cannot be deleted because there are subscriptions using it.", httpresp.Errors[0].Message)
	})

	t.Run("event type not found", func(t *testing.T) {
		eventTypes.EXPECT().DeleteEventType(gomock.Any(), gomock.Any()).Return(&event.ErrEventTypeNotFound{Name: event.TypeName("test.event")})

		resp := request(router, http.MethodDelete, "/v1/spaces/default/eventtypes/test.event", nil)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		assert.Equal(t, `Event Type "test.event" not found.`, httpresp.Errors[0].Message)
	})

	t.Run("internal error", func(t *testing.T) {
		eventTypes.EXPECT().DeleteEventType(gomock.Any(), gomock.Any()).Return(errors.New("internal error"))

		resp := request(router, http.MethodDelete, "/v1/spaces/default/eventtypes/test.event", nil)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.Equal(t, "internal error", httpresp.Errors[0].Message)
	})
}

func TestGetFunction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, _, functions, _ := setup(ctrl)

	t.Run("function returned", func(t *testing.T) {
		returnedFn := &function.Function{
			Space:        "default",
			ID:           function.ID("func1"),
			ProviderType: httpprovider.Type,
			Provider:     &httpprovider.HTTP{URL: "http://example.com"},
		}
		functions.EXPECT().GetFunction("default", function.ID("func1")).Return(returnedFn, nil)

		resp := request(router, http.MethodGet, "/v1/spaces/default/functions/func1", nil)

		fn := &function.Function{}
		json.Unmarshal(resp.Body.Bytes(), fn)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "default", fn.Space)
		assert.Equal(t, function.ID("func1"), fn.ID)
		assert.Equal(t, httpprovider.Type, fn.ProviderType)
		assert.Equal(t, &httpprovider.HTTP{URL: "http://example.com"}, fn.Provider)
	})

	t.Run("not found", func(t *testing.T) {
		returnedErr := &function.ErrFunctionNotFound{ID: function.ID("func1")}
		functions.EXPECT().GetFunction(gomock.Any(), gomock.Any()).Return(nil, returnedErr)

		resp := request(router, http.MethodGet, "/v1/spaces/default/functions/func1", nil)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		assert.Equal(t, `Function "func1" not found.`, httpresp.Errors[0].Message)
	})

	t.Run("internal error", func(t *testing.T) {
		functions.EXPECT().GetFunction(gomock.Any(), gomock.Any()).Return(nil, errors.New("processing failed"))

		resp := request(router, http.MethodGet, "/v1/spaces/default/functions/func1", nil)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.Equal(t, "processing failed", httpresp.Errors[0].Message)
	})
}

func TestGetFunctions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, _, functions, _ := setup(ctrl)

	t.Run("functions returned", func(t *testing.T) {
		returnedList := function.Functions{{
			ID:           function.ID("func1"),
			Space:        "default",
			ProviderType: httpprovider.Type,
			Provider:     &httpprovider.HTTP{},
		}}
		functions.EXPECT().GetFunctions("default").Return(returnedList, nil)

		resp := request(router, http.MethodGet, "/v1/spaces/default/functions", nil)

		fns := &httpapi.FunctionsResponse{}
		json.Unmarshal(resp.Body.Bytes(), fns)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "default", fns.Functions[0].Space)
		assert.Equal(t, function.ID("func1"), fns.Functions[0].ID)
	})

	t.Run("internal error", func(t *testing.T) {
		functions.EXPECT().GetFunctions(gomock.Any()).Return(nil, errors.New("processing failed"))

		resp := request(router, http.MethodGet, "/v1/spaces/default/functions", nil)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.Equal(t, "processing failed", httpresp.Errors[0].Message)
	})
}

func TestRegisterFunction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, _, functions, _ := setup(ctrl)

	fnPayload := []byte(`{"functionId":"func1","space":"test1","type":"http","provider":{"url":"http://example.com"}}`)

	t.Run("function registered", func(t *testing.T) {
		fn := &function.Function{
			ID:           function.ID("func1"),
			Space:        "test1",
			ProviderType: httpprovider.Type,
			Provider: &httpprovider.HTTP{
				URL: "http://example.com",
			},
		}
		functions.EXPECT().RegisterFunction(fn).Return(fn, nil)

		resp := request(router, http.MethodPost, "/v1/spaces/test1/functions", fnPayload)

		fn = &function.Function{}
		json.Unmarshal(resp.Body.Bytes(), fn)
		assert.Equal(t, http.StatusCreated, resp.Code)
		assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
		assert.Equal(t, function.ID("func1"), fn.ID)
		assert.Equal(t, "test1", fn.Space)
	})

	t.Run("function already exists", func(t *testing.T) {
		functions.EXPECT().RegisterFunction(gomock.Any()).
			Return(nil, &function.ErrFunctionAlreadyRegistered{ID: function.ID("func1")})

		resp := request(router, http.MethodPost, "/v1/spaces/default/functions", fnPayload)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, `Function "func1" already registered.`, httpresp.Errors[0].Message)
	})

	t.Run("validation error", func(t *testing.T) {
		functions.EXPECT().RegisterFunction(gomock.Any()).
			Return(nil, &function.ErrFunctionValidation{Message: "wrong function ID format"})

		payload := []byte(`{"functionID":"/","type":"http","provider":{"url":"http://test.com"}}}`)
		resp := request(router, http.MethodPost, "/v1/spaces/default/functions", payload)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, "Function doesn't validate. Validation error: wrong function ID format", httpresp.Errors[0].Message)
	})

	t.Run("malformed JSON", func(t *testing.T) {
		resp := request(router, http.MethodPost, "/v1/spaces/default/functions", []byte(`{`))

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, "Function doesn't validate. Validation error: unexpected EOF", httpresp.Errors[0].Message)
	})

	t.Run("internal error", func(t *testing.T) {
		functions.EXPECT().RegisterFunction(gomock.Any()).Return(nil, errors.New("processing error"))

		resp := request(router, http.MethodPost, "/v1/spaces/default/functions", fnPayload)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.Equal(t, `processing error`, httpresp.Errors[0].Message)
	})
}

func TestDeleteFunction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, _, functions, _ := setup(ctrl)

	t.Run("function deleted", func(t *testing.T) {
		functions.EXPECT().DeleteFunction("default", function.ID("func1")).Return(nil)

		resp := request(router, http.MethodDelete, "/v1/spaces/default/functions/func1", nil)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusNoContent, resp.Code)
	})

	t.Run("function has subscriptions", func(t *testing.T) {
		functions.EXPECT().DeleteFunction(gomock.Any(), gomock.Any()).Return(&function.ErrFunctionHasSubscriptions{})

		resp := request(router, http.MethodDelete, "/v1/spaces/default/functions/func1", nil)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, "Function cannot be deleted because it's subscribed to a least one event.", httpresp.Errors[0].Message)
	})

	t.Run("function not found", func(t *testing.T) {
		functions.EXPECT().DeleteFunction(gomock.Any(), gomock.Any()).Return(&function.ErrFunctionNotFound{ID: function.ID("testid")})

		resp := request(router, http.MethodDelete, "/v1/spaces/default/functions/func1", nil)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		assert.Equal(t, `Function "testid" not found.`, httpresp.Errors[0].Message)
	})

	t.Run("internal error", func(t *testing.T) {
		functions.EXPECT().DeleteFunction(gomock.Any(), gomock.Any()).Return(errors.New("internal error"))

		resp := request(router, http.MethodDelete, "/v1/spaces/default/functions/func1", nil)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.Equal(t, "internal error", httpresp.Errors[0].Message)
	})
}

func TestUpdateSubscription(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, _, _, subscriptions := setup(ctrl)

	updateSub := &subscription.Subscription{
		Space:      "default",
		ID:         subscription.ID("testid"),
		Type:       subscription.TypeSync,
		EventType:  "http.request",
		FunctionID: "func",
		Method:     "GET",
		Path:       "/",
		CORS: &subscription.CORS{
			Origins:          []string{"*"},
			Methods:          []string{"HEAD", "GET", "POST"},
			Headers:          []string{"Origin", "Accept", "Content-Type"},
			AllowCredentials: false,
		},
	}
	updatedValue := []byte(`{"space":"default","subscriptionId":"testid","type":"sync",` +
		`"eventType":"http.request","functionId":"func","method":"GET","path":"/",` +
		`"cors":{"origins":["*"],"methods":["HEAD","GET","POST"],` +
		`"headers":["Origin","Accept","Content-Type"],"allowCredentials":false}}`)

	t.Run("subscription updated", func(t *testing.T) {
		subscriptions.EXPECT().UpdateSubscription(subscription.ID("testid"), updateSub).Return(updateSub, nil)

		resp := request(router, http.MethodPut, "/v1/spaces/default/subscriptions/testid", updatedValue)

		sub := &subscription.Subscription{}
		json.Unmarshal(resp.Body.Bytes(), sub)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "default", sub.Space)
		assert.Equal(t, subscription.ID("testid"), sub.ID)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		resp := request(router, http.MethodPut, "/v1/spaces/default/subscriptions/testid", []byte(`{"name":"te`))

		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("invalid subscription payload", func(t *testing.T) {
		subscriptions.EXPECT().UpdateSubscription(gomock.Any(), gomock.Any()).Return(nil, &subscription.ErrInvalidSubscriptionUpdate{Field: "FunctionID"})

		resp := request(router, http.MethodPut, "/v1/spaces/default/subscriptions/testid", updatedValue)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, `Invalid update. 'FunctionID' of existing subscription cannot be updated.`, httpresp.Errors[0].Message)
	})

	t.Run("subscription not found", func(t *testing.T) {
		subscriptions.EXPECT().UpdateSubscription(gomock.Any(), gomock.Any()).Return(nil, &subscription.ErrSubscriptionNotFound{ID: subscription.ID("testid")})

		resp := request(router, http.MethodPut, "/v1/spaces/default/subscriptions/testid", updatedValue)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		assert.Equal(t, `Subscription "testid" not found.`, httpresp.Errors[0].Message)
	})

	t.Run("function not found", func(t *testing.T) {
		subscriptions.EXPECT().UpdateSubscription(gomock.Any(), gomock.Any()).Return(nil, &function.ErrFunctionNotFound{ID: function.ID("func")})

		resp := request(router, http.MethodPut, "/v1/spaces/default/subscriptions/testid", updatedValue)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, `Function "func" not found.`, httpresp.Errors[0].Message)
	})

	t.Run("validation error", func(t *testing.T) {
		subscriptions.EXPECT().UpdateSubscription(gomock.Any(), gomock.Any()).Return(nil, &subscription.ErrSubscriptionValidation{Message: ""})

		resp := request(router, http.MethodPut, "/v1/spaces/default/subscriptions/testid", updatedValue)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Contains(t, httpresp.Errors[0].Message, "Subscription doesn't validate. Validation error")
	})

	t.Run("internal error", func(t *testing.T) {
		subscriptions.EXPECT().UpdateSubscription(gomock.Any(), gomock.Any()).Return(nil, errors.New("processing failed"))

		resp := request(router, http.MethodPut, "/v1/spaces/default/subscriptions/testid", updatedValue)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.Equal(t, "processing failed", httpresp.Errors[0].Message)
	})
}

func TestDeleteSubscription(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, _, _, subscriptions := setup(ctrl)

	t.Run("subscription deleted", func(t *testing.T) {
		subscriptions.EXPECT().DeleteSubscription("default", subscription.ID("testid")).Return(nil)

		resp := request(router, http.MethodDelete, "/v1/spaces/default/subscriptions/testid", nil)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusNoContent, resp.Code)
	})

	t.Run("subscriptions not found", func(t *testing.T) {
		subscriptions.EXPECT().DeleteSubscription(gomock.Any(), gomock.Any()).Return(&subscription.ErrSubscriptionNotFound{ID: subscription.ID("testid")})

		resp := request(router, http.MethodDelete, "/v1/spaces/default/subscriptions/testid", nil)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		assert.Equal(t, `Subscription "testid" not found.`, httpresp.Errors[0].Message)
	})

	t.Run("internal error", func(t *testing.T) {
		subscriptions.EXPECT().DeleteSubscription(gomock.Any(), gomock.Any()).Return(errors.New("internal error"))

		resp := request(router, http.MethodDelete, "/v1/spaces/default/subscriptions/testid", nil)

		httpresp := &httpapi.Response{}
		json.Unmarshal(resp.Body.Bytes(), httpresp)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.Equal(t, "internal error", httpresp.Errors[0].Message)
	})
}

func request(router *httprouter.Router, method string, url string, payload []byte) *httptest.ResponseRecorder {
	resp := httptest.NewRecorder()
	body := bytes.NewReader(payload)
	req, _ := http.NewRequest(method, url, body)
	router.ServeHTTP(resp, req)

	return resp
}

func setup(ctrl *gomock.Controller) (
	*httprouter.Router,
	*mock.MockEventTypeService,
	*mock.MockFunctionService,
	*mock.MockSubscriptionService,
) {
	router := httprouter.New()
	eventTypes := mock.NewMockEventTypeService(ctrl)
	functions := mock.NewMockFunctionService(ctrl)
	subscriptions := mock.NewMockSubscriptionService(ctrl)

	httpapi := &httpapi.HTTPAPI{
		EventTypes:    eventTypes,
		Functions:     functions,
		Subscriptions: subscriptions,
	}
	httpapi.RegisterRoutes(router)

	return router, eventTypes, functions, subscriptions
}
