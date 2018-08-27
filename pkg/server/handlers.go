package server

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/segmentio/ksuid"
	"github.com/valyala/fasthttp"

	"github.com/usermirror/config-api/pkg/storage"
)

var store storage.Store = new(storage.Redis)

// GetHandler ...
func GetHandler(ctx *fasthttp.RequestCtx) {
	namespaceID := ctx.UserValue("namespaceId").(string)
	configID := ctx.UserValue("configId").(string)

	key := genKey(namespaceID, configID)

	value, err := store.Get(storage.GetInput{
		Key:     key,
		Timeout: 1000,
	})

	var cachedConfig storage.Config

	fromJSON(value, &cachedConfig)

	if err != nil || value == nil || (cachedConfig.Type == "" && cachedConfig.Body == "") {
		configType := "not_found"

		if err != nil {
			fmt.Println(fmt.Sprintf("handlers.config.get: error: %v", err))
			if strings.Contains(err.Error(), "sealed") {
				configType = "locked"
			}
		} else {
			fmt.Println(fmt.Sprintf("handlers.config.get: key not found: %s", key))
		}

		item := storage.Config{
			NamespaceID: namespaceID,
			ConfigID:    configID,
			Type:        configType,
			Body:        map[string]interface{}{},
		}

		ctx.Write(toJSON(item))
	} else {
		fmt.Println(fmt.Sprintf("handlers.config.get: success: (%s, %s)", namespaceID, configID))
		item := storage.Config{
			NamespaceID: namespaceID,
			ConfigID:    configID,
			Type:        cachedConfig.Type,
			Body:        cachedConfig.Body,
		}

		ctx.Write(toJSON(item))
	}
}

// ScanHandler ...
func ScanHandler(ctx *fasthttp.RequestCtx) {
	namespaceID := ctx.UserValue("namespaceId").(string)

	list, err := store.Scan(storage.ScanInput{
		Prefix:  namespaceID,
		Timeout: 1000,
	})

	if err != nil {
		fmt.Println(fmt.Sprintf("handlers.config.scan: error: %s", err))
		if strings.Contains(err.Error(), "sealed") {
			// configType = "locked"
		}
	} else {
		fmt.Println(fmt.Sprintf("handlers.config.scan: success: (%s)", namespaceID))
	}

	res := map[string]interface{}{
		"namespace_id": namespaceID,
		"type":         "list",
		"items":        []string{},
	}

	if len(list.Keys) != 0 {
		res["items"] = list.Keys
	}

	ctx.Write(toJSON(res))
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

	err := store.Set(storage.SetInput{
		Key:     key,
		Value:   toJSON(input),
		Timeout: 1000,
	})

	if err != nil {
		fmt.Println(fmt.Sprintf("handlers.config.put: error: %v", err))
		ctx.Write(toJSON(map[string]interface{}{
			"error":   true,
			"message": err.Error(),
		}))
	} else {
		fmt.Println(fmt.Sprintf("handlers.config.put: success: (%s, %s)", namespaceID, configID))
		ctx.Write(toJSON(storage.Config{
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

	err := store.Set(storage.SetInput{
		Key:     key,
		Value:   toJSON(input),
		Timeout: 1000,
	})

	if err != nil {
		fmt.Println(fmt.Sprintf("handlers.config.post: error: %v", err))
		ctx.Write(toJSON(map[string]interface{}{
			"error":   true,
			"message": err.Error(),
		}))
	} else {
		fmt.Println(fmt.Sprintf("handlers.config.post: success: (%s, %s)", namespaceID, configID))
		ctx.Write(toJSON(storage.Config{
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
