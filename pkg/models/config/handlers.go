package config

import (
	"encoding/json"
	"fmt"

	"github.com/segmentio/ksuid"
	"github.com/valyala/fasthttp"
)

// GetHandler ...
func GetHandler(ctx *fasthttp.RequestCtx) {
	namespaceID := ctx.UserValue("namespaceId").(string)
	configID := ctx.UserValue("configId").(string)

	key := genKey(namespaceID, configID)

	value, err := Get(GetInput{
		Key: key,
	})

	var cachedConfig CampaignConfig

	fromJSON(value, &cachedConfig)

	if err != nil || cachedConfig.Type == "" {
		if err != nil {
			fmt.Println(fmt.Sprintf("models.config.get: error: %v", err))
		} else {
			fmt.Println(fmt.Sprintf("models.config.get: key not found: %s", key))
		}

		item := CampaignConfig{
			NamespaceID: namespaceID,
			ConfigID:    configID,
			Type:        "not_found",
			Body:        map[string]interface{}{},
		}

		ctx.Write(toJSON(item))
	} else {
		fmt.Println("models.config.get: found value")
		item := CampaignConfig{
			NamespaceID: namespaceID,
			ConfigID:    configID,
			Type:        cachedConfig.Type,
			Body:        cachedConfig.Body,
		}

		ctx.Write(toJSON(item))
	}
}

// PutInput ...
type PutInput struct {
	Type string      `json:"type"`
	Body interface{} `json:"body"`
}

// PutHandler ...
func PutHandler(ctx *fasthttp.RequestCtx) {
	namespaceID := ctx.UserValue("namespaceId").(string)
	configID := ctx.UserValue("configId").(string)
	body := ctx.PostBody()

	var input PutInput

	fromJSON(body, &input)
	key := genKey(namespaceID, configID)

	err := Set(SetInput{
		Key:   key,
		Value: toJSON(input),
	})

	if err != nil {
		fmt.Println(fmt.Sprintf("models.config.put: error: %v", err))
		ctx.Write(toJSON(map[string]interface{}{
			"error": true,
		}))
	} else {
		ctx.Write(toJSON(CampaignConfig{
			NamespaceID: namespaceID,
			ConfigID:    configID,
			Type:        input.Type,
			Body:        input.Body,
		}))
	}
}

// PostInput ...
type PostInput struct {
	Type string      `json:"type"`
	Body interface{} `json:"body"`
}

// PostHandler ...
func PostHandler(ctx *fasthttp.RequestCtx) {
	namespaceID := ctx.UserValue("namespaceId").(string)
	configID := fmt.Sprintf("con_%s", ksuid.New().String())
	body := ctx.PostBody()

	var input PostInput

	json.Unmarshal(body, &input)

	key := genKey(namespaceID, configID)

	err := Set(SetInput{
		Key:   key,
		Value: toJSON(input),
	})

	if err != nil {
		fmt.Println(fmt.Sprintf("models.config.post: error: %v", err))
		ctx.Write(toJSON(map[string]interface{}{
			"error": true,
		}))
	} else {
		ctx.Write(toJSON(CampaignConfig{
			NamespaceID: namespaceID,
			ConfigID:    configID,
			Type:        input.Type,
			Body:        input.Body,
		}))
	}
}

func fromJSON(jsonBytes []byte, v interface{}) {
	json.Unmarshal(jsonBytes, v)
	return
}

func toJSON(i interface{}) []byte {
	json, _ := json.Marshal(i)

	return json
}

func genKey(a string, b string) string {
	return fmt.Sprintf("%s::%s", a, b)
}
