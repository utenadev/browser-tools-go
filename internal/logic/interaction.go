package logic

import (
	"context"
	"fmt"
	"strings"

	"browser-tools-go/internal/models"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

// PickElements extracts information from elements matching a CSS selector.
func PickElements(ctx context.Context, selector string, all bool) ([]models.ElementInfo, error) {
	var nodes []*cdp.Node
	if err := chromedp.Run(ctx, chromedp.Nodes(selector, &nodes, chromedp.NodeVisible, chromedp.ByQuery)); err != nil {
		return nil, fmt.Errorf("could not get nodes for selector '%s': %w", selector, err)
	}
	if len(nodes) == 0 {
		return []models.ElementInfo{}, nil
	}

	if !all {
		nodes = nodes[:1]
	}

	var infos []models.ElementInfo
	for _, node := range nodes {
		var text string
		var attrs map[string]string
		var rect map[string]interface{}

		err := chromedp.Run(ctx,
			chromedp.TextContent(node.NodeID, &text),
			chromedp.Attributes(node.NodeID, &attrs),
			chromedp.ActionFunc(func(ctx context.Context) error {
				result, err := GetBoundingBox(ctx, node.NodeID)
				if err != nil {
					// Don't fail the whole operation, just log a warning and continue
					fmt.Printf("Warning: could not get bounding box for node %d: %v\n", node.NodeID, err)
					rect = make(map[string]interface{})
				} else {
					rect = result
				}
				return nil
			}),
		)

		if err != nil {
			return nil, fmt.Errorf("failed to retrieve details for node %d: %w", node.NodeID, err)
		}

		infos = append(infos, models.ElementInfo{
			Tag:      strings.ToLower(node.NodeName),
			Text:     strings.TrimSpace(text),
			Attrs:    attrs,
			Rect:     rect,
			Children: []models.ElementInfo{},
		})
	}

	return infos, nil
}

// GetBoundingBox gets the bounding box for a given node ID.
func GetBoundingBox(ctx context.Context, nodeID cdp.NodeID) (map[string]interface{}, error) {
	// 1. Resolve node to get its object ID
	remoteObject, err := dom.ResolveNode().WithNodeID(nodeID).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not resolve node: %w", err)
	}
	if remoteObject == nil {
		return nil, fmt.Errorf("resolved node object is nil")
	}

	// 2. Call the getBoundingClientRect function on the resolved object
	var res map[string]interface{}
	err = chromedp.CallFunctionOn(
		"function() { const rect = this.getBoundingClientRect(); return { x: rect.x, y: rect.y, width: rect.width, height: rect.height, top: rect.top, right: rect.right, bottom: rect.bottom, left: rect.left }; }",
		&res,
		func(p *runtime.CallFunctionOnParams) *runtime.CallFunctionOnParams {
			return p.WithObjectID(remoteObject.ObjectID)
		},
	).Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("could not call function on node object: %w", err)
	}
	return res, nil
}

// EvaluateJS executes a JavaScript expression and returns the result.
func EvaluateJS(ctx context.Context, jsExpression string) (interface{}, error) {
	var result interface{}
	err := chromedp.Run(ctx, chromedp.Evaluate(jsExpression, &result))
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate javascript: %w", err)
	}
	return result, nil
}

// GetCookies retrieves all cookies for the current context.
func GetCookies(ctx context.Context) ([]*network.Cookie, error) {
	cookies, err := network.GetCookies().Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cookies: %w", err)
	}
	return cookies, nil
}
